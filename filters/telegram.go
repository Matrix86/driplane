package filters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/evilsocket/islazy/log"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/Matrix86/driplane/data"
)

const dateLayout = "2006-01-02_15-04-05"

// Telegram is a Filter to send message, file to Telegram
type Telegram struct {
	Base

	action       string
	downloadPath *template.Template
	toChat       *template.Template
	to           *template.Template
	message      *template.Template

	params map[string]string
}

// NewTelegramFilter is the registered method to instantiate a MailFilter
func NewTelegramFilter(p map[string]string) (Filter, error) {
	f := &Telegram{
		params: p,
		action: "send_message",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["action"]; ok {
		f.action = v
	}
	if v, ok := f.params["to"]; ok {
		t, err := template.New("TelegramToUserFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.to = t
	}
	if v, ok := f.params["to_chatid"]; ok {
		t, err := template.New("TelegramToChatIDFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.toChat = t
	}
	if v, ok := f.params["filename"]; ok {
		t, err := template.New("TelegramDowonloadFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.downloadPath = t
	}
	if v, ok := f.params["text"]; ok {
		t, err := template.New("TelegramMessageFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.message = t
	}

	if f.action == "download_file" && f.downloadPath == nil {
		return nil, fmt.Errorf("param 'filename' is mandatory with this action")
	}

	if f.action == "send_message" {
		if f.message == nil {
			return nil, fmt.Errorf("param 'text' is mandatory with this action")
		}
		if f.to == nil && f.toChat == nil {
			return nil, fmt.Errorf("param 'to' or 'to_chatid' are mandatory with this action")
		}
	}

	if f.action != "download_file" && f.action != "send_message" {
		return nil, fmt.Errorf("action '%s' is not valid", f.action)
	}

	return f, nil
}

func (f *Telegram) getDocFilename(doc *tg.Document) string {
	var filename, ext string
	for _, attr := range doc.Attributes {
		switch v := attr.(type) {
		case *tg.DocumentAttributeImageSize:
			switch doc.MimeType {
			case "image/png":
				ext = ".png"
			case "image/webp":
				ext = ".webp"
			case "image/tiff":
				ext = ".tif"
			default:
				ext = ".jpg"
			}
		case *tg.DocumentAttributeAnimated:
			ext = ".gif"
		case *tg.DocumentAttributeSticker:
			ext = ".webp"
		case *tg.DocumentAttributeVideo:
			switch doc.MimeType {
			case "video/mpeg":
				ext = ".mpeg"
			case "video/webm":
				ext = ".webm"
			case "video/ogg":
				ext = ".ogg"
			default:
				ext = ".mp4"
			}
		case *tg.DocumentAttributeAudio:
			switch doc.MimeType {
			case "audio/webm":
				ext = ".webm"
			case "audio/aac":
				ext = ".aac"
			case "audio/ogg":
				ext = ".ogg"
			default:
				ext = ".mp3"
			}
		case *tg.DocumentAttributeFilename:
			filename = v.FileName
		}
	}

	if filename == "" {
		filename = fmt.Sprintf(
			"doc%d_%s%s", doc.GetID(),
			time.Unix(int64(doc.Date), 0).Format(dateLayout),
			ext,
		)
	}

	return filename
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Telegram) DoFilter(msg *data.Message) (bool, error) {
	var tgClient *tg.Client

	target := msg.GetTarget("_telegram_api")
	if target == nil {
		return false, nil
	}
	if api, ok := target.(*tg.Client); ok {
		tgClient = api
	}

	if f.action == "download_file" {
		if mediaTarget := msg.GetTarget("_msg_media"); mediaTarget != nil {
			log.Info("message media found...")
			switch media := mediaTarget.(type) {
			case *tg.MessageMediaDocument:
				if docClass, ok := media.GetDocument(); ok {
					if d, ok := docClass.AsNotEmpty(); ok {
						filename := f.getDocFilename(d)
						loc := d.AsInputDocumentFileLocation()
						d := downloader.NewDownloader()

						// adding the filename to the message so this can be used as placeholder
						msg.SetExtra("msg_filename", filename)

						downloadPath, err := msg.ApplyPlaceholder(f.downloadPath)
						if err != nil {
							return false, err
						}

						log.Debug("downloading image file to : %s", downloadPath)

						folder := filepath.Dir(downloadPath)
						if folder != "" {
							if _, err := os.Stat(folder); os.IsNotExist(err) {
								log.Debug("folder %s doesn't exist...creating it", folder)
								os.MkdirAll(folder, 0700)
							}
						}

						if _, err := d.Download(tgClient, loc).ToPath(context.Background(), downloadPath); err == nil {
							log.Debug("telegramFilter: document downloaded to %s", downloadPath)
							return true, nil
						} else {
							log.Error("telegramFilter: couldn't download document to %s: %s", downloadPath, err)
						}
					}
				}

			case *tg.MessageMediaPhoto:
				if photoClass, ok := media.GetPhoto(); ok {
					if photo, ok := photoClass.AsNotEmpty(); ok {
						var thumbSize string
						maxH := 0
						maxW := 0

						// Get the biggest picture
						for _, g := range photo.Sizes {
							if sz, ok := g.(*tg.PhotoSize); ok {
								if sz.GetH() > maxH || sz.GetW() > maxW {
									thumbSize = sz.GetType()
									maxH = sz.GetH()
									maxW = sz.GetW()
								}
							}
						}

						filename := fmt.Sprintf(
							"photo%d_%s.jpg", photo.GetID(),
							time.Unix(int64(photo.Date), 0).Format(dateLayout),
						)

						loc := &tg.InputPhotoFileLocation{
							ID:            photo.ID,
							AccessHash:    photo.AccessHash,
							FileReference: photo.FileReference,
							ThumbSize:     thumbSize,
						}

						// adding the filename to the message so this can be used as placeholder
						msg.SetExtra("msg_filename", filename)
						downloadPath, err := msg.ApplyPlaceholder(f.downloadPath)
						if err != nil {
							return false, err
						}

						log.Debug("downloading image file to : %s", downloadPath)

						folder := filepath.Dir(downloadPath)
						if folder != "" {
							if _, err := os.Stat(folder); os.IsNotExist(err) {
								log.Debug("folder %s doesn't exist...creating it", folder)
								os.MkdirAll(folder, os.ModeDir)
							}
						}

						d := downloader.NewDownloader()
						if _, err := d.Download(tgClient, loc).ToPath(context.Background(), downloadPath); err == nil {
							log.Debug("telegramFilter: image downloaded to %s", downloadPath)
							return true, nil
						} else {
							log.Error("telegramFilter: couldn't download image to %s: %s", downloadPath, err)
						}
					}
				}
			}
		}
	} else if f.action == "send_message" {
		txt, err := msg.ApplyPlaceholder(f.message)
		if err != nil {
			return false, err
		}

		sender := message.NewSender(tgClient)

		if f.to != nil {
			to, err := msg.ApplyPlaceholder(f.to)
			if err != nil {
				return false, err
			}

			target := sender.Resolve(to)
			if _, err := target.Text(context.Background(), txt); err != nil {
				return false, err
			}
			return true, nil
		}
		if f.toChat != nil {
			sGroupID, err := msg.ApplyPlaceholder(f.toChat)
			if err != nil {
				return false, err
			}

			groupID, err := strconv.ParseInt(sGroupID, 10, 64)
			if err != nil {
				return false, err
			}

			peer := &tg.InputPeerChat{
				ChatID: groupID, // Replace with the ID of the group you want to send the message to
			}
			target := sender.To(peer)
			if _, err := target.Text(context.Background(), txt); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	// if we arrive here let's stop the propagation
	return false, nil
}

// OnEvent is called when an event occurs
func (f *Telegram) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("telegram", NewTelegramFilter)
}
