package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("reputation-source-store")

	// ErrAlreadyExists returns when the soure already exists in Store.
	ErrAlreadyExists = errors.New("source already exists")
	// ErrDoesntExists returns when the source isn't in the Store.
	ErrDoesntExists = errors.New("source doesn't exist")

	baseKey = datastore.NewKey("/reputation/store")
)

// Store contains Sources information.
type Store struct {
	ds datastore.TxnDatastore
}

// NewStore returns a new SourceStore.
func NewStore(ds datastore.TxnDatastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Add adds a new Source to the store.
func (ss *Store) Add(s Source) error {
	txn, err := ss.ds.NewTransaction(context.Background(), false)
	if err != nil {
		return err
	}
	defer txn.Discard(context.Background())

	k := genKey(s.ID)
	ok, err := txn.Has(context.Background(), k)
	if err != nil {
		return err
	}
	if ok {
		return ErrAlreadyExists
	}
	return ss.put(txn, s)
}

// Update updates a Source.
func (ss *Store) Update(s Source) error {
	txn, err := ss.ds.NewTransaction(context.Background(), false)
	if err != nil {
		return err
	}
	k := genKey(s.ID)
	ok, err := txn.Has(context.Background(), k)
	if err != nil {
		return err
	}
	if !ok {
		return ErrDoesntExists
	}
	return ss.put(txn, s)
}

// GetAll returns all Sources.
func (ss *Store) GetAll() ([]Source, error) {
	txn, err := ss.ds.NewTransaction(context.Background(), true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(context.Background())
	q := query.Query{Prefix: baseKey.String()}
	res, err := txn.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("error when closing query result: %s", err)
		}
	}()
	var ret []Source
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		s := Source{}
		if err := json.Unmarshal(r.Value, &s); err != nil {
			return nil, err
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (ss *Store) put(txn datastore.Txn, s Source) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := txn.Put(context.Background(), genKey(s.ID), b); err != nil {
		return err
	}
	return txn.Commit(context.Background())
}

func genKey(id string) datastore.Key {
	return baseKey.ChildString(id)
}
