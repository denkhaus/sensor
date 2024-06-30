package script

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/denkhaus/sensor/config"
	"github.com/denkhaus/sensor/logging"
	"github.com/denkhaus/sensor/store"
	"github.com/denkhaus/sensor/symbols"
	"github.com/denkhaus/sensor/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/sync/errgroup"
)

type EntrypointFunc func(ctx *types.ScriptContext) error

type ScriptRunner struct {
	i             *interp.Interpreter
	scriptFunc    reflect.Value
	setupFunc     reflect.Value
	scriptContext *types.ScriptContext
	content       string
}

func NewScriptRunner(scriptContent string, gopath string) (*ScriptRunner, error) {
	i := interp.New(interp.Options{GoPath: gopath})

	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, errors.Wrap(err, "load standard library")
	}
	if err := i.Use(symbols.Symbols); err != nil {
		return nil, errors.Wrap(err, "load buildin library")
	}

	_, err := i.Eval(scriptContent)
	if err != nil {
		return nil, errors.Wrap(err, "evaluate script")
	}

	scriptFunc, err := i.Eval(`main.Script`)
	if err != nil {
		return nil, errors.Wrap(err, "find script entrypoint")
	}

	setupFunc, err := i.Eval(`main.Setup`)
	if err != nil {
		return nil, errors.Wrap(err, "find setup entrypoint")
	}

	scriptContext := &types.ScriptContext{
		Logger:        logging.Logger(),
		SensorStore:   store.Sensor(),
		EmbeddedStore: store.Embedded(),
	}

	return &ScriptRunner{
		i:             i,
		content:       scriptContent,
		scriptFunc:    scriptFunc,
		setupFunc:     setupFunc,
		scriptContext: scriptContext,
	}, nil
}

func (s *ScriptRunner) Run() error {
	in := []reflect.Value{
		reflect.ValueOf(s.scriptContext),
	}

	out := s.scriptFunc.Call(in)
	if e := out[0].Interface(); e != nil {
		err := e.(error)
		return errors.Wrap(err, "execute script entrypoint")
	}

	return nil
}

func (s *ScriptRunner) Setup() error {
	in := []reflect.Value{
		reflect.ValueOf(s.scriptContext),
	}

	out := s.setupFunc.Call(in)
	if e := out[0].Interface(); e != nil {
		err := e.(error)
		return errors.Wrap(err, "execute setup entrypoint")
	}

	return nil
}

func Initialize(ctx context.Context, logger *logrus.Logger, config *config.Config, eg *errgroup.Group) error {
	absFilePath, err := filepath.Abs(config.Script.Path)
	if err != nil {
		return errors.Wrap(err, "get absolute path for input script")
	}

	_, err = os.Stat(absFilePath)
	if err != nil {
		return errors.New("no input script found")
	}

	contentBuf, err := os.ReadFile(absFilePath)
	if err != nil {
		return errors.Wrap(err, "read input script")
	}

	gopath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("can't lookup GOPATH")
	}

	runner, err := NewScriptRunner(string(contentBuf), gopath)
	if err != nil {
		return errors.Wrap(err, "create script runner")
	}

	durRunInterval := time.Second * time.Duration(config.Script.RunInterval)

	eg.Go(func() error {
		for {
			if err := runner.Run(); err != nil {
				return errors.Wrap(err, "execute script")
			}

			select {
			case <-ctx.Done():
				logger.Info("script-runner: done received -> closing")
				return nil
			default:
				time.Sleep(durRunInterval)
			}
		}
	})

	return nil
}
