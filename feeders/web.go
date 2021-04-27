package feeders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils"

	"github.com/evilsocket/islazy/log"
)

// Web is a Feeder that creates a stream from an URL
type Web struct {
	Base

	url         string
	frequency   time.Duration
	textOnly    bool
	checkStatus int
	method      string
	cookieFile  string
	rawData     string
	headers     map[string]string
	dataPost    map[string]string

	params  map[string]string
	cookies []*http.Cookie

	stopChan    chan bool
	ticker      *time.Ticker
	lastParsing time.Time
}

// NewWebFeeder is the registered method to instantiate a WebFeeder
func NewWebFeeder(conf map[string]string) (Feeder, error) {
	f := &Web{
		params:      conf,
		checkStatus: 0,
		stopChan:    make(chan bool),
		dataPost: make(map[string]string),
		frequency:   60 * time.Second,
		method:      "GET",
		lastParsing: time.Time{},
	}

	if val, ok := f.params["web.url"]; ok {
		f.url = val
	}
	if val, ok := f.params["web.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", val, err)
		}
		f.frequency = d
	}
	if v, ok := f.params["web.text_only"]; ok && v == "true" {
		f.textOnly = true
	}
	if v, ok := f.params["web.method"]; ok {
		f.method = v
	}
	if v, ok := f.params["web.headers"]; ok {
		err := json.Unmarshal([]byte(v), &f.headers)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := f.params["web.data"]; ok {
		err := json.Unmarshal([]byte(v), &f.dataPost)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := f.params["web.rawData"]; ok {
		f.rawData = v
	}
	if v, ok := f.params["web.status"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		f.checkStatus = i
	}
	if v, ok := f.params["web.cookies"]; ok {
		f.cookieFile = v
		cookies, err := utils.ParseCookieFile(v)
		if err != nil {
			return nil, err
		}
		f.cookies = cookies
	}

	return f, nil
}

func (f *Web) prepareRequest() (*http.Request, error) {
	var req *http.Request
	var err error

	var reader io.Reader
	if len(f.dataPost) > 0 {
		values := url.Values{}
		for key, value := range f.dataPost {
			values.Set(key, value)
		}
		reader = bytes.NewBufferString(values.Encode())
	} else if f.rawData != "" {
		reader = bytes.NewBufferString(f.rawData)
	}

	req, err = http.NewRequest(f.method, f.url, reader)
	if err != nil {
		return nil, err
	}

	if len(f.headers) > 0 {
		for key, value := range f.headers {
			req.Header.Add(key, value)
		}
	}

	if len(f.cookies) > 0 {
		for _, c := range f.cookies {
			req.AddCookie(c)
		}
	}
	return req, nil
}

func (f *Web) getBodyAsString(r *http.Response) string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func (f *Web) parseURL() error {
	var txt string
	extra := make(map[string]interface{})

	log.Debug("Start Web parsing: %s", f.url)
	req, err := f.prepareRequest()
	if err != nil {
		return err
	}

	//requestDump, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(requestDump))

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	//responseDump, err := httputil.DumpResponse(r, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(responseDump))

	txt = f.getBodyAsString(r)
	meta := utils.GetMetaFromHTML(txt)
	extra["url"] = f.url
	extra["title"] = meta.Title
	extra["description"] = meta.Description
	extra["image"] = meta.Image
	extra["sitename"] = meta.SiteName

	log.Debug("status -> %s", r.Status)
	if f.checkStatus == 0 || f.checkStatus == r.StatusCode {
		if f.textOnly {
			txt = utils.ExtractTextFromHTML(txt)
		}
	} else {
		return fmt.Errorf("unexpected status: %s", r.Status)
	}

	msg := data.NewMessageWithExtra(txt, extra)
	f.Propagate(msg)

	f.lastParsing = time.Now()
	log.Debug("Finished at %s...updating date", f.lastParsing.Format("2006-01-02 15:04:05"))

	return nil
}

// Start propagates a message every time the URL is read
func (f *Web) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		// first start!
		_ = f.parseURL()

		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				_ = f.parseURL()
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Web) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.ticker.Stop()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Web) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("web", NewWebFeeder)
}
