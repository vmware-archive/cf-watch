package filetree

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File interface {
	io.ReadCloser
	Basename() string
	Children() []File
	Mode() (os.FileMode, error)
	Size() (int64, error)
}

type Tree struct{}

func (t *Tree) New(path string) (File, error) {
	osFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var children []File
	fileInfo, err := osFile.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		childrenInfo, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, childInfo := range childrenInfo {
			child, err := t.New(filepath.Join(path, childInfo.Name()))
			if err != nil {
				return nil, err
			}

			children = append(children, child)
		}
	}
	return &file{
		File:     osFile,
		children: children,
	}, nil
}

type file struct {
	*os.File
	children []File
}

func (f *file) Basename() string {
	return filepath.Base(f.Name())
}

func (f *file) Children() []File {
	return f.children
}

func (f *file) Mode() (os.FileMode, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return fileInfo.Mode(), nil
}

func (f *file) Size() (int64, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
