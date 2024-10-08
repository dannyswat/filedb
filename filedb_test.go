package filedb

import (
	"strconv"
	"testing"
)

type TestEntity struct {
	ID   int
	Name string
	Age  int
}

func NewTestEntity(name string, age int) *TestEntity {
	return &TestEntity{
		Name: name,
		Age:  age,
	}
}

func (te *TestEntity) GetID() int {
	return te.ID
}

func (te *TestEntity) SetID(id int) {
	te.ID = id
}

func (te *TestEntity) GetValue(field string) string {
	switch field {
	case "Name":
		return te.Name
	case "Age":
		return strconv.Itoa(te.Age)
	}
	return ""
}

func TestFileDB(t *testing.T) {
	db := NewFileDB[*TestEntity]("test", []FileIndexConfig{
		{Unique: true, Field: "Name"},
		{Unique: false, Field: "Age"},
	})
	if err := db.Init(); err != nil {
		t.Error(err)
	}
	lastID := db.PeekNextID() - 1
	initCount := db.GetCount()

	if err := db.Insert(NewTestEntity("Alice", 20)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	if err := db.Insert(NewTestEntity("Bob", 30)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	if err := db.Insert(NewTestEntity("Peter", 20)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	e, err := db.Find(lastID + 2)
	if err != nil {
		t.Error(err)
	}
	if e.Name != "Bob" {
		t.Error("Find failed")
	}
	if db.GetCount() != 3 {
		t.Error("GetCount failed")
	}
	if err := db.Delete(lastID + 1); err != nil {
		t.Error(err)
	}
	if err := db.Delete(lastID + 2); err != nil {
		t.Error(err)
	}
	if err := db.Delete(lastID + 3); err != nil {
		t.Error(err)
	}
	if db.GetCount() != 0 {
		t.Error("GetCount after deleted failed")
	}

	if err := db.deleteDB(); err != nil {
		t.Error(err)
	}
}

func TestFileIndexInclude(t *testing.T) {
	db := NewFileDB[*TestEntity]("test", []FileIndexConfig{
		{Unique: true, Field: "Name"},
		{Unique: false, Field: "Age", Include: []string{"Name"}},
	})
	if err := db.Init(); err != nil {
		t.Error(err)
	}
	lastID := db.PeekNextID() - 1
	initCount := db.GetCount()

	if err := db.Insert(NewTestEntity("Alice", 20)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	if err := db.Insert(NewTestEntity("Bob", 30)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	if err := db.Insert(NewTestEntity("Peter", 20)); err != nil {
		if initCount > 0 {
			if err.Error() != "unique index violation: Name" {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	records, err := db.ListIndexFields("Age", "20")
	if err != nil {
		t.Error(err)
	}
	if len(records) != 2 {
		t.Error("ListIndexFields failed")
	}
	if records[0].Others["Name"] != "Alice" {
		t.Error("ListIndexFields failed")
	}
	if records[1].Others["Name"] != "Peter" {
		t.Error("ListIndexFields failed")
	}
	if err := db.Update(&TestEntity{ID: 1, Name: "Alice S", Age: 20}); err != nil {
		t.Error(err)
	}
	records, err = db.ListIndexFields("Age", "20")
	if err != nil {
		t.Error(err)
	}
	if len(records) != 2 {
		t.Error("ListIndexFields failed after update")
	}
	if records[0].Others["Name"] != "Alice S" {
		t.Error("ListIndexFields failed after update")
	}
	e, err := db.Find(lastID + 2)
	if err != nil {
		t.Error(err)
	}
	if e.Name != "Bob" {
		t.Error("Find failed")
	}
	if db.GetCount() != 3 {
		t.Error("GetCount failed")
	}
	if err := db.Delete(lastID + 1); err != nil {
		t.Error(err)
	}
	if err := db.Delete(lastID + 2); err != nil {
		t.Error(err)
	}
	if err := db.Delete(lastID + 3); err != nil {
		t.Error(err)
	}
	if db.GetCount() != 0 {
		t.Error("GetCount after deleted failed")
	}

	if err := db.deleteDB(); err != nil {
		t.Error(err)
	}
}
