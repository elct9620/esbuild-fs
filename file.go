package esbuildfs

import (
	"bytes"
	"io/fs"
	"time"
)

var _ fs.File = &file{}
var _ fs.FileInfo = &file{}

type file struct {
	name         string
	contents     *bytes.Reader
	modifiedTime time.Time
}

func (f file) Stat() (fs.FileInfo, error) {
	return &f, nil
}

func (f file) Read(buffer []byte) (int, error) {
	return f.contents.Read(buffer)
}

func (f file) Close() error {
	return nil
}

func (f file) Name() string {
	return f.name
}

func (f file) Size() int64 {
	return int64(f.contents.Size())
}

func (f file) Mode() fs.FileMode {
	return fs.ModeTemporary
}

func (f file) ModTime() time.Time {
	return f.modifiedTime
}

func (f file) IsDir() bool {
	return false
}

func (f file) Sys() any {
	return nil
}
