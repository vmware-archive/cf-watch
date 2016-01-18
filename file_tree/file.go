package file_tree

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	*os.File
	children []*File
}

func New(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}

	var children []*File
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %s", err)
	}

	if fileInfo.IsDir() {
		childrenInfo, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to list directory contents: %s", err)
		}

		workingDirName, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %s", err)
		}

		workingDir, err := os.Open(workingDirName)
		if err != nil {
			return nil, fmt.Errorf("failed to open current working directory: %s", err)
		}

		file.Chdir()
		defer workingDir.Chdir()
		for _, childInfo := range childrenInfo {
			child, err := New(filepath.Base(childInfo.Name()))
			if err != nil {
				return nil, err
			}

			children = append(children, child)
		}
	}
	return &File{
		File:     file,
		children: children,
	}, nil
}

func (f *File) BaseName() string {
	return filepath.Base(f.Name())
}

func (f *File) Children() []*File {
	return f.children
}
