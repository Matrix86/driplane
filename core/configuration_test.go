package core

import (
	"errors"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		FileContent    string
		ExpectedConfig *Configuration
		ExpectedError  error
	}
	tests := []Test{
		{"FailOpenFile", path.Join(os.TempDir(), "file_not_exist_for_sure"), false, "", &Configuration{}, errors.New("")},
		{"NotYamlFile", path.Join(os.TempDir(), "configuration_file_test.yml"), true, "asd {}", &Configuration{}, errors.New("")},
		{"EmptyYamlFile", path.Join(os.TempDir(), "configuration_file_test.yml"), true, "", &Configuration{flat: map[string]string{}}, nil},
		{"GoodYamlFile", path.Join(os.TempDir(), "configuration_file_test.yml"), true, "general:\n  config: true", &Configuration{flat: map[string]string{"general.config": "true"}}, nil},
	}
	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte(v.FileContent)); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		res, err := LoadConfiguration(v.Filename)
		if v.ExpectedError != nil {
			if err == nil || errors.Is(v.ExpectedError, err) {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
			}
		} else {
			if err != nil {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
			}
			if res.FilePath != v.Filename {
				t.Errorf("%s: wrong res.Filepath: expected=%#v had=%#v", v.Name, v.ExpectedConfig.FilePath, res.FilePath)
			}
			if reflect.DeepEqual(v.ExpectedConfig.flat, res.flat) == false {
				t.Errorf("%s: wrong res.flat: expected=%#v had=%#v", v.Name, v.ExpectedConfig.flat, res.flat)
			}
		}
	}
}

func TestConfiguration_Get(t *testing.T) {
	type Test struct {
		Name           string
		ParamName      string
		Config         *Configuration
		ExpectedReturn string
	}
	tests := []Test{
		{"ConfigExist", "general.test", &Configuration{flat: map[string]string{"general.test": "true"}}, "true"},
		{"ConfigNotExist", "name", &Configuration{}, ""},
	}
	for _, v := range tests {
		res := v.Config.Get(v.ParamName)
		if res != v.ExpectedReturn {
			t.Errorf("%s: wrong value: expected=%#v had=%#v", v.Name, v.ExpectedReturn, res)
		}
	}
}

func TestConfiguration_Set(t *testing.T) {
	type Test struct {
		Name          string
		ParamName     string
		ParamValue    string
		Config        *Configuration
		ExpectedError error
	}
	tests := []Test{
		{"SetConfig", "general.test", "true", &Configuration{flat: map[string]string{}}, nil},
		{"OverwriteConfig", "general.test", "true", &Configuration{flat: map[string]string{"general.test": "true"}}, nil},
	}
	for _, v := range tests {
		err := v.Config.Set(v.ParamName, v.ParamValue)
		if v.ExpectedError != nil {
			if err == nil || errors.Is(v.ExpectedError, err) {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
			}
		} else {
			if err != nil {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
			}
			if v.Config.flat[v.ParamName] != v.ParamValue {
				t.Errorf("%s: wrong value: expected=%#v had=%#v", v.Name, v.ParamValue, v.Config.flat[v.ParamName])
			}
		}
	}
}

func TestConfiguration_GetConfig(t *testing.T) {
	type Test struct {
		Name           string
		ExpectedConfig *Configuration
	}
	tests := []Test{
		{"ConfigReturned", &Configuration{flat: map[string]string{"general.test": "true"}} },
	}
	for _, v := range tests {
		res := v.ExpectedConfig.GetConfig()
		if reflect.DeepEqual(v.ExpectedConfig.flat, res) == false {
			t.Errorf("%s: wrong res.flat: expected=%#v had=%#v", v.Name, v.ExpectedConfig.flat, res)
		}
	}
}
