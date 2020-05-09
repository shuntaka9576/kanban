package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const GITHUB_URL = "github.com"

type Config interface {
	AuthToken() (string, error)
}

type cliConfig struct {
	config *configEntry
}

func (c *cliConfig) AuthToken() (string, error) {
	if c.config == nil || c.config.Token != "" {
		err := c.initConfigEntry()
		if err != nil {
			return "", fmt.Errorf("init conifg error: %v", err)
		}
		return c.config.Token, nil
	}
	return c.config.Token, nil
}

func (c *cliConfig) initConfigEntry() error {
	config, err := parseConfigFile()
	if err != nil {
		return err
	}
	c.config = config

	return nil
}

func New() Config {
	return &cliConfig{}
}

type configEntry struct {
	User  string
	Token string `yaml:"oauth_token"`
}

func configFile() string {
	dir, _ := homedir.Expand("~/.config/kanban")
	return dir
}

func parseConfigFile() (*configEntry, error) {
	f, err := os.Open(configFile())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteArray, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var config yaml.Node
	err = yaml.Unmarshal(byteArray, &config)
	if err != nil {
		return nil, err
	}

	for _, topNodeList := range config.Content {
		for i := 0; i < len(topNodeList.Content); i++ {
			if topNodeList.Content[i].Value == GITHUB_URL {
				var entries []configEntry
				topNodeList.Content[i+1].Decode(&entries)
				return &entries[0], nil
			}
		}
	}

	return nil, fmt.Errorf("could not find config entry for %q", GITHUB_URL)
}
