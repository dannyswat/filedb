package filedb

import (
	"os"
)

type FileDB[T FileEntity] interface {
	HasIndex(field string, unique bool) FileDB[T]
	Init() error
	Insert(e T) error
	Update(e T) error
	Delete(id int) error
	Find(id int) (T, error)
	List(field, value string) ([]T, error)
}

type fileDB[T FileEntity] struct {
	path    string
	indexes []FileIndexConfig
	nextID  int
	count   int
}

func NewFileDB[T FileEntity](path string) *fileDB[T] {
	return &fileDB[T]{path: path, indexes: make([]FileIndexConfig, 0), nextID: 1, count: 0}
}

func (db *fileDB[T]) HasIndex(field string, unique bool) *fileDB[T] {
	db.indexes = append(db.indexes, FileIndexConfig{Field: field, Unique: unique})
	return db
}

func (db *fileDB[T]) Init() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		if err = os.Mkdir(db.path, 0755); err != nil {
			return err
		}
	}
	return nil
}
