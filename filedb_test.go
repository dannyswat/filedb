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
	if err := db.Insert(NewTestEntity("Alice", 20)); err != nil {
		t.Error(err)
	}
	if err := db.Insert(NewTestEntity("Bob", 30)); err != nil {
		t.Error(err)
	}
	if err := db.Insert(NewTestEntity("Peter", 20)); err != nil {
		t.Error(err)
	}
}
