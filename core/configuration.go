package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	sync.RWMutex

	FilePath string
	flat     map[string]string
}

func LoadConfiguration(path string) (*Configuration, error) {
	configuration := &Configuration{
		FilePath: path,
	}

	file, err := os.Open(path)
	if err != nil {
		return configuration, fmt.Errorf("loading configuration: file opening: %s", err)
	}

	bytes, _ := ioutil.ReadAll(file)

	var cc map[interface{}]interface{}

	if err := yaml.Unmarshal(bytes, &cc); err != nil {
		return configuration, fmt.Errorf("loading configuration: %s", err)
	}
	configuration.flat = configuration.flatMap(cc)

	return configuration, nil
}

func (c *Configuration) flatMap(m map[interface{}]interface{}) map[string]string {
	flatten := make(map[string]string)
	for k, v := range m {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			mv := c.flatMap(v.(map[interface{}]interface{}))
			for kk, vv := range mv {
				key := fmt.Sprintf("%s.%s", k, kk)
				flatten[key] = vv
			}

		case reflect.Array, reflect.Slice:
			for i, vv := range v.([]interface{}) {
				mv := c.flatMap(vv.(map[interface{}]interface{}))
				for kk, vv := range mv {
					key := fmt.Sprintf("%s.%s%d.%s", k, k, i, kk)
					flatten[key] = vv
				}
			}

		default:
			flatten[k.(string)] = fmt.Sprint(v)
		}
	}
	return flatten
}

func (c *Configuration) Get(name string) string {
	c.RLock()
	defer c.RUnlock()
	if _, ok := c.flat[name]; !ok {
		return ""
	}

	return c.flat[name]
}

func (c *Configuration) Set(name string, value string) error {
	c.Lock()
	defer c.Unlock()

	c.flat[name] = value
	return nil
}

func (c *Configuration) GetConfig() map[string]string {
	c.RLock()
	defer c.RUnlock()
	return c.flat
}
