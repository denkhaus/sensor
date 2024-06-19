package store

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	gap "github.com/muesli/go-app-paths"
	"github.com/pkg/errors"
	"github.com/timshannon/badgerhold/v4"
)

var (
	ErrDatabaseNotCreated = errors.New("database instance not created")
)

type embeddedStore struct {
	storageId string
	database  *badgerhold.Store
}

type EmbeddedStore interface {
	Open() (err error)
	Close() (err error)
	Insert(key string, v any) error
	Update(key string, v any) error
	Upsert(key string, v any) error
	Delete(key string, v any) error
	Get(key string, v interface{}) error
	MustGet(key string, v interface{}) bool
}

func NewEmbeddedStore(storageId string) EmbeddedStore {
	return &embeddedStore{
		storageId: storageId,
	}
}

func (p *embeddedStore) Insert(key string, v any) error {
	if p.database == nil {
		return ErrDatabaseNotCreated
	}

	if err := p.database.Insert(key, v); err != nil {
		return errors.Wrap(err, "Insert")
	}

	return nil
}

func (p *embeddedStore) Update(key string, v any) error {
	if p.database == nil {
		return ErrDatabaseNotCreated
	}

	if err := p.database.Update(key, v); err != nil {
		return errors.Wrap(err, "Update")
	}

	return nil
}

func (p *embeddedStore) Upsert(key string, v any) error {
	if p.database == nil {
		return ErrDatabaseNotCreated
	}

	if err := p.database.Upsert(key, v); err != nil {
		return errors.Wrap(err, "Upsert")
	}

	return nil
}

func (p *embeddedStore) Delete(key string, v any) error {
	if p.database == nil {
		return ErrDatabaseNotCreated
	}

	if err := p.database.Delete(key, v); err != nil {
		return errors.Wrap(err, "Delete")
	}

	return nil
}

func (p *embeddedStore) MustGet(key string, v any) bool {
	err := p.Get(key, v)
	if err != nil {
		if errors.Is(err, badgerhold.ErrNotFound) {
			return false
		}

		panic(errors.Wrapf(err, "storage error while retrieving key %s", key))
	}

	return true
}

func (p *embeddedStore) Get(key string, v any) error {
	if p.database == nil {
		return ErrDatabaseNotCreated
	}

	if err := p.database.Get(key, v); err != nil {
		return errors.Wrap(err, "Get")
	}

	return nil
}

func (p *embeddedStore) Open() (err error) {
	if p.database != nil {
		return
	}

	dd, err := p.dataPath(p.dbName())
	if err != nil {
		return errors.Wrap(err, "dataPath")
	}

	options := badgerhold.DefaultOptions
	options.Options = badger.DefaultOptions(dd).
		WithValueLogFileSize(10000000).
		WithLogger(nil)

	p.database, err = badgerhold.Open(options)
	if err != nil {
		return errors.Wrap(err, "Open")
	}

	return
}

func (p *embeddedStore) Close() (err error) {
	if p.database != nil {
		err = p.database.Close()
		p.database = nil
	}

	return
}

func (p *embeddedStore) dbName() string {
	return fmt.Sprintf("%s.db", p.storageId)
}

func (p *embeddedStore) dataPath(id string) (string, error) {
	scope := gap.NewScope(gap.User, filepath.Join("tradesys", id))
	dataPath, err := scope.DataPath("")
	if err != nil {
		return "", errors.Wrap(err, "DataPath")
	}

	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return "", errors.Wrap(err, "MkdirAll")
	}

	return dataPath, nil
}
