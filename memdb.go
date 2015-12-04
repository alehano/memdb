package memdb

import (
	"encoding/gob"
	"os"
	"sync"
)

type Item interface {
	GetIndex(name string) interface{}
}

func NewMemDB(filename string) *memDB {
	mdb := &memDB{
		mu:           &sync.Mutex{},
		filename:     filename,
		idx:          make(map[int64]int64),
		secondaryIdx: make(map[string]sIndexes),
	}
	return mdb
}

type sIndexes map[interface{}][]int64

type dbItem struct {
	Item
	ID  int64
	Del bool
}

func (dbi dbItem) getItem() (Item, error) {
	if dbi.Del {
		return nil, ErrIDNotExists
	}
	return dbi.Item, nil
}

func (dbi dbItem) getID() int64 {
	return dbi.ID
}

func (dbi dbItem) setID() int64 {
	return dbi.ID
}

func (dbi *dbItem) setDeleted() {
	dbi.Del = true
}

type memDB struct {
	mu           *sync.Mutex
	filename     string
	items        []dbItem
	idx          map[int64]int64 // [id]position
	secondaryIdx map[string]sIndexes
	maxId        int64
	lenItems     int64
}

func (mdb *memDB) Create(item Item) (int64, error) {
	id := mdb.maxId + 1
	mdb.mu.Lock()
	mdb.items = append(mdb.items, dbItem{Item: item, ID: id})
	mdb.lenItems++
	mdb.maxId = id
	mdb.idx[id] = mdb.lenItems - 1

	// secondary indexes
	for name := range mdb.secondaryIdx {
		sIdxVal := item.GetIndex(name)
		mdb.secondaryIdx[name][sIdxVal] =
			append(mdb.secondaryIdx[name][sIdxVal], id)
	}

	mdb.mu.Unlock()
	return id, nil
}

func (mdb memDB) Get(id int64) (Item, error) {
	position, err := mdb.getPositionByID(id)
	if err != nil {
		return nil, err
	}
	return mdb.items[position].getItem()
}

func (mdb memDB) GetAll(limit, skip int64) ([]Item, error) {
	var count int64
	var skipped int64
	var items []Item
	for pos := range mdb.items {
		if !mdb.items[pos].Del {
			if skipped >= skip {
				items = append(items, mdb.items[pos].Item)
				count++
				if limit > 0 && count >= limit {
					break
				}
			} else {
				skipped++
			}
		}
	}
	return items, nil
}

func (mdb memDB) GetAllByIndex(indexName string, value interface{}) ([]Item, error) {
	if ids, ok := mdb.secondaryIdx[indexName][value]; ok {
		var items []Item
		for _, id := range ids {
			item, err := mdb.Get(id)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	}
	return nil, ErrIndexNotFound
}

func (mdb memDB) Iterate(fn func(id int64, item Item) (stop bool, err error)) error {
	for _, dbItem := range mdb.items {
		if !dbItem.Del {
			stop, err := fn(dbItem.ID, dbItem.Item)
			if err != nil {
				return err
			}
			if stop {
				break
			}
		}
	}
	return nil
}

func (mdb *memDB) Update(id int64, item Item) error {
	if id == 0 {
		return ErrNoIDProvided
	}
	position, err := mdb.getPositionByID(id)
	if err != nil {
		return err
	}
	mdb.mu.Lock()
	mdb.items[position].Item = item
	mdb.mu.Unlock()
	return nil
}

func (mdb *memDB) Delete(id int64) error {
	position, err := mdb.getPositionByID(id)
	if err != nil {
		return err
	}
	mdb.mu.Lock()
	mdb.items[position].Del = true
	mdb.mu.Unlock()
	return nil
}

// You have to pass your implementation of Item
// E.g. SaveDB(myItem{})
func (mdb *memDB) SaveDB(dataType interface{}) error {
	if mdb.filename == "" {
		return ErrFilenameWasntSet
	}

	f, err := os.Create(mdb.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	gob.Register(dataType)
	enc := gob.NewEncoder(f)

	mdb.CleanUp()

	mdb.mu.Lock()
	err = enc.Encode(mdb.items)
	mdb.mu.Unlock()
	return err
}

// You have to pass your implementation of Item
// E.g. SaveDB(myItem{})
func (mdb *memDB) LoadDB(dataType interface{}) error {
	if mdb.filename == "" {
		return ErrFilenameWasntSet
	}

	f, err := os.Open(mdb.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	gob.Register(dataType)
	dec := gob.NewDecoder(f)
	var dbItems []dbItem
	err = dec.Decode(&dbItems)
	mdb.mu.Lock()
	mdb.items = dbItems
	mdb.mu.Unlock()
	mdb.ReindexAll()

	return err
}

// AddIndex adds secondary index by name
// You should provide GetIndex(name) method on your Item interface implementation
func (mdb *memDB) AddIndex(name string) error {
	mdb.mu.Lock()
	mdb.secondaryIdx[name] = sIndexes{}
	mdb.mu.Unlock()
	mdb.reindexSecondary()
	return nil
}

// Clean deleted items
// Iterate all DB and create a new
func (mdb *memDB) CleanUp() {
	newItems := []dbItem{}
	for _, dbItem := range mdb.items {
		if !dbItem.Del {
			newItems = append(newItems, dbItem)
		}
	}
	mdb.mu.Lock()
	mdb.items = newItems
	mdb.mu.Unlock()
	mdb.ReindexAll()
}

func (mdb *memDB) ReindexAll() {
	mdb.reindex()
	mdb.reindexSecondary()
}

func (mdb *memDB) reindex() {
	idx := map[int64]int64{}
	var id int64 = 0
	var maxId int64 = 0
	var lenItems int64 = 0
	for _, dbItem := range mdb.items {
		if !dbItem.Del {
			id = dbItem.ID
			lenItems++
			idx[id] = lenItems - 1
			if id > maxId {
				maxId = id
			}
		}
	}

	mdb.mu.Lock()
	defer mdb.mu.Unlock()
	mdb.idx = idx
	mdb.maxId = maxId
	mdb.lenItems = lenItems
}

func (mdb *memDB) reindexSecondary() {
	var val interface{}
	mdb.mu.Lock()
	defer mdb.mu.Unlock()
	for name := range mdb.secondaryIdx {
		// clean ids
		mdb.secondaryIdx[name] = sIndexes{}
		for _, dbItem := range mdb.items {
			if !dbItem.Del {
				val = dbItem.GetIndex(name)
				if val != nil {
					mdb.secondaryIdx[name][val] = append(mdb.secondaryIdx[name][val],
						dbItem.ID)
				}
			}
		}
	}
}

func (mdb memDB) getPositionByID(id int64) (int64, error) {
	if position, ok := mdb.idx[id]; ok {
		return position, nil

	}
	return 0, ErrIDNotExists
}
