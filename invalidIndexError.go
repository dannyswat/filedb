package filedb

type InvalidIndexError struct {
	Message string
}

func (e *InvalidIndexError) Error() string {
	return e.Message
}
