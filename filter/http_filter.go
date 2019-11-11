package filter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Matrix86/driplane/com"
)

type HttpFilter struct {
	FilterBase

	urlFromInput bool
	getBody      bool
	checkStatus  int
	urlString    string
	method       string
	headers      map[string]string
	dataPost     map[string]string

	params map[string]string
}

func NewHttpFilter(p map[string]string) (Filter, error) {
	f := &HttpFilter{
		params:       p,
		urlFromInput: true,
		getBody:      true,
		method:       "GET",
		headers:      make(map[string]string),
		dataPost:     make(map[string]string),
		checkStatus:  200,
	}

	if v, ok := f.params["useinput"]; ok && v == "false" {
		f.urlFromInput = false
	}
	if v, ok := f.params["url"]; ok {
		f.urlString = v
	}
	if v, ok := f.params["method"]; ok {
		f.method = v
	}
	if v, ok := f.params["headers"]; ok {
		err := json.Unmarshal([]byte(v), &f.headers)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := f.params["data"]; ok {
		err := json.Unmarshal([]byte(v), &f.dataPost)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := f.params["status"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		f.checkStatus = i
	}

	return f, nil
}

func (f *HttpFilter) DoFilter(msg *com.DataMessage) (bool, error) {
	var req *http.Request
	var err error

	text := msg.GetMessage()

	client := &http.Client{}

	urlString := ""
	if f.urlFromInput {
		urlString = text
	} else {
		urlString = f.urlString
	}

	if len(f.dataPost) > 0 {
		data := url.Values{}
		for key, value := range f.dataPost {
			data.Set(key, value)
		}

		req, err = http.NewRequest(f.method, urlString, bytes.NewBufferString(data.Encode()))
		if err != nil {
			return false, err
		}
	} else {
		req, err = http.NewRequest(f.method, urlString, nil)
		if err != nil {
			return false, err
		}
	}

	if len(f.headers) > 0 {
		for key, value := range f.headers {
			req.Header.Add(key, value)
		}
	}

	r, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer r.Body.Close()

	ret := false
	if f.checkStatus == r.StatusCode {
		ret = true
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return false, err
		}

		if f.getBody {
			msg.SetMessage(string(body))
		}
	}

	return ret, nil
}

// Set the name of the filter
func init() {
	register("http", NewHttpFilter)
}