package boltdb

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/inpime/sdata"
	"github.com/stretchr/testify/assert"
)

var dbFileFlag = flag.String("dbfile", "_testdbfile.db", "db file name")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestBoltdbStore_simplestrategy(t *testing.T) {
	db, err := bolt.Open(*dbFileFlag, 0644, &bolt.Options{})
	assert.NoError(t, err, "open db")
	defer func() {
		os.Remove(*dbFileFlag)
	}()

	list := New("default", db)
	list.Add(context.Background(), "email6")
	list.Add(context.Background(), "email1")
	list.Add(context.Background(), "email5")
	list.Add(context.Background(), "email2", "email4", "email3")

	list.Del("email2")
	list.Del("email5")

	_, err = list.Get("email1")
	assert.NoError(t, err)

	_, err = list.Get("email2")
	assert.Error(t, err, "not exists")

	_, err = list.Get("email3")
	assert.NoError(t, err)

	_, err = list.Get("email4")
	assert.NoError(t, err)

	_, err = list.Get("email5")
	assert.Error(t, err, "not exists")

	_, err = list.Get("email6")
	assert.NoError(t, err)

	// Search full

	arr, err := list.Search(context.Background(), "", 10)
	assert.NoError(t, err)

	expected := sdata.NewArray()
	expected.Add("email1").
		Add("email3").
		Add("email4").
		Add("email6")

	assert.Equal(t, arr.Size(), expected.Size())

	for _, value := range arr.Data() {
		_value := value.(*sdata.Map)

		assert.Equal(t, expected.Exist(_value.Get("email")), true)
	}

	// Search perpage

	arr, err = list.Search(context.Background(), "email4", 10)
	assert.NoError(t, err)

	expected = sdata.NewArray()
	expected.
		Add("email4").
		Add("email6")

	assert.Equal(t, arr.Size(), expected.Size())

	for _, value := range arr.Data() {
		_value := value.(*sdata.Map)

		assert.Equal(t, expected.Exist(_value.Get("email")), true)
	}
}
