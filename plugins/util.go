package plugins

import (
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