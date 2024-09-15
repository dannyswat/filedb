package filedb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type FileDB[T FileEntity] interface {
	Init() error
	Insert(e T) error
	Update(e T) error
	Delete(id int) error
	Find(id int) (T, error)
	List(field, value string) ([]T, error)
	ListIndexFields(field string, value string) ([]*IndexEntry, error)
	ListAllIndexFields(field string) ([]*IndexEntry, error)
	GetCount() int
	PeekNextID() int
	deleteDB() error
}

type fileDB[T FileEntity] struct {
	path    string
	indexes []FileIndexConfig
	stat    FileStat
	index   FileIndex[T]
}

func NewFileDB[T FileEntity](path string, indexes []FileIndexConfig) FileDB[T] {
	return &fileDB[T]{
		path:    path,
		indexes: indexes,
		stat:    NewFileStat(path),
		index:   NewFileIndex[T](path, indexes),
	}
}

func (db *fileDB[T]) Init() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		if err = os.Mkdir(db.path, 0755); err != nil {
			return err
		}
	}
	if err := db.stat.Init(); err != nil {
		return err
	}

	if err := db.index.Init(); err != nil {
		return err
	}
	return nil
}

func (db *fileDB[T]) Insert(e T) error {
	e.SetID(db.stat.GetNextID(false))
	if err := db.index.Insert(e); err != nil {
		return err
	}
	if err := db.stat.AddCount(1); err != nil {
		return err
	}
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(db.GetObjectPath(e.GetID()), bytes, 0644)
}

func (db *fileDB[T]) Update(e T) error {
	prev, err := db.Find(e.GetID())
	if err != nil {
		return err
	}
	if err = db.index.Update(e, prev); err != nil {
		return err
	}
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(db.GetObjectPath(e.GetID()), bytes, 0644)
}

func (db *fileDB[T]) Delete(id int) error {
	prev, err := db.Find(id)
	if err != nil {
		return err
	}
	if err = db.stat.AddCount(-1); err != nil {
		return err
	}
	if err = db.index.Delete(prev); err != nil {
		return err
	}
	return os.Remove(db.GetObjectPath(id))
}

func (db *fileDB[T]) Find(id int) (T, error) {
	return ReadObject[T](db.GetObjectPath(id))
}

func (db *fileDB[T]) List(field, value string) ([]T, error) {
	ids := db.index.SearchId(field, value)
	es := make([]T, 0)
	for _, id := range ids {
		e, err := db.Find(id)
		if err != nil {
			return nil, err
		}
		es = append(es, e)
	}
	return es, nil
}

func (db *fileDB[T]) ListIndexFields(field string, value string) ([]*IndexEntry, error) {
	return db.index.SearchIndex(field, value), nil
}

func (db *fileDB[T]) ListAllIndexFields(field string) ([]*IndexEntry, error) {
	return db.index.SearchAllIndex(field), nil
}

func (db *fileDB[T]) GetCount() int {
	return db.stat.GetCount()
}

func (db *fileDB[T]) PeekNextID() int {
	return db.stat.GetNextID(true)
}

//lint:ignore U1000 Ignore unused function temporarily for testing
func (db *fileDB[T]) deleteDB() error {
	return os.RemoveAll(db.path)
}

func ReadObject[T FileEntity](path string) (T, error) {
	e := reflect.New(reflect.TypeOf(new(T)).Elem().Elem()).Interface().(T)
	bytes, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		return e, err
	}
	if err = json.Unmarshal(bytes, e); err != nil {
		return e, err
	}
	return e, nil
}

func (db *fileDB[T]) GetObjectPath(id int) string {
	nums := make([]string, 0)
	i := id / 1000
	for i > 0 {
		if i%10 > 0 {
			nums = append(nums, strconv.Itoa(i%10))
			CreateDir(filepath.FromSlash(db.path + "/" + strings.Join(nums, "/")))
		}
		i /= 10
	}
	return filepath.FromSlash(db.path + "/" + strings.Join(nums, "/") + strconv.Itoa(id) + ".dat")
}

func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0755)
	}
	return nil
}
