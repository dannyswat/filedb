package filedb

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type FileIndexConfig struct {
	Unique bool
	Field  string
}

type FileIndex[T FileEntity] interface {
	Init() error
	RebuildIndex(field string) error
	Insert(e T) error
	Update(e, prev T) error
	Delete(prev T) error
	FindId(field string, value string) int
	SearchId(field string, value string) []int
}

type fileIndex[T FileEntity] struct {
	path         string
	indexConfigs []FileIndexConfig
	indexes      map[string]map[string][]int
}

func NewFileIndex[T FileEntity](path string, indexConfig []FileIndexConfig) FileIndex[T] {
	fi := &fileIndex[T]{
		path:         path,
		indexConfigs: indexConfig,
		indexes:      make(map[string]map[string][]int),
	}
	for _, ic := range indexConfig {
		fi.indexes[ic.Field] = make(map[string][]int)
	}
	return fi
}

func (fi *fileIndex[T]) Init() error {
	for _, ic := range fi.indexConfigs {
		if _, err := os.Stat(fi.GetPath(ic.Field)); os.IsNotExist(err) {
			file, err := os.OpenFile(fi.GetPath(ic.Field), os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			err = fi.RebuildIndex(ic.Field)
			if err != nil {
				return err
			}
			for k, v := range fi.indexes[ic.Field] {
				for _, id := range v {
					file.WriteString(fmt.Sprintf("%s\t%d\n", k, id))
				}
			}
		} else {
			if err = fi.LoadIndex(ic.Field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fi *fileIndex[T]) RebuildIndex(field string) error {
	fi.indexes[field] = make(map[string][]int)
	fi.rebuildIndexInternal(field, "", fi.indexes[field])
	return nil
}

func (fi *fileIndex[T]) rebuildIndexInternal(field, path string, index map[string][]int) error {
	entries, err := os.ReadDir(filepath.FromSlash(fi.path + path))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			fi.rebuildIndexInternal(field, path+"/"+entry.Name(), index)
		} else {
			var e T
			if e, err = ReadObject[T](fi.path + path + "/" + entry.Name()); err != nil {
				return err
			}
			index[e.GetValue(field)] = append(index[e.GetValue(field)], e.GetID())
		}
	}

	return nil
}

func (fi *fileIndex[T]) Insert(e T) error {
	for _, ic := range fi.indexConfigs {
		if ic.Unique {
			if idx, ok := fi.indexes[ic.Field][e.GetValue(ic.Field)]; ok && len(idx) > 0 {
				return fmt.Errorf("unique index violation: %s", ic.Field)
			}
		}
	}

	for _, ic := range fi.indexConfigs {
		index, ok := fi.indexes[ic.Field][e.GetValue(ic.Field)]
		if !ok {
			index = make([]int, 0)
		}
		index = append(index, e.GetID())
		fi.indexes[ic.Field][e.GetValue(ic.Field)] = index
		file, err := os.OpenFile(fi.GetPath(ic.Field), os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		file.WriteString(fmt.Sprintf("%s\t%d\n", e.GetValue(ic.Field), e.GetID()))
	}
	return nil
}

func (fi *fileIndex[T]) Update(e, prev T) error {
	for _, ic := range fi.indexConfigs {
		if e.GetValue(ic.Field) == prev.GetValue(ic.Field) {
			continue
		}
		if ic.Unique {
			if idx, ok := fi.indexes[ic.Field][e.GetValue(ic.Field)]; ok && len(idx) > 0 {
				return fmt.Errorf("unique index violation: %s", ic.Field)
			}
		}
	}

	for _, ic := range fi.indexConfigs {
		if e.GetValue(ic.Field) == prev.GetValue(ic.Field) {
			continue
		}
		index := fi.indexes[ic.Field][prev.GetValue(ic.Field)]
		ci := slices.Index(index, e.GetID())
		index = append(index[:ci], index[ci+1:]...)
		fi.indexes[ic.Field][prev.GetValue(ic.Field)] = index
		fi.indexes[ic.Field][e.GetValue(ic.Field)] = append(fi.indexes[ic.Field][e.GetValue(ic.Field)], e.GetID())
		fi.Save(ic.Field)
	}
	return nil
}

func (fi *fileIndex[T]) Delete(prev T) error {
	for _, ic := range fi.indexConfigs {
		index := fi.indexes[ic.Field][prev.GetValue(ic.Field)]
		ci := slices.Index(index, prev.GetID())
		index = append(index[:ci], index[ci+1:]...)
		fi.indexes[ic.Field][prev.GetValue(ic.Field)] = index
		fi.Save(ic.Field)
	}
	return nil
}

func (fi *fileIndex[T]) FindId(field string, value string) int {
	if index, ok := fi.indexes[field][value]; ok {
		if len(index) > 0 {
			return index[0]
		}
	}
	return 0
}

func (fi *fileIndex[T]) SearchId(field string, value string) []int {
	if index, ok := fi.indexes[field][value]; ok {
		return index
	}
	return nil
}

func (fi *fileIndex[T]) Load() error {
	for _, ic := range fi.indexConfigs {
		if err := fi.LoadIndex(ic.Field); err != nil {
			return err
		}
	}
	return nil
}

func (fi *fileIndex[T]) LoadIndex(name string) error {
	file, err := os.OpenFile(fi.GetPath(name), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	fi.indexes[name] = make(map[string][]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var id int
		var value string
		fmt.Sscanf(line, "%s\t%d", &value, &id)
		fi.indexes[name][value] = append(fi.indexes[name][value], id)
	}
	return nil
}

func (fi *fileIndex[T]) Save(name string) error {
	file, err := os.OpenFile(fi.GetPath(name), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	for k, v := range fi.indexes[name] {
		for _, id := range v {
			file.WriteString(fmt.Sprintf("%s\t%d\n", k, id))
		}
	}
	return nil
}

func (fi *fileIndex[T]) GetPath(name string) string {
	return filepath.FromSlash(fi.path + "/_" + name + ".idx")
}
