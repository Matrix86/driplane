package plugins

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestUtilPluginSleepMethod(t *testing.T) {
	u := GetUtil()

	res := u.Sleep(0)
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}
}

func TestUtilPluginGetEnvMethod(t *testing.T) {
	u := GetUtil()

	res := u.Getenv("ENVTESTVAR")
	if res.Value != "" {
		t.Errorf("the env var should be empty")
	}

	os.Setenv("ENVTESTVAR", "1")

	res = u.Getenv("ENVTESTVAR")
	if res.Value != "1" {
		t.Errorf("the env var should contain '1'")
	}
}

func TestUtilPluginMd5Method(t *testing.T) {
	u := GetUtil()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	res := u.Md5File("/tmp/notexistentfile_")
	if res.Status == true {
		t.Errorf("Status should be false")
	}

	res = u.Md5File(file.Name())
	if res.Status == false {
		t.Errorf("Status should be true")
	}

	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if res.Value != expected {
		t.Errorf("Value has a bad value: expected=%s had=%s", expected, res.Value)
	}
}