package filters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	ll "log"
	"os"
	"path/filepath"

	"github.com/evilsocket/islazy/log"
	"github.com/slack-go/slack"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils"
)

// Slack is a Filter to send message, file to a slack channel
type Slack struct {
	Base

	action      string
	target      string
	botToken    string
	blocks      bool
	filename    *template.Template
	downloadURL *template.Template
	body        *template.Template
	to          *template.Template

	params map[string]string
}

// NewSlackFilter is the registered method to instantiate a MailFilter
func NewSlackFilter(p map[string]string) (Filter, error) {
	f := &Slack{
		params:      p,
		action:      "send_message",
		downloadURL: nil,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["text"]; ok {
		t, err := template.New("SlackFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.body = t
	}
	if v, ok := f.params["blocks"]; ok && v == "true" {
		f.blocks = true
	}
	if v, ok := f.params["action"]; ok {
		f.action = v
	}
	if v, ok := f.params["filename"]; ok {
		t, err := template.New("SlackFilenameFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.filename = t
	}
	if v, ok := f.params["url"]; ok {
		t, err := template.New("SlackDwnUrlFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.downloadURL = t
	}
	if v, ok := f.params["target"]; ok {
		f.target = v
	}
	if v, ok := f.params["botToken"]; ok {
		f.botToken = v
	}
	if v, ok := f.params["to"]; ok {
		t, err := template.New("SlackToFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.to = t
	}

	if (f.action == "send_message" || f.action == "send_file") && f.to == nil {
		return nil, fmt.Errorf("destination 'to' is mandatory with this action")
	}
	return f, nil
}

func (f *Slack) sendMessageText(client *slack.Client, dst string, text string) error {
	log.Debug("Slack: send message to %s", dst)
	_, _, err := client.PostMessage(dst, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("sendMessage: slack returned: %s", err)
	}
	return nil
}

func (f *Slack) sendMessageBlocks(client *slack.Client, dst string, jsonBlocks string) error {
	log.Debug("Slack: send message to %s", dst)

	// This is done because the blocks.Unmarshal returns the error:
	// "cannot unmarshal object into Go value of type []json.RawMessage"
	// So we remove the blocks field from the json
	var jsonArray string
	var i interface{}
	if err := json.Unmarshal([]byte(jsonBlocks), &i); err != nil {
		return fmt.Errorf("sendMessageBlocks: unmarshalling: %s", err)
	}
	if m, ok := i.(map[string]interface{}); ok {
		q := m["blocks"]
		output, err := json.Marshal(q)
		if err != nil {
			return fmt.Errorf("sendMessageBlocks: marshalling: %s", err)
		}
		jsonArray = string(output)
	}

	var blocks slack.Blocks
	err := blocks.UnmarshalJSON([]byte(jsonArray))
	if err != nil {
		return fmt.Errorf("sendMessageBlocks: blocks unmarshalling: %s", err)
	}

	_, _, err = client.PostMessage(dst, slack.MsgOptionBlocks(blocks.BlockSet...))
	if err != nil {
		return fmt.Errorf("sendMessage: slack returned: %s", err)
	}
	return nil
}

func (f *Slack) sendFile(client *slack.Client, dst string, filename string) error {
	log.Debug("Slack: send file to %s", dst)
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("sendFile: file '%s': %s", filename, err)
	}
	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("sendFile: file '%s': %s", filename, err)
	}

	params := slack.UploadFileV2Parameters{
		Filename: filepath.Base(filename),
		FileSize: int(fi.Size()),
		Reader:   file,
		Channel:  dst,
	}
	if _, err := client.UploadFileV2(params); err != nil {
		return fmt.Errorf("sendFile: file '%s': %s", filename, err)
	}
	return nil
}

func (f *Slack) sendFileFromBuffer(client *slack.Client, dst string, filename string, buffer []byte) error {
	log.Debug("Slack: send file to %s", dst)
	r := bytes.NewReader(buffer)
	params := slack.UploadFileV2Parameters{
		Filename: filepath.Base(filename),
		FileSize: int(r.Size()),
		Reader:   r,
		Channel:  dst}
	if _, err := client.UploadFileV2(params); err != nil {
		return fmt.Errorf("sendFile: file '%s': %s", filename, err)
	}
	return nil
}

func (f *Slack) downloadFile(client *slack.Client, url string, filename string) (*bytes.Buffer, error) {
	log.Debug("Slack: download file to %s", filename)
	buffer := &bytes.Buffer{}
	err := client.GetFile(url, buffer)
	if err != nil {
		return nil, fmt.Errorf("downloadFile: slack returned: %s", err)
	}
	// write to file
	if filename != "" {
		err = os.WriteFile(filename, buffer.Bytes(), 0644)
		if err != nil {
			return nil, fmt.Errorf("writing file '%s': %s", filename, err)
		}
	}
	return buffer, nil
}

func (f *Slack) getUserInfo(client *slack.Client, user string) (*slack.User, error) {
	return client.GetUserInfo(user)
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Slack) DoFilter(msg *data.Message) (bool, error) {
	var err error
	var dst, token string

	if f.botToken != "" {
		token = f.botToken
	} else {
		t := msg.GetTarget("slackfeeder.botToken")
		if v, ok := t.(string); ok {
			token = v
		} else {
			return false, fmt.Errorf("slack bot botToken not found")
		}
	}

	// check if the botToken is known
	client := slack.New(
		token,
		slack.OptionDebug(false),
		slack.OptionLog(ll.New(os.Stdout, "slackfilter: ", ll.Lshortfile|ll.LstdFlags)),
	)

	switch f.action {
	case "send_message":
		var text string
		if f.to != nil {
			dst, err = msg.ApplyPlaceholder(f.to)
			if err != nil {
				return false, err
			}
		} else {
			return false, fmt.Errorf("destination not specified (param to)")
		}

		if f.body != nil {
			text, err = msg.ApplyPlaceholder(f.body)
			if err != nil {
				return false, err
			}
		} else {
			if f.target == "" {
				f.target = "main"
			}
			if v, ok := msg.GetTarget(f.target).(string); ok {
				text = v
			} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
				text = string(v)
			} else {
				// ERROR this filter can't be used with different types
				return false, fmt.Errorf("received data is not a string")
			}
		}
		if f.blocks {
			err = f.sendMessageBlocks(client, dst, text)
			if err != nil {
				return false, fmt.Errorf("%s", err)
			}
		} else {
			err = f.sendMessageText(client, dst, text)
			if err != nil {
				return false, fmt.Errorf("%s", err)
			}
		}
	case "send_file":
		var filename string
		var buffer []byte
		if f.to != nil {
			dst, err = msg.ApplyPlaceholder(f.to)
			if err != nil {
				return false, err
			}
		} else {
			return false, fmt.Errorf("destination not specified (param to)")
		}
		if f.filename != nil {
			filename, err = msg.ApplyPlaceholder(f.filename)
			if err != nil {
				return false, err
			}
			err = f.sendFile(client, dst, filename)
			if err != nil {
				return false, fmt.Errorf("%s", err)
			}
		} else if f.target != "" {
			if filename == "" {
				return false, fmt.Errorf("filename parameter has to be specified if target is used")
			}
			t := msg.GetTarget(f.target)
			if v, ok := t.([]byte); ok {
				buffer = v
			} else {
				return false, fmt.Errorf("target can be casted to []byte type")
			}
			err = f.sendFileFromBuffer(client, dst, filename, buffer)
			if err != nil {
				return false, fmt.Errorf("%s", err)
			}
		}
	case "download_file":
		var filename, url string

		if f.downloadURL != nil {
			url, err = msg.ApplyPlaceholder(f.downloadURL)
			if err != nil {
				return false, err
			}
		} else {
			target := f.target
			if target == "" {
				target = "urlprivate"
			}
			t := msg.GetTarget(target)
			if v, ok := t.(string); ok {
				url = v
			} else {
				return false, nil
			}
		}
		if f.filename != nil {
			filename, err = msg.ApplyPlaceholder(f.filename)
			if err != nil {
				return false, err
			}
		}
		buffer, err := f.downloadFile(client, url, filename)
		if err != nil {
			return false, err
		}
		msg.SetMessage(buffer.Bytes())
	case "user_info":
		target := f.target
		if target == "" {
			target = "user"
		}
		if v, ok := msg.GetTarget(target).(string); ok {
			u, err := f.getUserInfo(client, v)
			if err != nil {
				return false, fmt.Errorf("get_user_info: %s", err)
			}
			m := utils.FlatStruct(*u)
			for k, v := range m {
				msg.SetTarget("user_"+k, v)
			}
		}

	}
	return true, nil
}

// OnEvent is called when an event occurs
func (f *Slack) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("slack", NewSlackFilter)
}
