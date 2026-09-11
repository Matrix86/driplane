package feeders

import (
	"fmt"
	"io"
	"net"
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
	timeout   time.Duration
	batchSize uint32

	ticker             *time.Ticker
	startFromBeginning bool
	initialized        bool
	uidValidity        uint32
	lastUID            uint32
	enableAttachments  bool
	stopChan           chan bool
}

// NewFolderFeeder is the registered method to instantiate a FolderFeeder
func NewImapFeeder(conf map[string]string) (Feeder, error) {
	f := &Imap{
		stopChan:           make(chan bool, 1),
		frequency:          1 * time.Minute,
		timeout:            2 * time.Minute,
		batchSize:          50,
		startFromBeginning: true,
		enableAttachments:  false,
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
	if val, ok := conf["imap.timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified timeout cannot be parsed '%s': %s", val, err)
		}
		f.timeout = d
	}
	if val, ok := conf["imap.batch_size"]; ok {
		i, err := strconv.ParseUint(val, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("batch_size error: %s", err)
		}
		if i == 0 {
			return nil, fmt.Errorf("batch_size cannot be 0")
		}
		f.batchSize = uint32(i)
	}
	if val, ok := conf["imap.start_from_beginning"]; ok && val == "false" {
		f.startFromBeginning = false
	}
	if val, ok := conf["imap.get_attachments"]; ok && val == "true" {
		f.enableAttachments = true
	}

	c, err := f.connect()
	if err != nil {
		return nil, fmt.Errorf("connection error: %s", err)
	}
	f.disconnect(c)

	return f, nil
}

func (f *Imap) connect() (*client.Client, error) {
	dialer := &net.Dialer{Timeout: f.timeout}
	c, err := client.DialWithDialerTLS(dialer, fmt.Sprintf("%s:%d", f.host, f.port), nil)
	if err != nil {
		return nil, err
	}
	// the deadline is applied to every IMAP command: without it a stalled server
	// would block the feeder forever
	c.Timeout = f.timeout

	if err := c.Login(f.username, f.password); err != nil {
		f.disconnect(c)
		return nil, err
	}

	return c, nil
}

// disconnect terminates the session: LOGOUT makes the server send a BYE, which
// is what actually closes the underlying connection. Note that client.Close()
// only issues the IMAP CLOSE command (it deselects the mailbox), so relying on
// it leaks one connection and one goroutine per call until the server refuses
// new sessions with "imap: connection closed".
func (f *Imap) disconnect(c *client.Client) {
	if err := c.Logout(); err != nil {
		log.Debug("%s: logout: %s", f.Name(), err)
	}
}

func joinAddresses(addresses []*imap.Address) string {
	list := []string{}
	for _, a := range addresses {
		list = append(list, fmt.Sprintf("<%s> %s", a.Address(), a.PersonalName))
	}

	return strings.Join(list, ",")
}

func (f *Imap) parseMessage(email *imap.Message, section *imap.BodySectionName) error {
	if email.Envelope == nil {
		return fmt.Errorf("server didn't return the envelope of the message %d", email.Uid)
	}

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

	r := email.GetBody(section)
	if r == nil {
		return fmt.Errorf("server didn't return the body of the message %d", email.Uid)
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
	c, err := f.connect()
	if err != nil {
		return fmt.Errorf("imap connection: %s", err)
	}
	defer f.disconnect(c)

	// read-only: fetching the body must not flag the emails as \Seen
	mailbox, err := c.Select(f.mailbox, true)
	if err != nil {
		return fmt.Errorf("imap box select: %s", err)
	}

	// the UIDs are valid only as long as UIDVALIDITY doesn't change
	if !f.initialized || f.uidValidity != mailbox.UidValidity {
		f.uidValidity = mailbox.UidValidity
		f.lastUID = 0
		if !f.startFromBeginning && mailbox.UidNext > 0 {
			f.lastUID = mailbox.UidNext - 1
		}
		f.initialized = true
	}

	// if the server doesn't advertise UIDNEXT we cannot split the request in
	// batches: ask for everything following the last seen UID in one shot
	if mailbox.UidNext == 0 {
		maxUID, err := f.fetchRange(c, f.lastUID+1, 0)
		if maxUID > f.lastUID {
			f.lastUID = maxUID
		}
		return err
	}

	// UidNext is the UID that will be assigned to the next email: nothing to do
	// if we already saw everything. Asking for 'lastUID+1:*' in that case would
	// return the last email again (RFC 3501 section 6.4.8).
	if mailbox.UidNext <= f.lastUID+1 {
		return nil
	}

	// the emails are fetched in batches: a single FETCH of the whole mailbox
	// keeps every body in memory and can easily exceed the limits the server
	// enforces on a session
	for f.lastUID+1 < mailbox.UidNext {
		from := f.lastUID + 1
		to := mailbox.UidNext - 1
		if to-from >= f.batchSize {
			to = from + f.batchSize - 1
		}

		if _, err := f.fetchRange(c, from, to); err != nil {
			return err
		}
		// the server can leave holes in the UID sequence, so move forward even
		// when the batch returned nothing
		f.lastUID = to
	}

	return nil
}

// fetchRange downloads the emails in the [from, to] UID range and propagates
// them, returning the highest UID it has seen. A 'to' of 0 means '*'.
func (f *Imap) fetchRange(c *client.Client, from, to uint32) (uint32, error) {
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	// BODY.PEEK[] instead of BODY[]: the latter would set the \Seen flag
	section := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchUid}
	done := make(chan error, 1)
	messages := make(chan *imap.Message, 10)
	go func() {
		done <- c.UidFetch(seqSet, items, messages)
	}()

	fetched := 0
	maxUID := uint32(0)
	for email := range messages {
		// the server answers with the last email of the mailbox if it has no UID
		// in the requested range, so the boundaries are checked here too
		if email == nil || email.Uid < from || (to != 0 && email.Uid > to) {
			continue
		}
		if email.Uid > maxUID {
			maxUID = email.Uid
		}
		fetched++
		if err := f.parseMessage(email, section); err != nil {
			log.Error("%s: %s", f.Name(), err)
		}
	}

	if err := <-done; err != nil {
		return maxUID, fmt.Errorf("fetching: %s", err)
	}
	log.Debug("%s: fetched %d emails in the UID range %d:%d", f.Name(), fetched, from, to)

	return maxUID, nil
}

// Start propagates a message every time a new fs event happens in the folder
func (f *Imap) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				f.ticker.Stop()
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
	// non blocking: a fetch could be in progress
	select {
	case f.stopChan <- true:
	default:
	}
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Imap) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("imap", NewImapFeeder)
}
