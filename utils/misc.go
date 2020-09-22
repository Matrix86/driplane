package utils

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io"
	"os"
)

// IsFlagPassed returns true if the flag is present on the command line
func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// FileExists returns true if the file is present on disk
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists returns true if the directory is present on disk
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// MD5Sum calculate the MD5 Hash of a string or []byte
func MD5Sum(content interface{}) string {
	var hash [16]byte
	if v, ok := content.([]byte); ok {
		hash = md5.Sum(v)
	} else if v, ok := content.(string); ok {
		hash = md5.Sum([]byte(v))
	}
	return hex.EncodeToString(hash[:])
}

// Md5File calculate the MD5 Hash of a file
func Md5File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)[:]), nil
}