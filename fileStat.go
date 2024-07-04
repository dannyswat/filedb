package filedb

type FileStat interface {
	Init() error
	GetNextID() int
	GetCount() int
	AddCount(c int) error
}

type fileStat struct {
	path   string
	nextID int
	count  int
}

func NewFileStat(path string) FileStat {
	return &fileStat{path: path, nextID: 1, count: 0}
}

func (fs *fileStat) Init() error {
	return nil
}

func (fs *fileStat) GetNextID() int {
	id := fs.nextID
	fs.nextID++
	return id
}

func (fs *fileStat) GetCount() int {
	return fs.count
}

func (fs *fileStat) AddCount(c int) error {
	fs.count += c
	return nil
}
