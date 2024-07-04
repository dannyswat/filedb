package filedb

type FileEntity interface {
	GetID() int
	SetID(id int)
	GetValue(field string) string
}
