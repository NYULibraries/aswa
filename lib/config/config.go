package config

import (
	"errors"
	a "github.com/NYULibraries/aswa/lib/application"
	"gopkg.in/yaml.v3"
	"os"
)

// Config struct to replace environment variables
type Config struct {
	Applications map[string][]*a.Application
}

// Check if any required App field is empty
func hasEmptyRequiredFields(app *a.Application) bool {
	return app.Name == "" || app.URL == "" || app.ExpectedStatusCode == 0
}

// Loop through all categories and applications, check if any required field is empty
func (list *Config) isConfigAnyRequiredFieldEmpty() bool {
	for _, apps := range list.Applications {
		for _, app := range apps {
			if hasEmptyRequiredFields((*a.Application)(app)) {
				return true
			}
		}
	}
	return false
}

func loadConfig(yamlPath string) (*Config, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	if config.isConfigAnyRequiredFieldEmpty() {
		return nil, errors.New("config file is missing one or more required fields: name, url, expected_status code")
	}

	return &config, nil
}

func ContainApp(applications []*a.Application, e string) bool {
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
