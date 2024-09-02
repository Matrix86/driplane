package plugins

import (
	"github.com/Matrix86/driplane/utils"
	"os"
	"time"
)

// UtilPackage contains useful generic methods
type UtilPackage struct{}

// GetUtil returns an UtilPackage
func GetUtil() *UtilPackage {
	return &UtilPackage{}
}

// UtilResponse contains the return values
type UtilResponse struct {
	Error  error
	Status bool
	Value  string
}

// Sleep call Sleep method for N seconds
func (c *UtilPackage) Sleep(seconds int) UtilResponse {
	time.Sleep(time.Duration(seconds) * time.Second)
	return UtilResponse{
		Error:  nil,
		Status: true,
	}
}

// Getenv returns an environment variable if it exists
func (c *UtilPackage) Getenv(name string) UtilResponse {
	return UtilResponse{
		Error:  nil,
		Status: true,
		Value:  os.Getenv(name),
	}
}

// Md5File returns the MD5 hash of the file
func (c *UtilPackage) Md5File(filename string) UtilResponse {
	hash, err := utils.Md5File(filename)
	if err != nil {
		return UtilResponse{
			Error:  err,
			Status: false,
			Value:  "",
		}
	}
	return UtilResponse{
		Error:  nil,
		Status: true,
		Value:  hash,
	}
}

// Sha1File returns the SHA1 hash of the file
func (c *UtilPackage) Sha1File(filename string) UtilResponse {
	hash, err := utils.Sha1File(filename)
	if err != nil {
		return UtilResponse{
			Error:  err,
			Status: false,
			Value:  "",
		}
	}
	return UtilResponse{
		Error:  nil,
		Status: true,
		Value:  hash,
	}
}

// Sha256File returns the SHA256 hash of the file
func (c *UtilPackage) Sha256File(filename string) UtilResponse {
	hash, err := utils.Sha256File(filename)
	if err != nil {
		return UtilResponse{
			Error:  err,
			Status: false,
			Value:  "",
		}
	}
	return UtilResponse{
		Error:  nil,
		Status: true,
		Value:  hash,
	}
}

// Sha512File returns the SHA512 hash of the file
func (c *UtilPackage) Sha512File(filename string) UtilResponse {
	hash, err := utils.Sha512File(filename)
	if err != nil {
		return UtilResponse{
			Error:  err,
			Status: false,
			Value:  "",
		}
	}
	return UtilResponse{
		Error:  nil,
		Status: true,
		Value:  hash,
	}
}
