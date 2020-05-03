package plugins

import (
	"fmt"
	"io"
	"os"
)

type filePackage struct {
}

var fp = filePackage{}

func GetFile() filePackage {
	return fp
}

type fileResponse struct {
	Error    error
	Status   bool
}

func (c *filePackage) Copy(src, dst string) (fileResponse) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fileResponse{Error: fmt.Errorf("%s is not a regular file", src)}
	}

	source, err := os.Open(src)
	if err != nil {
		return fileResponse{Error: err}
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fileResponse{Error: err}
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fileResponse{Error: err}
	}

	return fileResponse{Status: true}
}

func (c *filePackage) Move(src, dst string) (fileResponse) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fileResponse{Error: fmt.Errorf("%s is not a regular file", src)}
	}

	err = os.Rename(src, dst)
	if err != nil {
		return fileResponse{Error: err}
	}
	return fileResponse{Status: true}
}

func (c *filePackage) Truncate(filename string, size int64) fileResponse {
	sourceFileStat, err := os.Stat(filename)
	if err != nil {
		return fileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fileResponse{Error: fmt.Errorf("%s is not a regular file", filename)}
	}

	err = os.Truncate(filename, size)
	if err != nil {
		return fileResponse{Error: err}
	}
	return fileResponse{Status: true}
}

func (c *filePackage) Delete(filename string) fileResponse {
	sourceFileStat, err := os.Stat(filename)
	if err != nil {
		return fileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fileResponse{Error: fmt.Errorf("%s is not a regular file", filename)}
	}

	err = os.Remove(filename)
	if err != nil {
		return fileResponse{Error: err}
	}
	return fileResponse{Status: true}
}

func (c *filePackage) Exists(filename string) fileResponse {
	sourceFileStat, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return fileResponse{Status: false}
	}
	return fileResponse{Status: !sourceFileStat.IsDir()}
}