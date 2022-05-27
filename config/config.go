package config

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

//Config stores configuration data on applications
type Config struct {
	Applications *ResourceList
}

//ResourceList stores data on applications
type ResourceList struct {
	resources map[string]*resourceConfig
}

//resourceConfig stores data on an application
type resourceConfig struct {
	Name                     string `yaml:"name"`
	URL                      string `yaml:"url"`
	ExpectedStatus           int    `yaml:"expected_status"`
	ExpectedRedirectLocation string `yaml:"expected_redirect_location"`
}

type configList struct {
	Applications []*resourceConfig `yaml:"applications"`
}

//NewConfig returns a Config pointer loaded from a yaml file
func NewConfig(yamlPath string) (*Config, error) {
	list := &configList{}
	err := list.loadConfig(yamlPath)
	if err != nil {
		return nil, err
	}

	applications := &ResourceList{}
	err = applications.loadConfig(list.Applications)
	if err != nil {
		return nil, err
	}

	return &Config{applications}, nil
}

func (r *ResourceList) loadConfig(list []*resourceConfig) error {
	missingConfig := false
	r.resources = make(map[string]*resourceConfig)
	for _, resource := range list {
		r.resources[resource.Name] = resource
		if resource.anyRequiredEmpty() {
			missingConfig = true
		}
	}

	if missingConfig {
		return errors.New("incomplete configuration")
	}

	return nil
}

func (c *resourceConfig) anyRequiredEmpty() bool {
	return c.Name == "" || c.URL == "" || c.ExpectedStatus == 0
}

// loadConfig returns a list of applications structs from a yaml file
func (list *configList) loadConfig(yamlPath string) error {
	yamlData, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlData, list)

	return err
}
