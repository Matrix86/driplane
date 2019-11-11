package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/evilsocket/islazy/log"
)

type Configuration struct {
	config map[string]string

	LogPath  string `json:"log_path"`
	LogLevel string `json:"log_level"`
}

func flatMap(m map[string]interface{}) map[string]string {
	flatten := make(map[string]string)
	for k, v := range m {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			mv := flatMap(v.(map[string]interface{}))
			for kk, vv := range mv {
				key := fmt.Sprintf("%s.%s", k, kk)
				flatten[key] = vv
			}

		case reflect.Array, reflect.Slice:
			for i, vv := range v.([]interface{}) {
				mv := flatMap(vv.(map[string]interface{}))
				for kk, vv := range mv {
					key := fmt.Sprintf("%s.%s%d.%s", k, k, i, kk)
					flatten[key] = vv
				}
			}

		default:
			flatten[k] = fmt.Sprint(v)
		}
	}
	return flatten
}

func LoadConfiguration(path string) (Configuration, error) {
	configuration := Configuration{}

	file, err := os.Open(path)
	if err != nil {
		return configuration, fmt.Errorf("LoadConfiguration: file opening: %s", err)
	}

	bytes, _ := ioutil.ReadAll(file)

	var c map[string]interface{}
	if err := json.Unmarshal([]byte(bytes), &c); err != nil {
		return configuration, fmt.Errorf("Unmarshal problem: %s", err)
	}

	configuration.config = flatMap(c)

	if val, ok := configuration.config["log_level"]; ok {
		configuration.LogLevel = val
	} else {
		configuration.LogLevel = "info"
	}

	if val, ok := configuration.config["log_path"]; ok {
		configuration.LogPath = val
	}

	return configuration, nil
}

func (c *Configuration) Get(name string) (interface{}, error) {
	if _, ok := c.config[name]; !ok {
		return nil, fmt.Errorf("configuration named '%s' doesn't exist", name)
	}

	return c.config[name], nil
}

func (c *Configuration) Set(name string, value string) error {
	c.config[name] = value

	return nil
}

func (c *Configuration) GetConfig() map[string]string {
	return c.config
}

func (c *Configuration) GetLogLevel() log.Verbosity {
	switch c.LogLevel {
	case "debug":
		return log.DEBUG
	case "info":
		return log.INFO
	case "error":
		return log.ERROR

	default:
		return log.INFO
	}
}
