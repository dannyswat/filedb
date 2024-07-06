package filedb

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileStat interface {
	Init() error
	GetNextID(peek bool) int
	GetCount() int
	AddCount(c int) error
}

type fileStat struct {
	path     string
	statPath string
	nextID   int
	count    int
}

func NewFileStat(path string) FileStat {
	return &fileStat{
		path:     path,
		statPath: filepath.FromSlash(path + "/_stat.dat"),
		nextID:   1,
		count:    0,
	}
}

func (fs *fileStat) Init() error {
	if _, err := os.Stat(fs.statPath); os.IsNotExist(err) {
		file, err := os.OpenFile(fs.statPath, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		file.WriteString(fmt.Sprintf("%d\n%d\n", fs.nextID, fs.count))
		file.Close()
	} else {
		if err = fs.Load(); err != nil {
			return err
		}
	}
	return nil
}

func (fs *fileStat) GetNextID(peek bool) int {
	if peek {
		return fs.nextID
	}
	id := fs.nextID
	fs.nextID++
	fs.Save()
	return id
}

func (fs *fileStat) GetCount() int {
	return fs.count
}

func (fs *fileStat) AddCount(c int) error {
	fs.count += c
	fs.Save()
	return nil
}

func (fs *fileStat) Load() error {
	file, err := os.Open(fs.statPath)
	if err != nil {
		return err
	}
	fmt.Fscanf(file, "%d\n%d\n", &fs.nextID, &fs.count)
	file.Close()
	return nil
}

func (fs *fileStat) Save() error {
	file, err := os.OpenFile(fs.statPath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	file.WriteString(fmt.Sprintf("%d\n%d\n", fs.nextID, fs.count))
	file.Close()
	return nil
}
