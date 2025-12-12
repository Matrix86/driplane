package plugins

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/evilsocket/islazy/log"
)

// HTTPPackage contains all the methods to perform http requests
type HTTPPackage struct{}

// GetHTTP returns the HTTPPackage
func GetHTTP() *HTTPPackage {
	return &HTTPPackage{}
}

// HTTPResponse contains the return values
type HTTPResponse struct {
	Error    error
	Response *http.Response
	Raw      []byte
	Body     string
	Status   bool
}

func (c *HTTPPackage) createRequest(method string, uri string, headers interface{}, data interface{}) (*http.Request, error) {
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

// Request performs a HTTP request
func (c *HTTPPackage) Request(method string, uri string, headers interface{}, data interface{}) HTTPResponse {
	client := &http.Client{}

	req, err := c.createRequest(method, uri, headers, data)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return HTTPResponse{Error: err, Status: false}
	}

	resp, err := client.Do(req)
	if err != nil {
		return HTTPResponse{Error: err, Status: false}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{Error: err, Status: false}
	}

	return HTTPResponse{
		Error:    nil,
		Response: resp,
		Raw:      raw,
		Body:     string(raw),
		Status:   true,
	}
}

// Get performs a GET request
func (c *HTTPPackage) Get(url string, headers map[string]string) HTTPResponse {
	return c.Request("GET", url, headers, nil)
}

// Post performs a POST request
func (c *HTTPPackage) Post(url string, headers map[string]string, data interface{}) HTTPResponse {
	return c.Request("POST", url, headers, data)
}

// DownloadFile gets a file from an URL and store in on disk
func (c *HTTPPackage) DownloadFile(filepath string, method string, uri string, headers interface{}, data interface{}) HTTPResponse {
	client := &http.Client{}

	req, err := c.createRequest(method, uri, headers, data)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return HTTPResponse{Error: err, Status: false}
	}

	resp, err := client.Do(req)
	if err != nil {
		return HTTPResponse{Error: err, Status: false}
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		log.Error("http.DownloadFile: %s", err)
		return HTTPResponse{Error: err, Status: false}
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Error("%s", err)
		return HTTPResponse{Error: err, Status: false}
	}

	return HTTPResponse{
		Error:    nil,
		Response: resp,
		Status:   true,
	}
}

// UploadFile sends a file to an URL
func (c *HTTPPackage) UploadFile(filename string, fieldname string, method string, uri string, headers interface{}, data interface{}) HTTPResponse {
	client := &http.Client{}

	file, err := os.Open(filename)
	if err != nil {
		log.Error("%s", err)
		return HTTPResponse{Error: err, Status: false}
	}
	defer file.Close()

	bodyfile := bytes.Buffer{}
	writer := multipart.NewWriter(&bodyfile)

	part, err := writer.CreateFormFile(fieldname, filepath.Base(filename))
	if err != nil {
		log.Error("%s", err)
		return HTTPResponse{Error: err, Status: false}
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Error("http.createRequest : io.Copy : %s", err)
		return HTTPResponse{Error: err, Status: false}
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
		return HTTPResponse{Error: err, Status: false}
	}

	req, err := c.createRequest(method, uri, headers, &bodyfile)
	if err != nil {
		log.Error("http.createRequest : %s", err)
		return HTTPResponse{Error: err, Status: false}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		log.Error("http.Upload client do : %s", err)
		return HTTPResponse{Error: err, Status: false}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{Error: err, Status: false}
	}

	return HTTPResponse{
		Error:    nil,
		Response: resp,
		Raw:      raw,
		Body:     string(raw),
		Status:   true,
	}
}
