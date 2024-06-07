package script

import (
	"os"
	"path/filepath"
	"time"

	"github.com/denkhaus/sensor/logging"
	"github.com/denkhaus/sensor/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/sync/errgroup"
)

type ScriptRunner struct {
	i       *interp.Interpreter
	content string
}

type EntrypointFunc func(logger *logrus.Logger, store *store.DataStore) error

func NewScriptRunner(scriptContent string, gopath string) *ScriptRunner {
	i := interp.New(interp.Options{GoPath: gopath})
	if err := i.Use(stdlib.Symbols); err != nil {
		panic(err)
	}
	return &ScriptRunner{i: i, content: scriptContent}
}

// Run executes the given script using the ScriptRunner.
//
// The script is evaluated using the ScriptRunner's interpreter. If there is an error during evaluation,
// it wraps the error with the message "can't evaluate script" and returns it.
//
// The script is expected to define a function named "main.Script" as the entry point. The entry point is
// evaluated using the ScriptRunner's interpreter. If there is an error during evaluation, it wraps the
// error with the message "can't find script entrypoint" and returns it.
//
// The entry point is expected to have a signature of `func(logger *logrus.Logger, store *store.DataStore) error`.
// If the entry point does not have this signature, it wraps the error with the message "wrong signature of entrypoint"
// and returns it.
//
// The entry point function is then called with the logger and store provided by the ScriptRunner. If there is an error
// during the execution of the entry point, it wraps the error with the message "can't execute entrypoint" and returns it.
//
// If all the steps are successful, it returns nil.
func (s *ScriptRunner) Run() error {
	_, err := s.i.Eval(s.content)
	if err != nil {
		return errors.Wrap(err, "evaluate script")
	}

	scriptFunc, err := s.i.Eval(`main.Script()`)
	if err != nil {
		return errors.Wrap(err, "find script entrypoint")
	}

	entryPoint, ok := scriptFunc.Interface().(EntrypointFunc)
	if !ok {
		return errors.Wrap(err, "wrong signature of entrypoint")
	}

	if err := entryPoint(logging.Logger(), store.Store()); err != nil {
		return errors.Wrap(err, "execute entrypoint")
	}

	return nil
}

func Initialize(scriptPath string, runInterval int, eg *errgroup.Group) error {
	absFilePath, err := filepath.Abs(scriptPath)
	if err != nil {
		return errors.Wrap(err, "get absolute path for input script")
	}

	_, err = os.Stat(absFilePath)
	if err != nil {
		return errors.New("can't find input script")
	}

	contentBuf, err := os.ReadFile(scriptPath)
	if err != nil {
		return errors.Wrap(err, "read input script")
	}

	gopath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("can't lookup GOPATH")
	}

	runner := NewScriptRunner(string(contentBuf), gopath)
	durRunInterval := time.Second * time.Duration(runInterval)

	eg.Go(func() error {
		for {
			if err := runner.Run(); err != nil {
				return errors.Wrap(err, "execute script")
			}

			time.Sleep(durRunInterval)
		}
	})

	return nil
}
