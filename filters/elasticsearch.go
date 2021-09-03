package filters

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/evilsocket/islazy/log"
)

// ElasticSearch is a filter that imports JSON documents in an Elastsic Search database.
type ElasticSearch struct {
	Base

	client   *elasticsearch.Client
	address  string
	username string
	password string
	index    string
	retries  int

	params map[string]string
}

// NewElasticSearchFilter is the registered method to instantiate a ElasticSearchFilter
func NewElasticSearchFilter(p map[string]string) (Filter, error) {
	f := &ElasticSearch{
		params:  p,
		client:  nil,
		retries: 1,
		address: "localhost:9200",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["address"]; ok {
		f.address = v
	}

	if v, ok := f.params["username"]; ok {
		f.username = v
	}

	if v, ok := f.params["password"]; ok {
		f.password = v
	}

	if v, ok := f.params["index"]; ok {
		f.index = v
	}

	if v, ok := f.params["retries"]; ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		f.retries = n
	}

	return f, nil
}

func (f *ElasticSearch) connect() (err error) {
	wait := time.Duration(5) * time.Second

	log.Debug("connecting to %s ...", f.address)

	for attempt := 0; attempt < f.retries; attempt++ {
		cfg := elasticsearch.Config{
			Addresses: []string{f.address},
			Username:  f.username,
			Password:  f.password,
		}

		if f.client, err = elasticsearch.NewClient(cfg); err != nil {
			return fmt.Errorf("error creating ES client: %v", err)
		}

		if res, err := f.client.Info(); err != nil {
			log.Debug("waiting for ES to come up ... [%v] (attempt %d of %d, retrying in %s)", err, attempt+1,
				f.retries, wait)
			time.Sleep(wait)
			continue
		} else {
			defer res.Body.Close()
			if res.IsError() {
				return fmt.Errorf("error getting information from ES cluster: %v", res)
			}
			log.Debug("%+v", res)
			return nil
		}
	}

	return fmt.Errorf("could not connect")
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *ElasticSearch) DoFilter(msg *data.Message) (bool, error) {
	if f.client == nil {
		if err := f.connect(); err != nil {
			return true, err
		}
	}

	rawJSON := msg.GetMessage().(string)
	// make the document id contents dependent so that if we have multiple
	// events for the same object we're not going to create duplicate events
	docID := fmt.Sprintf("%s", sha256.Sum256([]byte(rawJSON)))

	req := esapi.IndexRequest{
		Index:      f.index,
		DocumentID: docID,
		Body:       strings.NewReader(rawJSON),
		Refresh:    "true",
	}
	log.Debug("IndexRequest: %#v", req)

	res, err := req.Do(context.Background(), f.client)
	if err != nil {
		return true, fmt.Errorf("could not save document %s to index: %v", docID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return true, fmt.Errorf("could not index document %s: %v", docID, res)
	}

	msg.SetMessage(docID)

	return true, nil
}

// OnEvent is called when an event occurs
func (f *ElasticSearch) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("elasticsearch", NewElasticSearchFilter)
}
