package plugins

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

func TestHTTPPackage_Get(t *testing.T) {
	msg := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}))
	defer ts.Close()

	h := GetHTTP()

	res := h.Get(ts.URL, nil)
	if res.Status == false {
		t.Errorf("wrong string: expected=%#v had=%#v", true, res.Status)
	}
	if res.Body != msg {
		t.Errorf("wrong string: expected=%#v had=%#v", msg, res.Body)
	}
}

func TestHTTPPackage_Post(t *testing.T) {
	msg := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}))
	defer ts.Close()

	h := GetHTTP()

	res := h.Post(ts.URL, nil, nil)
	if res.Status == false {
		t.Errorf("wrong string: expected=%#v had=%#v", true, res.Status)
	}
	if res.Body != msg {
		t.Errorf("wrong string: expected=%#v had=%#v", msg, res.Body)
	}
}

func TestHTTPPackage_Request(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	type Test struct {
		Name             string
		Method           string
		URI              string
		Headers          interface{}
		Data             interface{}
		ExpectedResponse HTTPResponse
	}
	tests := []Test{
		{"FailCreateRequest", "GET", ts.URL, nil, []byte{}, HTTPResponse{Status: false, Error: fmt.Errorf("wrong data type")}},
		{"FailDoRequest", "GET", "wrongurl", nil, nil, HTTPResponse{Status: false, Error: fmt.Errorf("Get wrongurl: unsupported protocol scheme ")}},
		{"RequestDone", "GET", ts.URL, nil, nil, HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
		{"RequestWithHeaders", "GET", ts.URL, map[string]interface{}{"name": "value"}, nil, HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
		{"RequestWithDataAsString", "GET", ts.URL, nil, "name=value", HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
		{"RequestWithDataAsMap", "GET", ts.URL, nil, map[string]string{"name": "value"}, HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
		{"RequestWithDataAsBuffer", "GET", ts.URL, nil, bytes.NewBufferString("name=value"), HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
	}

	h := GetHTTP()

	for _, v := range tests {
		res := h.Request(v.Method, v.URI, v.Headers, v.Data)
		if res.Status != v.ExpectedResponse.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Status, res.Status)
		}
		if v.ExpectedResponse.Status == false && strings.Trim(res.Error.Error(), "\\\"") != strings.Trim(v.ExpectedResponse.Error.Error(), "\\\"") {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Error.Error(), res.Error.Error())
		}
		if v.ExpectedResponse.Status && v.ExpectedResponse.Body != res.Body {
			t.Errorf("%s: wrong body: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Body, res.Body)
		}
	}
}

func TestHTTPPackage_DownloadFile(t *testing.T) {
	msg := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}))
	defer ts.Close()

	type Test struct {
		Name             string
		Filepath         string
		Method           string
		URI              string
		Headers          interface{}
		Data             interface{}
		ExpectedResponse HTTPResponse
	}
	tests := []Test{
		{"FailCreateRequest", "", "GET", ts.URL, nil, []byte{}, HTTPResponse{Status: false, Error: fmt.Errorf("wrong data type")}},
		{"FailDoRequest", "", "GET", "wrongurl", nil, nil, HTTPResponse{Status: false, Error: fmt.Errorf("Get wrongurl: unsupported protocol scheme ")}},
		{"FailCreateFile", os.TempDir(), "GET", ts.URL, nil, nil, HTTPResponse{Status: false, Error: fmt.Errorf("open /tmp: is a directory")}},
		{"RequestDone", path.Join(os.TempDir(), "download_test"), "GET", ts.URL, nil, nil, HTTPResponse{Status: true, Error: nil, Body: "Hello, client"}},
	}

	h := GetHTTP()

	for _, v := range tests {
		res := h.DownloadFile(v.Filepath, v.Method, v.URI, v.Headers, v.Data)
		if res.Status != v.ExpectedResponse.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Status, res.Status)
		}
		if v.ExpectedResponse.Status == false && strings.Trim(res.Error.Error(), "\\\"") != strings.Trim(v.ExpectedResponse.Error.Error(), "\\\"") {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, strings.Trim(v.ExpectedResponse.Error.Error(), "\\\""), strings.Trim(res.Error.Error(), "\\\""))
		}
		if v.ExpectedResponse.Status {
			dat, _ := ioutil.ReadFile(v.Filepath)
			if string(dat) != msg {
				t.Errorf("%s : wrong file content: expected=%#v had=%#v", v.Name, msg, string(dat))
			}
		}
		if res.Status {
			os.Remove(v.Filepath)
		}
	}
}

func TestHTTPPackage_UploadFile(t *testing.T) {
	msg := "file content"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}))
	defer ts.Close()

	type Test struct {
		Name             string
		Filepath         string
		CreateFile       bool
		Fieldname        string
		Method           string
		URI              string
		Headers          interface{}
		Data             interface{}
		ExpectedResponse HTTPResponse
	}
	tests := []Test{
		{"FailOpenFile", path.Join(os.TempDir(), "file_not_exist.atall"), false, "", "POST", ts.URL, nil, nil, HTTPResponse{Status: false, Error: fmt.Errorf("open /tmp/file_not_exist.atall: no such file or directory")}},
		{"FailOpenFile2", os.TempDir(), false, "", "POST", ts.URL, nil, nil, HTTPResponse{Status: false, Error: fmt.Errorf("read /tmp: is a directory")}},
		{"FailDoRequest", path.Join(os.TempDir(), "upload_test.test"), true, "upload", "POST", "nourl", nil, []byte{}, HTTPResponse{Status: false, Error: fmt.Errorf("Post nourl: unsupported protocol scheme ")}},
		{"RequestDone", path.Join(os.TempDir(), "upload_test.test"), true, "upload", "POST", ts.URL, nil, nil, HTTPResponse{Status: true, Error: nil}},
	}

	h := GetHTTP()

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filepath)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filepath)

			if _, err = file.Write([]byte("file content")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		// UploadFile(filename string, fieldname string, method string, uri string, headers interface{}, data interface{})
		res := h.UploadFile(v.Filepath, v.Fieldname, v.Method, v.URI, v.Headers, v.Data)
		if res.Status != v.ExpectedResponse.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Status, res.Status)
		}
		if v.ExpectedResponse.Status == false && strings.Trim(res.Error.Error(), "\\\"") != strings.Trim(v.ExpectedResponse.Error.Error(), "\\\"") {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedResponse.Error.Error(), res.Error.Error())
		}
		if v.ExpectedResponse.Status {
			dat, _ := ioutil.ReadFile(v.Filepath)
			if string(dat) != msg {
				t.Errorf("%s : wrong file content: expected=%#v had=%#v", v.Name, msg, string(dat))
			}
		}
		if res.Status {
			os.Remove(v.Filepath)
		}
	}
}
