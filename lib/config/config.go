package config

import (
	"errors"
	"io/ioutil"
	"time"

	a "github.com/NYULibraries/aswa/lib/application"
	"gopkg.in/yaml.v3"
)

//Config struct to replace environment variables
type Config struct {
	Applications []*Application
}

type Application a.Application

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
		return nil, err
	}
	if config.anyRequiredEmpty() {
		return nil, errors.New("config file is missing required fields")
	}

	return &config, nil
}

func ExtractValuesFromConfig(app *Application) (name string, url string, expectedStatusCode int, timeout time.Duration, expectedActualLocation string) {
	name = app.Name
	url = app.URL
	expectedStatusCode = app.ExpectedStatusCode
	timeout = app.Timeout
	expectedActualLocation = app.ExpectedLocation
	return
}

func ContainApp(applications []*Application, e string) bool {
	for _, application := range applications {
		if application.Name == e {
			return true
		}
	}
	return false
}

func NewConfig(yamlPath string) (*Config, error) {
	return loadConfig(yamlPath)
}
