package models

import (
	"io"
)

type File struct {
	Id       string
	Reader   io.ReadCloser
	Filename string
	Len      int64
}
