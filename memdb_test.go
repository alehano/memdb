package memdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const indexName = "name"

type myData struct {
	Name   string
	Number int
}

func (md myData) GetIndex(name string) interface{} {
	if name == indexName {
		return md.Name
	}
	return nil
}

func TestDB(t *testing.T) {
	db := NewMemDB("db.json")
	db.AddIndex(indexName)

	t.Log("==========")
	t.Logf("%+v", db.secondaryIdx)

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

	items, err := db.GetAll(0, 0)
	assert.Nil(t, err)
	assert.Equal(t, item1, items[0])
	assert.Equal(t, item2, items[1])
	assert.Equal(t, item3, items[2])

	db.Iterate(func(id int64, item Item) (stop bool, err error) {
		myItem := item.(myData)
		t.Logf("%d - Name: %s, Number: %d\n", id, myItem.Name, myItem.Number)
		if id == 2 {
			return true, nil
		}
		return false, nil
	})

	// delete
	db.Delete(1)

	_, err = db.Get(1)
	assert.Equal(t, ErrIDNotExists, err)

	t.Log("--------")

	db.Iterate(func(id int64, item Item) (stop bool, err error) {
		myItem := item.(myData)
		t.Logf("%d - Name: %s, Number: %d\n", id, myItem.Name, myItem.Number)
		return false, nil
	})

	items, err = db.GetByIndex(indexName, "Alex")
	assert.Nil(t, err)
	assert.Len(t, items, 1)
	if len(items) == 1 {
		myItem := items[0].(myData)
		assert.Equal(t, 2, myItem.Number)
	}

}
