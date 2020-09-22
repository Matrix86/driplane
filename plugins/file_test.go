package plugins

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFilePluginCopyMethod(t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	dst := path.Join(os.TempDir(), "test1")
	res := h.Copy(file.Name(), dst)
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}

	defer os.Remove(dst)

	info, err := os.Stat(dst)
	if os.IsNotExist(err) {
		t.Errorf("file not copied" )
	}
	if !info.IsDir() == false {
		t.Errorf("file is a directory" )
	}
}

func TestFilePluginMoveMethod( t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}

	dst := path.Join(os.TempDir(), "test1")
	res := h.Move(file.Name(), dst)
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}

	defer os.Remove(dst)

	info, err := os.Stat(dst)
	if os.IsNotExist(err) {
		t.Errorf("file not moved" )
	}
	if !info.IsDir() == false {
		t.Errorf("file is a directory" )
	}
}

func TestFilePluginTruncateMethod(t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	text := []byte("This is a test!")
	if _, err = file.Write(text); err != nil {
		t.Error("can't write on file")
	}

	res := h.Truncate(file.Name(), 0)
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}

	info, err := os.Stat(file.Name())
	if os.IsNotExist(err) {
		t.Errorf("file not exist" )
	}

	if info.Size() != 0 {
		t.Errorf("size should be 0" )
	}
}

func TestFilePluginDeleteMethod(t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	res := h.Delete("/tmp/this_file_doesnt_exist")
	if res.Status == true {
		t.Errorf("bad response: expected=%t had=%t", false, res.Status )
	}

	res = h.Delete(file.Name())
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}

	_, err = os.Stat(file.Name())
	if !os.IsNotExist(err) {
		t.Errorf("file %s not deleted", file.Name() )
	}
}

func TestFilePluginExistsMethod(t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	res := h.Exists("/tmp/this_file_doesnt_exist")
	if res.Status == true {
		t.Errorf("bad response: expected=%t had=%t", false, res.Status )
	}

	res = h.Exists(file.Name())
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status )
	}
}

func TestFilePluginAppendStringMethod(t *testing.T) {
	h := GetFile()

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file" )
	}
	defer os.Remove(file.Name())

	res := h.AppendString(file.Name(), "test")
	if res.Status != true {
		t.Errorf("bad response: expected=%t had=%t", false, res.Status )
	}

	info, err := os.Stat(file.Name())
	if os.IsNotExist(err) {
		t.Errorf("file not exist" )
	}

	if info.Size() != 4 {
		t.Errorf("size should be 4" )
	}
}

