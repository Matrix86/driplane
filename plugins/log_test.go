package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/evilsocket/islazy/log"
)

func TestLogPackage_Debug(t *testing.T) {
	l := GetLog()

	logfile := path.Join(os.TempDir(), "log.unit_test")
	log.Output = logfile
	log.Format = "{message}"
	log.Level = log.DEBUG
	log.Open()
	defer log.Close()
	defer os.Remove(logfile)

	message := "test debug"
	expected := fmt.Sprintf("%s\n", message)
	l.Debug("%s", message)

	dat, _ := ioutil.ReadFile(logfile)
	if string(dat) != expected {
		t.Errorf("wrong string: expected=%#v had=%#v", expected, string(dat))
	}
}

func TestLogPackage_Info(t *testing.T) {
	l := GetLog()
	logfile := path.Join(os.TempDir(), "log.unit_test")
	log.Output = logfile
	log.Format = "{message}"
	log.Level = log.DEBUG
	log.Open()
	defer log.Close()
	defer os.Remove(logfile)

	message := "test debug"
	expected := fmt.Sprintf("%s\n", message)
	l.Info("%s", message)

	dat, _ := ioutil.ReadFile(logfile)
	if string(dat) != expected {
		t.Errorf("wrong string: expected=%#v had=%#v", expected, string(dat))
	}
}


func TestLogPackage_Error(t *testing.T) {
	l := GetLog()

	logfile := path.Join(os.TempDir(), "log.unit_test")
	log.Output = logfile
	log.Format = "{message}"
	log.Level = log.DEBUG
	log.Open()
	defer log.Close()
	defer os.Remove(logfile)

	message := "test debug"
	expected := fmt.Sprintf("%s\n", message)
	l.Error("%s", message)

	dat, _ := ioutil.ReadFile(logfile)
	if string(dat) != expected {
		t.Errorf("wrong string: expected=%#v had=%#v", expected, string(dat))
	}
}