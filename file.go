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
	contents     *bytes.Buffer
	size         int64
	modifiedTime time.Time
	closed       bool
}

func (f *file) Stat() (fs.FileInfo, error) {
	if f.closed {
		return nil, fs.ErrClosed
	}

	return f, nil
}

func (f *file) Read(buffer []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}

	return f.contents.Read(buffer)
}

func (f *file) Close() error {
	if f.closed {
		return fs.ErrClosed
	}

	f.closed = true
	return nil
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.size
}

func (f *file) Mode() fs.FileMode {
	return fs.ModeTemporary
}

func (f *file) ModTime() time.Time {
	return f.modifiedTime
}

func (f *file) IsDir() bool {
	return false
}

func (f *file) Sys() any {
	return nil
}

func (f *file) Clone() *file {
	return &file{
		name:         f.name,
		contents:     bytes.NewBuffer(f.contents.Bytes()),
		size:         f.size,
		modifiedTime: f.modifiedTime,
	}
}
