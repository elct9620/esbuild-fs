package esbuildfs_test

import (
	"bytes"
	"io"
	"io/fs"
	"testing"

	esbuildfs "github.com/elct9620/esbuild-fs"
	"github.com/google/go-cmp/cmp"
)

type mockFile struct {
	Name    string
	Content io.Reader
}

type expectedFile struct {
	Name    string
	Content []byte
	Size    int64
}

func Test_FS_Open(t *testing.T) {
	tests := []struct {
		Name          string
		Files         []mockFile
		ExpectedFiles []expectedFile
	}{
		{
			Name: "single file",
			Files: []mockFile{
				{
					Name:    "app.js",
					Content: bytes.NewBufferString("hello world"),
				},
			},
			ExpectedFiles: []expectedFile{
				{
					Name:    "app.js",
					Content: []byte("hello world"),
					Size:    11,
				},
			},
		},
		{
			Name: "multiple files",
			Files: []mockFile{
				{
					Name:    "app.js",
					Content: bytes.NewBufferString("hello app.js"),
				},
				{
					Name:    "extension.js",
					Content: bytes.NewBufferString("hello extension.js"),
				},
			},
			ExpectedFiles: []expectedFile{
				{
					Name:    "app.js",
					Content: []byte("hello app.js"),
					Size:    12,
				},
				{
					Name:    "extension.js",
					Content: []byte("hello extension.js"),
					Size:    18,
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			fsys := givenAnNewFS(t)
			givenThereHaveManyFiles(t, fsys, tc.Files)

			for _, expectedFile := range tc.ExpectedFiles {
				file := whenOpenFile(t, fsys, expectedFile.Name)
				info := whenGetFileInfo(t, file)
				defer file.Close()

				thenContentToEqual(t, file, expectedFile.Content)
				thenNameToEqual(t, info, expectedFile.Name)
				thenSizeToEqual(t, info, expectedFile.Size)
			}
		})
	}
}

func givenAnNewFS(t *testing.T) *esbuildfs.FS {
	t.Helper()

	return esbuildfs.New()
}

func givenThereHaveManyFiles(t *testing.T, fsys *esbuildfs.FS, files []mockFile) {
	t.Helper()

	for _, file := range files {
		err := fsys.Write(file.Name, file.Content)
		if err != nil {
			t.Fatal("unable to write file", err)
		}
	}
}

func whenOpenFile(t *testing.T, fsys *esbuildfs.FS, name string) fs.File {
	t.Helper()

	file, err := fsys.Open(name)
	if err != nil {
		t.Error("unable to open file", err)
	}

	return file
}

func whenGetFileInfo(t *testing.T, file fs.File) fs.FileInfo {
	t.Helper()

	info, err := file.Stat()
	if err != nil {
		t.Fatal("unable to get file information", err)
	}

	return info
}

func thenContentToEqual(t *testing.T, file fs.File, expected []byte) {
	t.Helper()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Error("unable to read file", err)
	}

	if cmp.Equal(expected, content) {
		return
	}

	t.Error("content mismatch", cmp.Diff(expected, content))
}

func thenNameToEqual(t *testing.T, info fs.FileInfo, expected string) {
	t.Helper()

	name := info.Name()
	if cmp.Equal(expected, name) {
		return
	}

	t.Error("name mismatch", cmp.Diff(expected, name))
}

func thenSizeToEqual(t *testing.T, info fs.FileInfo, expected int64) {
	t.Helper()

	size := info.Size()
	if cmp.Equal(expected, size) {
		return
	}

	t.Error("size mismatch", cmp.Diff(expected, size))
}
