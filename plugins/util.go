package plugins

import (
	"os"
	"time"
)

type utilPackage struct {}

func GetUtil() *utilPackage {
	return &utilPackage{}
}

type utilResponse struct {
	Error    error
	Status   bool
}

func (c *utilPackage) Sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func (c *utilPackage) Getenv(name string) string {
	return os.Getenv(name)
}