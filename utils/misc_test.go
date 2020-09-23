package utils

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestIsFlagPassed(t *testing.T) {
	if IsFlagPassed("test") {
		t.Errorf("it should return false")
	}

	if IsFlagPassed("test.timeout") == false {
		t.Errorf("it should return true")
	}
}

func TestFileExists(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	if FileExists(file.Name()) == false {
		t.Errorf("it should return true")
	}

	if FileExists("/false/file") == true {
		t.Errorf("it should return false")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := "/tmp/test_directory"
	os.Mkdir(tmpDir, os.ModeDir)
	defer os.Remove(tmpDir)

	if DirExists(tmpDir) == false {
		t.Errorf("it should return true")
	}

	if DirExists("/false/directory") {
		t.Errorf("it should return false")
	}
}

func TestMD5Sum(t *testing.T) {
	type Test struct {
		Key      interface{}
		Expected string
	}
	tests := []Test{
		Test{"test1", "5a105e8b9d40e1329780d62ea2265d8a"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{[]byte("test"), "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, v := range tests {
		had := MD5Sum(v.Key)
		if had != v.Expected {
			t.Errorf("wrong md5: expected=%s had=%s", v.Expected, had)
		}
	}
}

func TestMd5File(t *testing.T) {
	type Test struct {
		Filename       string
		CreateFile     bool
		ExpectedString string
		ExpectedError  string
	}
	tests := []Test{
		{path.Join(os.TempDir(), "testmd5"), true, "702edca0b2181c15d457eacac39de39b", ""},
		{path.Join(os.TempDir(), "notexist"), false, "", "open /tmp/notexist: no such file or directory"},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("cannot create a temporary file")
			}
			defer os.Remove(v.Filename)

			text := []byte("This is a test!")
			if _, err = file.Write(text); err != nil {
				t.Error("can't write on file")
			}
		}

		had, err := Md5File(v.Filename)
		if v.ExpectedError == "" && err != nil {
			t.Errorf("wrong error: expected=nil had=%#v", err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("wrong error: expected=%#v had=%#v", v.ExpectedError, err.Error())
		}
		if had != v.ExpectedString {
			t.Errorf("wrong hash: expected=%s had=%s", v.ExpectedString, had)
		}
	}
}
