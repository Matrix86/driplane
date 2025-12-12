package plugins

import (
	"fmt"
	"io"
	"os"
)

// FilePackage contains file manipulation methods
type FilePackage struct{}

// GetFile returns the FilePackage struct
func GetFile() *FilePackage {
	return &FilePackage{}
}

// FileResponse contains the return values
type FileResponse struct {
	Error  error
	Status bool
	Binary []byte
	String string
}

func (c *FilePackage) Read(fileName string) FileResponse {
	resp := FileResponse{}

	resp.Binary, resp.Error = os.ReadFile(fileName)
	resp.Status = resp.Error == nil
	resp.String = string(resp.Binary)

	return resp
}

func (c *FilePackage) Write(fileName string, data []byte) FileResponse {
	resp := FileResponse{}

	resp.Error = os.WriteFile(fileName, data, os.ModePerm)
	resp.Status = resp.Error == nil

	return resp
}

// Copy handles the copy of a file to another file
func (c *FilePackage) Copy(src, dst string) FileResponse {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return FileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return FileResponse{Error: fmt.Errorf("%s is not a regular file", src)}
	}

	source, err := os.Open(src)
	if err != nil {
		return FileResponse{Error: err}
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return FileResponse{Error: err}
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return FileResponse{Error: err}
	}

	return FileResponse{Status: true}
}

// Move renames a file
func (c *FilePackage) Move(src, dst string) FileResponse {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return FileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return FileResponse{Error: fmt.Errorf("%s is not a regular file", src)}
	}

	err = os.Rename(src, dst)
	if err != nil {
		return FileResponse{Error: err}
	}
	return FileResponse{Status: true}
}

// Truncate set file length to size
func (c *FilePackage) Truncate(filename string, size int64) FileResponse {
	sourceFileStat, err := os.Stat(filename)
	if err != nil {
		return FileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return FileResponse{Error: fmt.Errorf("%s is not a regular file", filename)}
	}

	err = os.Truncate(filename, size)
	if err != nil {
		return FileResponse{Error: err}
	}
	return FileResponse{Status: true}
}

// Delete removes a file
func (c *FilePackage) Delete(filename string) FileResponse {
	sourceFileStat, err := os.Stat(filename)
	if err != nil {
		return FileResponse{Error: err}
	}

	if !sourceFileStat.Mode().IsRegular() {
		return FileResponse{Error: fmt.Errorf("%s is not a regular file", filename)}
	}

	err = os.Remove(filename)
	if err != nil {
		return FileResponse{Error: err}
	}
	return FileResponse{Status: true}
}

// Exists returns true if the file exists
func (c *FilePackage) Exists(filename string) FileResponse {
	sourceFileStat, err := os.Stat(filename)
	if err != nil {
		return FileResponse{Error: err}
	}
	if !sourceFileStat.Mode().IsRegular() {
		return FileResponse{Error: fmt.Errorf("%s is not a regular file", filename)}
	}

	return FileResponse{Status: !sourceFileStat.IsDir()}
}

// AppendString opens the file in appending mode and write text on it
func (c *FilePackage) AppendString(filename string, text string) FileResponse {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return FileResponse{Error: err, Status: false}
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		return FileResponse{Error: err, Status: false}
	}
	return FileResponse{Status: true}
}
