package filedb

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type FileIndexConfig struct {
	Unique  bool
	Field   string
	Include []string
}

type IndexEntry struct {
	Value  string
	ID     int
	Others map[string]string
}

type FileIndex[T FileEntity] interface {
	Init() error
	RebuildIndex(config *FileIndexConfig) error
	Insert(e T) error
	Update(e, prev T) error
	Delete(prev T) error
	FindId(field string, value string) int
	SearchId(field string, value string) []int
	SearchIndex(field string, value string) []*IndexEntry
	SearchAllIndex(field string) []*IndexEntry
	FindMaxIdAndCount() (int, int)
	ListAllIds() []int
}

type fileIndex[T FileEntity] struct {
	path         string
	indexConfigs []FileIndexConfig
	indexes      map[string]map[string][]*IndexEntry
}

func NewFileIndex[T FileEntity](path string, indexConfig []FileIndexConfig) FileIndex[T] {
	fi := &fileIndex[T]{
		path:         path,
		indexConfigs: indexConfig,
		indexes:      make(map[string]map[string][]*IndexEntry),
	}
	for _, ic := range indexConfig {
		fi.indexes[ic.Field] = make(map[string][]*IndexEntry)
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
			err = fi.RebuildIndex(&ic)
			if err != nil {
				return err
			}
			for k, v := range fi.indexes[ic.Field] {
				for _, entry := range v {
					file.WriteString(fmt.Sprintf("%s\t%d", k, entry.ID))
					for _, i := range ic.Include {
						file.WriteString(fmt.Sprintf("\t%s", entry.Others[i]))
					}
					file.WriteString("\n")
				}
			}
		} else {
			if err = fi.LoadIndex(ic.Field); err != nil {
				_, ok := err.(*InvalidIndexError)
				if !ok {
					return err
				}
				if err = fi.RebuildIndex(&ic); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (fi *fileIndex[T]) RebuildIndex(config *FileIndexConfig) error {
	fi.indexes[config.Field] = make(map[string][]*IndexEntry)
	fi.rebuildIndexInternal(config.Field, "", config.Include, fi.indexes[config.Field])
	return nil
}

func getFields(e FileEntity, includes []string) map[string]string {
	fields := make(map[string]string)
	for _, f := range includes {
		fields[f] = e.GetValue(f)
	}
	return fields
}

func createIndexEntry(e FileEntity, field string, includes []string) *IndexEntry {
	return &IndexEntry{
		Value:  e.GetValue(field),
		ID:     e.GetID(),
		Others: getFields(e, includes),
	}
}

func (fi *fileIndex[T]) rebuildIndexInternal(field, path string, includes []string, index map[string][]*IndexEntry) error {
	entries, err := os.ReadDir(filepath.FromSlash(fi.path + path))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			fi.rebuildIndexInternal(field, path+"/"+entry.Name(), includes, index)
		} else {
			var e T
			if e, err = ReadObject[T](fi.path + path + "/" + entry.Name()); err != nil {
				return err
			}
			index[e.GetValue(field)] = append(
				index[e.GetValue(field)],
				createIndexEntry(e, field, includes))
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
			index = make([]*IndexEntry, 0)
		}
		index = append(index, createIndexEntry(e, ic.Field, ic.Include))
		fi.indexes[ic.Field][e.GetValue(ic.Field)] = index
		file, err := os.OpenFile(fi.GetPath(ic.Field), os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		file.WriteString(fmt.Sprintf("%s\t%d", e.GetValue(ic.Field), e.GetID()))
		for _, i := range ic.Include {
			file.WriteString(fmt.Sprintf("\t%s", e.GetValue(i)))
		}
		file.WriteString("\n")
	}
	return nil
}

func (fi *fileIndex[T]) Update(e, prev T) error {

	// Validation
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
	idComparer := func(item *IndexEntry) bool {
		return prev.GetID() == item.ID
	}

	// Update index
	for _, ic := range fi.indexConfigs {
		indexFile := fi.indexes[ic.Field]
		newKeyValue := e.GetValue(ic.Field)
		oldKeyValue := prev.GetValue(ic.Field)
		if newKeyValue == oldKeyValue {
			// No change in index field, check if include fields have changed
			indexEntries := indexFile[newKeyValue]
			entryIndex := slices.IndexFunc(indexEntries, idComparer)
			changed := false
			for _, field := range ic.Include {
				if e.GetValue(field) != prev.GetValue(field) {
					indexEntries[entryIndex].Others[field] = e.GetValue(field)
					changed = true
				}
			}
			if changed {
				indexFile[newKeyValue] = indexEntries
				fi.Save(ic.Field)
			}
			continue
		}

		// Change in index field, remove the old entry and add the new entry
		oldIndexEntries := indexFile[oldKeyValue]
		oldEntryIndex := slices.IndexFunc(oldIndexEntries, idComparer)
		oldIndexEntries = append(oldIndexEntries[:oldEntryIndex], oldIndexEntries[oldEntryIndex+1:]...)
		indexFile[oldKeyValue] = oldIndexEntries
		indexFile[newKeyValue] = append(indexFile[newKeyValue], createIndexEntry(e, ic.Field, ic.Include))
		fi.Save(ic.Field)
	}
	return nil
}

func (fi *fileIndex[T]) Delete(prev T) error {
	idComparer := func(item *IndexEntry) bool {
		return prev.GetID() == item.ID
	}
	for _, ic := range fi.indexConfigs {
		index := fi.indexes[ic.Field][prev.GetValue(ic.Field)]
		ci := slices.IndexFunc(index, idComparer)
		index = append(index[:ci], index[ci+1:]...)
		fi.indexes[ic.Field][prev.GetValue(ic.Field)] = index
		fi.Save(ic.Field)
	}
	return nil
}

func (fi *fileIndex[T]) FindId(field string, value string) int {
	if index, ok := fi.indexes[field][value]; ok {
		if len(index) > 0 {
			return index[0].ID
		}
	}
	return 0
}

func (fi *fileIndex[T]) SearchId(field string, value string) []int {
	if index, ok := fi.indexes[field][value]; ok {
		ids := make([]int, len(index))
		for i, v := range index {
			ids[i] = v.ID
		}
		return ids
	}
	return nil
}

func (fi *fileIndex[T]) SearchIndex(field string, value string) []*IndexEntry {
	if index, ok := fi.indexes[field][value]; ok {
		return index
	}
	return nil
}

func (fi *fileIndex[T]) SearchAllIndex(field string) []*IndexEntry {
	var result []*IndexEntry
	for _, v := range fi.indexes[field] {
		result = append(result, v...)
	}
	return result
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
	ic := fi.GetIndexConfig(name)
	if ic == nil {
		return fmt.Errorf("index config not found: %s", name)
	}
	file, err := os.OpenFile(fi.GetPath(name), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	fi.indexes[name] = make(map[string][]*IndexEntry)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var id int
		var value string
		parts := strings.Split(line, "\t")
		if len(parts) != len(ic.Include)+2 {
			return &InvalidIndexError{Message: "invalid index file format"}
		}
		value = parts[0]
		fmt.Sscanf(parts[1], "%d", &id)
		others := make(map[string]string)
		for i := 2; i < len(parts); i++ {
			others[ic.Include[i-2]] = parts[i]
		}
		entry := &IndexEntry{Value: value, ID: id, Others: others}
		fi.indexes[name][value] = append(fi.indexes[name][value], entry)
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
		for _, entry := range v {
			file.WriteString(fmt.Sprintf("%s\t%d", k, entry.ID))
			for _, i := range fi.GetIndexConfig(name).Include {
				file.WriteString(fmt.Sprintf("\t%s", entry.Others[i]))
			}
			file.WriteString("\n")
		}
	}
	return nil
}

func (fi *fileIndex[T]) GetPath(name string) string {
	return filepath.FromSlash(fi.path + "/_" + name + ".idx")
}

func (fi *fileIndex[T]) GetIndexConfig(name string) *FileIndexConfig {
	for _, ic := range fi.indexConfigs {
		if ic.Field == name {
			return &ic
		}
	}
	return nil
}

func (fi *fileIndex[T]) FindMaxIdAndCount() (int, int) {
	max := 0
	count := 0
	for _, index := range fi.indexes {
		for _, entries := range index {
			for _, entry := range entries {
				if entry.ID > max {
					max = entry.ID
				}
			}
			count = count + len(entries)
		}
		break
	}
	return max, count
}

func (fi *fileIndex[T]) ListAllIds() []int {
	ids := make([]int, 0)
	if len(fi.indexConfigs) == 0 {
		return ids
	}
	index := fi.indexes[fi.indexConfigs[0].Field]
	for _, entries := range index {
		for _, entry := range entries {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}
