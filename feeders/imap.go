package feeders

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
)

// Imap is a Feeder that creates a stream from an IMAP server
type Imap struct {
	Base

	host      string
	username  string
	password  string
	mailbox   string
	port      int64
	frequency time.Duration

	ticker            *time.Ticker
	lastCheck         time.Time
	enableAttachments bool
	stopChan          chan bool
}

// NewFolderFeeder is the registered method to instantiate a FolderFeeder
func NewImapFeeder(conf map[string]string) (Feeder, error) {
	f := &Imap{
		stopChan:          make(chan bool),
		frequency:         1 * time.Minute,
		enableAttachments: false,
	}

	if val, ok := conf["imap.host"]; ok {
		f.host = val
	}
	if val, ok := conf["imap.username"]; ok {
		f.username = val
	}
	if val, ok := conf["imap.password"]; ok {
		f.password = val
	}
	if val, ok := conf["imap.mailbox"]; ok {
		f.mailbox = val
	} else {
		f.mailbox = "INBOX"
	}
	if val, ok := conf["imap.port"]; ok {
		i, err := strconv.ParseInt(val, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("port error: %s", err)
		}
		f.port = i
	}
	if val, ok := conf["imap.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", val, err)
		}
		f.frequency = d
	}
	if val, ok := conf["imap.start_from_beginning"]; ok && val == "false" {
		f.lastCheck = time.Now()
	}
	if val, ok := conf["imap.get_attachments"]; ok && val == "true" {
		f.enableAttachments = true
	}

	c, err := f.connect()
	if err != nil {
		return nil, fmt.Errorf("connection error: %s", err)
	}
	defer func() {
		c.Logout()
		c.Close()
	}()

	return f, nil
}

func (f *Imap) connect() (*client.Client, error) {
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", f.host, f.port), nil)
	if err != nil {
		return nil, err
	}

	if err := c.Login(f.username, f.password); err != nil {
		return nil, err
	}

	return c, nil
}

func joinAddresses(addresses []*imap.Address) string {
	list := []string{}
	for _, a := range addresses {
		list = append(list, fmt.Sprintf("<%s> %s", a.Address(), a.PersonalName))
	}

	return strings.Join(list, ",")
}

func (f *Imap) parseMessage(email *imap.Message) error {
	msg := data.NewMessage(email.Envelope.Subject)
	msg.SetExtra("from", joinAddresses(email.Envelope.From))
	msg.SetExtra("to", joinAddresses(email.Envelope.To))
	msg.SetExtra("reply_to", joinAddresses(email.Envelope.ReplyTo))
	msg.SetExtra("in_reply_to", email.Envelope.InReplyTo)
	msg.SetExtra("cc", joinAddresses(email.Envelope.Cc))
	msg.SetExtra("bcc", joinAddresses(email.Envelope.Bcc))
	msg.SetExtra("sender", joinAddresses(email.Envelope.Sender))
	msg.SetExtra("message_id", email.Envelope.MessageId)
	msg.SetExtra("subject", email.Envelope.Subject)
	msg.SetExtra("date", email.Envelope.Date.UTC().Format(time.RFC3339))
	msg.SetExtra("is_attachment", "false")

	var section imap.BodySectionName
	r := email.GetBody(&section)
	if r == nil {
		log.Fatal("Server didn't returned message body")
	}

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		return fmt.Errorf("parse email: %s", err)
	}
	defer mr.Close()

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("parsing email: %s", err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
			msg.SetExtra("body", string(b))
		case *mail.AttachmentHeader:
			if f.enableAttachments {
				filename, _ := h.Filename()
				b, _ := io.ReadAll(p.Body)
				clonedMsg := msg.Clone()
				clonedMsg.SetExtra("is_attachment", "true")
				clonedMsg.SetExtra("attachment_filename", filename)
				clonedMsg.SetExtra("attachment_body", b)
				f.Propagate(clonedMsg)
			}
		}
	}

	f.Propagate(msg)
	return nil
}

func (f *Imap) fetchMessages() error {
	client, err := f.connect()
	if err != nil {
		return fmt.Errorf("imap connection: %s", err)
	}
	defer client.Close()

	mailbox, err := client.Select(f.mailbox, false)
	if err != nil {
		return fmt.Errorf("imap box select: %s", err)
	}

	// Define the range of emails to fetch
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mailbox.Messages)

	// Fetch the required message attributes
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}
	done := make(chan error, 1)
	messages := make(chan *imap.Message, 20)
	go func() {
		done <- client.Fetch(seqSet, items, messages)
	}()

	for email := range messages {
		if email != nil && f.lastCheck.Before(email.Envelope.Date) {
			f.parseMessage(email)
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("fetching: %s", err)
	}

	return nil
}

// Start propagates a message every time a new fs event happens in the folder
func (f *Imap) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				if err := f.fetchMessages(); err != nil {
					log.Error("%s: %s", f.Name(), err)
				}
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Imap) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Imap) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("imap", NewImapFeeder)
}
