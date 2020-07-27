package plugins

import (
	"strings"
)

type stringsPackage struct {}

func GetStrings() *stringsPackage {
	return &stringsPackage{}
}

type stringsResponse struct {
	Error    error
	Status   bool
}

func (c *stringsPackage) StartsWith(str, substr string) stringsResponse {
	ret := strings.HasPrefix(str, substr)
	return stringsResponse{
		Error: nil,
		Status: ret,
	}
}