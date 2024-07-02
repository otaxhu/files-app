package dto

import "io"

type SaveFile struct {
	Filename string
	Reader   io.Reader
	Len      int64
}

type GetFile struct {
	Filename string
	Reader   io.ReadCloser
	Len      int64
}

type GetFileInfo struct {
	Filename string
	Len      int64
}
