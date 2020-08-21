package filters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
)

type HTTP struct {
	Base

	urlFromInput bool
	textOnly     bool
	getBody      bool
	checkStatus  int
	method       string
	rawData      *template.Template
	headers      map[string]string
	dataPost     map[string]*template.Template

	params map[string]string

	urlTemplate *template.Template
}

func NewHttpFilter(p map[string]string) (Filter, error) {
	f := &HTTP{
		params:       p,
		urlFromInput: false,
		getBody:      true,
		method:       "GET",
		headers:      make(map[string]string),
		dataPost:     make(map[string]*template.Template),
		checkStatus:  200,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["url_from_input"]; ok && v == "true" {
		f.urlFromInput = true
	}
	if v, ok := f.params["text_only"]; ok && v == "true" {
		f.textOnly = true
	}
	if v, ok := f.params["url"]; ok {
		t, err := template.New("httpFilterUrlString").Parse(v)
		if err != nil {
			return nil, err
		}
		f.urlTemplate = t
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
		tmpMap := make(map[string]string)
		err := json.Unmarshal([]byte(v), &tmpMap)
		if err != nil {
			return nil, err
		}
		for i, v := range tmpMap {
			t, err := template.New("httpFilterdataPost" + i).Parse(v)
			if err != nil {
				return nil, err
			}
			f.dataPost[i] = t
		}
	}
	if v, ok := f.params["rawData"]; ok {
		t, err := template.New("httpFilterRawData").Parse(v)
		if err != nil {
			return nil, err
		}
		f.rawData = t
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

func (f *HTTP) DoFilter(msg *data.Message) (bool, error) {
	var req *http.Request
	var err error

	text := msg.GetMessage()

	client := &http.Client{}

	urlString := ""
	if f.urlFromInput {
		urlString = text
	} else {
		urlString, err = msg.ApplyPlaceholder(f.urlTemplate)
		if err != nil {
			return false, err
		}
	}

	var reader io.Reader
	if len(f.dataPost) > 0 {
		values := url.Values{}
		for key, value := range f.dataPost {
			v, err := msg.ApplyPlaceholder(value)
			if err != nil {
				return false, err
			}
			values.Set(key, v)
		}
		reader = bytes.NewBufferString(values.Encode())
	} else if f.rawData != nil {
		body, err := msg.ApplyPlaceholder(f.rawData)
		if err != nil {
			return false, err
		}
		reader = bytes.NewBufferString(body)
	}

	req, err = http.NewRequest(f.method, urlString, reader)
	if err != nil {
		return false, err
	}

	if len(f.headers) > 0 {
		for key, value := range f.headers {
			req.Header.Add(key, value)
		}
	}

	//requestDump, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(requestDump))

	r, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer r.Body.Close()

	ret := false
	log.Debug("[httpFilter] status %s", r.Status)
	if f.checkStatus == r.StatusCode {
		ret = true
		if f.getBody {
			txt := f.getBodyAsString(r)
			if f.textOnly {
				txt = f.TextExtraction(txt)
			}
			msg.SetMessage(txt)
		}
	} else {
		return false, fmt.Errorf("httpFilter received status: %s", r.Status)
	}

	return ret, nil
}

func (f *HTTP) getBodyAsString(r *http.Response) string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func (f *HTTP) TextExtraction(s string) string {
	ret := ""
	domDocTest := html.NewTokenizer(strings.NewReader(s))
	previousStartTokenTest := domDocTest.Token()
loopDomTest:
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = domDocTest.Token()
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" ||
				previousStartTokenTest.Data == "noscript" ||
				previousStartTokenTest.Data == "style" {
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(domDocTest.Text())))
			if len(TxtContent) > 0 {
				ret = fmt.Sprintf("%s %s", ret, TxtContent)
			}
		}
	}

	return ret
}

// Set the name of the filter
func init() {
	register("http", NewHttpFilter)
}
