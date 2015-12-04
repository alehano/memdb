# SimpleMemDB - simple in-memory database written in Go

Stores any structure in memory. Provide some methods to access data and save/load on a disc.

Example:
```go

    const indexSlug = "slug"
    
    type myItem struct {
        Name  string
        Slug  string
        Value int
    }
    
    func (md myItem) GetIndex(name string) interface{} {
        switch name {
        case indexSlug:
            return md.Slug
        default:
            return nil
        }
    }
    
    
    db := NewMemDB("/path/to/db/dbname.gob")
    // Set index
    db.AddIndex(indexSlug)
  
    id1, err := db.Create(myItem{
        Name:  "Sirius",
        Slug:  "sir",
        Value: 3,
    })
    if err != nil {
        fmt.Println(err)
    }
    fmt.Printf("New ID: %d\n", id1)
  
    id2, err := db.Create(myItem{
        Name:  "Orion",
        Slug:  "or",
        Value: 5,
    })
    if err != nil {
        fmt.Println(err)
    }
    fmt.Printf("New ID: %d\n", id2)
  
    // Get
    val, err := db.Get(id2)
    if err != nil {
        fmt.Println(err)
    }
    secondItem := val.(myItem)
    fmt.Printf("Second item: %s\n", secondItem.Name)
  
    // Update
    secondItem.Value = 10
    err = db.Update(id2, secondItem)
    if err != nil {
        fmt.Println(err)
    }
  
    // Get by index
    vals, _ := db.GetAllByIndex(indexSlug, "sir")
    if len(vals) == 1 {
        item := vals[0].(myItem)
        fmt.Printf("Item with slug 'sir': %s\n", item.Name)
    }
  
    // Iterate
    err = db.Iterate(func(id int64, item Item) (stop bool, err error) {
        myItem := item.(myItem)
        fmt.Printf("%d - Name: %s, Value: %d\n", id, myItem.Name, myItem.Value)
        return false, nil
    })
    if err != nil {
        fmt.Println(err)
    }
  
    // Get all with limit and skip
    allItems, err := db.GetAll(1, 1)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(allItems)
  
    // Delete
    err = db.Delete(id1)
    if err != nil {
        fmt.Println(err)
    }
    // Clean deleted items
    db.CleanUp()
  
    // Save / Load
    err = db.SaveDB(myItem{})
    if err != nil {
        fmt.Println(err)
    }
  
    err = db.LoadDB(myItem{})
    if err != nil {
        fmt.Println(err)
    }
    
   
```