package plugins

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/evilsocket/islazy/log"
)

type httpPackage struct {}

func GetHttp() *httpPackage {
	return &httpPackage{}
}

type httpResponse struct {
	Error    error
	Response *http.Response
	Raw      []byte
	Body     string
}

func (c *httpPackage) createRequest(method string, uri string, headers interface{}, data interface{}) (*http.Request, error) {
	var reader io.Reader

	if data != nil {
		switch t := data.(type) {
		case string:
			reader = bytes.NewBufferString(t)

		case map[string]string:
			dt := url.Values{}
			for k, v := range t {
				dt.Set(k, v)
			}
			reader = bytes.NewBufferString(dt.Encode())

		case *bytes.Buffer:
			reader = t

		default:
			return nil, fmt.Errorf("wrong data type")
		}
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		if h, ok := headers.(map[string]interface{}); ok {
			for name, value := range h {
				if v, ok := value.(string); ok {
					req.Header.Add(name, v)
				}
			}
		}
	}

	return req, nil
}

func (c *httpPackage) Request(method string, uri string, headers interface{}, data interface{}) httpResponse {
	client := &http.Client{}

	req, err := c.createRequest(method, uri, headers, data)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return httpResponse{Error: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return httpResponse{Error: err}
	}
	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpResponse{Error: err}
	}

	return httpResponse{
		Error:    nil,
		Response: resp,
		Raw:      raw,
		Body:     string(raw),
	}
}

func (c *httpPackage) Get(url string, headers map[string]string) httpResponse {
	return c.Request("GET", url, headers, nil)
}

func (c *httpPackage) Post(url string, headers map[string]string, data interface{}) httpResponse {
	return c.Request("POST", url, headers, data)
}

func (c *httpPackage) DownloadFile(filepath string, method string, uri string, headers interface{}, data interface{}) httpResponse {
	client := &http.Client{}

	req, err := c.createRequest(method, uri, headers, data)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return httpResponse{Error: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return httpResponse{Error: err}
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		log.Error("http.DownloadFile: %s", err)
		return httpResponse{Error: err}
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Error("%s", err)
		return httpResponse{Error: err}
	}

	return httpResponse{
		Error:    nil,
		Response: resp,
	}
}

func (c *httpPackage) UploadFile(filename string, fieldname string, method string, uri string, headers interface{}, data interface{}) httpResponse {
	client := &http.Client{}

	file, err := os.Open(filename)
	if err != nil {
		log.Error("%s", err)
		return httpResponse{Error: err}
	}
	defer file.Close()

	bodyfile := bytes.Buffer{}
	writer := multipart.NewWriter(&bodyfile)

	part, err := writer.CreateFormFile(fieldname, filepath.Base(filename))
	if err != nil {
		log.Error("%s", err)
		return httpResponse{Error: err}
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Error("http.createRequest : io.Copy : %s", err)
		return httpResponse{Error: err}
	}

	if v, ok := data.(map[string]interface{}); ok {
		for key, val := range v {
			if v, ok := val.(string); ok {
				_ = writer.WriteField(key, v)
			}
		}
	}

	err = writer.Close()
	if err != nil {
		log.Error("http.UploadFile : writer.Close : %s", err)
		return httpResponse{Error: err}
	}

	req, err := c.createRequest(method, uri, headers, &bodyfile)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return httpResponse{Error: err}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		log.Error("http.Upload client do : %s", err)
		return httpResponse{Error: err}
	}
	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpResponse{Error: err}
	}

	return httpResponse{
		Error:    nil,
		Response: resp,
		Raw:      raw,
		Body:     string(raw),
	}
}
