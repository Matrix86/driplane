package feeders

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils/apt"

	"github.com/evilsocket/islazy/log"
)

// Apt is a Feeder that creates a stream from a Apt feed
type Apt struct {
	Base

	url           string
	distribution  string
	indexURL      string
	packageType   string
	frequency     time.Duration
	ignorePubDate bool

	stopChan chan bool
	ticker   *time.Ticker
}

// NewAptFeeder is the registered method to instantiate a AptFeeder
func NewAptFeeder(conf map[string]string) (Feeder, error) {
	f := &Apt{
		stopChan:      make(chan bool),
		frequency:     60 * time.Second,
		ignorePubDate: false,
	}

	if val, ok := conf["apt.url"]; ok {
		f.url = val
	}
	if val, ok := conf["apt.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", val, err)
		}
		f.frequency = d
	}
	if val, ok := conf["apt.dist"]; ok {
		f.distribution = val
	}
	if val, ok := conf["apt.index"]; ok {
		f.indexURL = val
	}

	return f, nil
}

func getExtraFromPackage(item apt.BinaryPackage) map[string]interface{} {
	extra := make(map[string]interface{})
	elems := reflect.ValueOf(item).Elem()
	typeOfT := elems.Type()
	for i := 0; i < elems.NumField(); i++ {
		f := elems.Field(i)
		if f.Type().String() == "string" {
			extra[strings.ToLower(typeOfT.Field(i).Name)] = f.Interface().(string)
		} else if f.Type().String() == "[]string" {
			slice := f.Interface().([]string)
			quoted := make([]string, len(slice))
			for x, v := range slice {
				quoted[x] = fmt.Sprintf("'%s'", v)
			}
			extra[strings.ToLower(typeOfT.Field(i).Name)] = strings.Join(quoted, ",")
		} else if f.Type().String() == "int" {
			extra[strings.ToLower(typeOfT.Field(i).Name)] = fmt.Sprintf("%d", f.Interface().(int))
		}
	}
	return extra
}

func (f *Apt) parseFeed() error {
	if f.indexURL != "" {
		// using directly the path
		repo := apt.Repository{}
		repo.ForceIndexURL(f.indexURL)
		packages, err := repo.GetPackages()
		if err != nil {
			return err
		}
		for _, item := range packages {
			extra := getExtraFromPackage(item)
			main := ""
			if item.Filename != "" {
				main = item.Filename
			}
			msg := data.NewMessageWithExtra(main, extra)
			f.Propagate(msg)
		}
	} else {

	}
	return nil
}

// Start propagates a message every time a new row is published
func (f *Apt) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		// first start!
		_ = f.parseFeed()

		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				_ = f.parseFeed()
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Apt) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.ticker.Stop()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Apt) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("apt", NewAptFeeder)
}
