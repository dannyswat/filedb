package filedb

type FileIndexConfig struct {
	Unique bool
	Field  string
}

type FileIndex[T FileEntity] interface {
	Init() error
	RebuildIndex(field string) error
	Insert(e T) error
	Update(e T) error
	Delete(id int) error
	FindId(field string, value string) (int, error)
	SearchId(field string, value string) ([]int, error)
}

type fileIndex[T FileEntity] struct {
	path         string
	indexConfigs []FileIndexConfig
	indexes      map[string]map[string][]int
}

func NewFileIndex[T FileEntity](path string, indexConfig []FileIndexConfig) FileIndex[T] {
	return &fileIndex[T]{path: path, indexConfigs: indexConfig, indexes: make(map[string]map[string][]int)}
}

func (fi *fileIndex[T]) Init() error {
	return nil
}

func (fi *fileIndex[T]) RebuildIndex(field string) error {
	return nil
}

func (fi *fileIndex[T]) Insert(e T) error {
	return nil
}

func (fi *fileIndex[T]) Update(e T) error {
	return nil
}

func (fi *fileIndex[T]) Delete(id int) error {
	return nil
}

func (fi *fileIndex[T]) FindId(field string, value string) (int, error) {
	return 0, nil
}

func (fi *fileIndex[T]) SearchId(field string, value string) ([]int, error) {
	return nil, nil
}
