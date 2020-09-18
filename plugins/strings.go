package plugins

import (
	"strings"
)

// StringsPackage contains string manipulation methods
type StringsPackage struct {}

// GetStrings returns a StringsPackage
func GetStrings() *StringsPackage {
	return &StringsPackage{}
}

// StringsResponse contains the return values
type StringsResponse struct {
	Error    error
	Status   bool
}

// StartsWith returns true if a string start with a substring
func (c *StringsPackage) StartsWith(str, substr string) StringsResponse {
	ret := strings.HasPrefix(str, substr)
	return StringsResponse{
		Error: nil,
		Status: ret,
	}
}