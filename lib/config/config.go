package config

import (
	"errors"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)


//Config struct to replace environment variables
type Config struct {
	Applications [] *Application
}

type Application struct {
	Name 			   string `yaml:"name"`
	URL                string `yaml:"url"`
	ExpectedStatusCode int    `yaml:"expected_status"`
	Timeout            time.Duration `default:"1 * time.Minute"`
	ExpectedLocation   string `yaml:"expected_location"`
}

// Check if any required App field is empty
func (app *Application) anyRequiredField() bool {
	return app.Name == "" || app.URL == "" || app.ExpectedStatusCode == 0 
}

// Loop through all applications and check if any required field is empty
func (list *Config) anyRequiredEmpty() bool {
	for _, app := range list.Applications {
		if app.anyRequiredField() {
			return true
		}
	}
	return false
}

func loadConfig(yamlPath string) (*Config, error) {
	data, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic (err)
	}

	if config.anyRequiredEmpty() {
		return nil, errors.New("config file is missing required fields")
	}

	return &config, nil
}

func NewConfig(yamlPath string) (*Config, error) {
	return loadConfig(yamlPath)
}

