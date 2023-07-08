package esbuildfs

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"
)

var _ fs.FS = &FS{}
var _ FSEvent = &FS{}

type FS struct {
	mux              sync.RWMutex
	files            map[string]*file
	onChangedHandler []EventHandler
}

func New() *FS {
	return &FS{
		files:            make(map[string]*file),
		onChangedHandler: make([]EventHandler, 0),
	}
}

func (fsys *FS) Open(name string) (fs.File, error) {
	fsys.mux.RLock()
	defer fsys.mux.RUnlock()

	if file, ok := fsys.files[name]; ok {
		return file.Clone(), nil
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

	fsys.files[name] = &file{
		name:         name,
		contents:     bytes.NewBuffer(buffer),
		size:         int64(len(buffer)),
		modifiedTime: time.Now(),
	}
	emitOnChanged(fsys, fsys.files[name])

	return nil
}

func (fsys *FS) OnChanged(handler EventHandler) {
	fsys.onChangedHandler = append(fsys.onChangedHandler, handler)
}

func emitOnChanged(fsys *FS, file fs.File) {
	for idx := range fsys.onChangedHandler {
		fsys.onChangedHandler[idx](file)
	}
}
