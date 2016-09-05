package boltdb

import (
	"context"
	"errors"
	"log"
	"time"

	"pochtalion"
	"pochtalion/utils"

	"github.com/boltdb/bolt"
	"github.com/inpime/sdata"
)

var bucketList = []byte("_buckets")
var _ pochtalion.Store = (*Store)(nil)

func New(name string, db *bolt.DB) *Store {
	name = utils.Format(name)

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})

	return &Store{
		db:   db,
		name: []byte(name),
	}
}

type Store struct {
	db   *bolt.DB
	name []byte
}

// NewBucket создать новый пустой список
func (s *Store) New(name string) pochtalion.Store {
	return New(name, s.db)
}

func (s *Store) Del(email string) error {
	email = utils.Format(email)

	return s.del(email)
}

func (s *Store) Get(email string) (*sdata.Map, error) {
	email = utils.Format(email)

	return s.get(email)
}

func (s *Store) Set(email string, data *sdata.Map) error {
	email = utils.Format(email)

	_data, err := s.get(email)

	if err != nil {
		return err
	}

	if err := sdata.Merge(_data, data); err != nil {
		return err
	}

	return s.set(email, _data)
}

func (s *Store) Add(ctx context.Context, emails ...string) error {
	var done = make(chan error, 1)

	go func() {
		for _, email := range emails {
			email = utils.Format(email)

			data, err := s.get(email)

			if data == nil {
				err = s.set(email, sdata.
					NewMap().
					Set("email", email).
					Set("createdat", time.Now().UnixNano()))

				if err != nil {
					done <- err
				}
			}
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}

}

func (s *Store) Search(ctx context.Context, prevkey string, size int) (*sdata.Array, error) {
	res := sdata.NewArray()
	var done = make(chan error, 1)

	go func() {
		err := s.db.View(func(tx *bolt.Tx) error {
			var _prevKey = []byte(utils.Format(prevkey))
			c := tx.Bucket(s.name).Cursor()

			for k, v := c.Seek(_prevKey); ; k, v = c.Next() {
				// If we hit the end of our sessions then exit.
				if k == nil {
					_prevKey = nil
					return nil
				}

				if size <= 0 {
					_prevKey = nil
					return nil
				}

				data := sdata.NewMap()
				err := data.UnmarshalMsgpack(v)
				if err != nil {
					log.Println("Error unmarshal for email", string(k), err)
				} else {
					res.Add(data)
				}

				_prevKey = make([]byte, len(k))
				copy(_prevKey, k)
				size--
			}
		})

		done <- err
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-done:
		if err == nil {
			return res, nil
		}

		return nil, err
	}
}

func (s *Store) get(key string) (*sdata.Map, error) {
	var (
		data   *sdata.Map
		raw    []byte
		exists bool
	)

	err := s.db.View(func(tx *bolt.Tx) error {
		raw = tx.
			Bucket(s.name).
			Get([]byte(key))
		exists = len(raw) != 0 // exists
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("not exists")
	}

	data = sdata.NewMap()
	if err := data.UnmarshalMsgpack(raw); err != nil {
		return nil, err
	}

	return data, err
}

func (s *Store) del(email string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.
			Bucket(s.name).
			Delete([]byte(email))
	})
}

func (s *Store) set(email string, data *sdata.Map) (err error) {
	data.Set("updatedat", time.Now().UnixNano())

	raw, err := data.MarshalMsgpack()

	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.
			Bucket(s.name).
			Put([]byte(email), raw)
	})

	return err
}
