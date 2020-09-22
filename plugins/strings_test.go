package plugins

import "testing"

func TestStringsPluginStartsWithMethod(t *testing.T) {
	s := GetStrings()

	res := s.StartsWith("testing string", "test")
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}

	res = s.StartsWith("no testing string", "test")
	if res.Status == true {
		t.Errorf("bad response: expected=%t had=%t", false, res.Status )
	}
}
