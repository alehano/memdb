package memdb

import (
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

const indexName = "name"

type myData struct {
	Name   string
	Number int
}

func (md myData) GetIndex(name string) interface{} {
	switch name {
	case indexName:
		return md.Name
	default:
		return nil
	}
}

func TestDB(t *testing.T) {
	noDb := NewMemDB("")
	err := noDb.SaveDB(nil)
	assert.Equal(t, ErrFilenameWasntSet, err)

	var dbFile = os.Getenv("GOPATH") + "/src/github.com/alehano/memdb/db.gob"
	db := NewMemDB(dbFile)
	db.AddIndex(indexName)

	item1 := myData{Name: "John", Number: 3}
	item2 := myData{Name: "Alex", Number: 2}
	item3 := myData{Name: "Tim", Number: 1}

	id, err := db.Create(item1)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, id)

	id, err = db.Create(item2)
	assert.NotEqual(t, 0, id)
	assert.Nil(t, err)
	id, err = db.Create(item3)
	assert.NotEqual(t, 0, id)
	assert.Nil(t, err)

	err = db.Update(1, myData{Name: "John Updated", Number: 3})
	assert.Nil(t, err)
	item0, err := db.Get(1)
	myItem := item0.(myData)
	assert.Equal(t, "John Updated", myItem.Name)

	// save / load
	err = db.SaveDB(myData{})
	assert.Nil(t, err)

	db2 := NewMemDB(dbFile)
	db2.AddIndex(indexName)
	err = db2.LoadDB(myData{})
	assert.Nil(t, err)

	item2Db2, err := db2.Get(2)
	assert.Nil(t, err)
	assert.Equal(t, item2, item2Db2)

	items, err := db.GetAll(1, 1)
	assert.Nil(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, item2, items[0])

	db.Iterate(func(id int64, item Item) (stop bool, err error) {
		myItem := item.(myData)
		t.Logf("%d - Name: %s, Number: %d\n", id, myItem.Name, myItem.Number)
		if id == 2 {
			return true, nil
		}
		return false, nil
	})

	// delete
	err = db.Delete(1)
	assert.Nil(t, err)
	assert.Len(t, db.items, 3)
	db.CleanUp()
	assert.Len(t, db.items, 2)

	_, err = db.Get(1)
	assert.Equal(t, ErrIDNotExists, err)

	db.Iterate(func(id int64, item Item) (stop bool, err error) {
		myItem := item.(myData)
		t.Logf("%d - Name: %s, Number: %d\n", id, myItem.Name, myItem.Number)
		return false, nil
	})

	items, err = db.GetAllByIndex(indexName, "Alex")
	assert.Nil(t, err)
	assert.Len(t, items, 1)
	if len(items) == 1 {
		myItem := items[0].(myData)
		assert.Equal(t, 2, myItem.Number)
	}
}
