package filters

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"

	gomail "gopkg.in/gomail.v2"
)

// Mail is a Filter to send e-mail using the input Message
type Mail struct {
	Base

	template *template.Template
	username string
	password string
	server   string
	fromAddr string
	fromName string
	port     int
	to       []string
	subject  string
	useAuth  bool
	params   map[string]string
}

// NewMailFilter is the registered method to instantiate a MailFilter
func NewMailFilter(p map[string]string) (Filter, error) {
	f := &Mail{
		params:  p,
		useAuth: false,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["body"]; ok {
		t, err := template.New("MailFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.template = t
	}
	if v, ok := f.params["username"]; ok {
		f.username = v
	}
	if v, ok := f.params["password"]; ok {
		f.password = v
	}
	if v, ok := f.params["host"]; ok {
		f.server = v
	}
	if v, ok := f.params["port"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		f.port = i
	}
	if v, ok := f.params["fromAddr"]; ok {
		f.fromAddr = v
	}
	if v, ok := f.params["fromName"]; ok {
		f.fromName = v
	}
	if v, ok := f.params["to"]; ok {
		f.to = strings.Split(v, ",")
	}
	if v, ok := f.params["subject"]; ok {
		f.subject = v
	}
	if v, ok := f.params["use_auth"]; ok && v == "true" {
		f.useAuth = true
	}
	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Mail) DoFilter(msg *data.Message) (bool, error) {
	var err error
	var text string

	if v, ok := msg.GetMessage().(string); ok {
		text = v
	} else if v, ok := msg.GetMessage().([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	if f.template != nil {
		text, err = msg.ApplyPlaceholder(f.template)
		if err != nil {
			return false, err
		}
	}

	var d gomail.Dialer
	if f.useAuth {
		d = gomail.Dialer{Host: f.server, Port: f.port, Username: f.username, Password: f.password}
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		d = gomail.Dialer{Host: f.server, Port: f.port}
	}

	s, err := d.Dial()
	if err != nil {
		return false, err
	}

	m := gomail.NewMessage()
	for _, n := range f.to {
		m.SetAddressHeader("From", f.fromAddr, f.fromName)
		m.SetHeader("To", n)
		m.SetHeader("Subject", f.subject)
		m.SetBody("text/html", text)

		if err := gomail.Send(s, m); err != nil {
			log.Error("could not send email to %s: %s", n, err)
		}
		m.Reset()
	}

	return true, nil
}

// Set the name of the filter
func init() {
	register("mail", NewMailFilter)
}
