package esbuildfs

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"
)

var _ fs.FS = &FS{}

type FS struct {
	mux   sync.RWMutex
	files map[string]file
}

func New() *FS {
	return &FS{
		files: make(map[string]file),
	}
}

func (fsys *FS) Open(name string) (fs.File, error) {
	fsys.mux.RLock()
	defer fsys.mux.RUnlock()

	if file, ok := fsys.files[name]; ok {
		return &file, nil
	}

	return nil, fs.ErrNotExist
}

func (fsys *FS) Write(name string, content io.Reader) error {
	fsys.mux.Lock()
	defer fsys.mux.Unlock()

	buffer, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	fsys.files[name] = file{
		name:         name,
		contents:     bytes.NewReader(buffer),
		modifiedTime: time.Now(),
	}
	return nil
}
