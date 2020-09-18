package plugins

import (
	"os"
	"time"
)

// UtilPackage contains useful generic methods
type UtilPackage struct {}

// GetUtil returns an UtilPackage
func GetUtil() *UtilPackage {
	return &UtilPackage{}
}

// UtilResponse contains the return values
type UtilResponse struct {
	Error    error
	Status   bool
	Value    string
}

// Sleep call Sleep method for N seconds
func (c *UtilPackage) Sleep(seconds int) UtilResponse {
	time.Sleep(time.Duration(seconds) * time.Second)
	return UtilResponse{
		Error: nil,
		Status: true,
	}
}

// Getenv returns an environment variable if it exists
func (c *UtilPackage) Getenv(name string) UtilResponse {
	return UtilResponse{
		Error: nil,
		Status: true,
		Value: os.Getenv(name),
	}
}