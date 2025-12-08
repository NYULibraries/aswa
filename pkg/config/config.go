package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	a "github.com/NYULibraries/aswa/pkg/application"
	"gopkg.in/yaml.v3"
)

// Define whitelist using a map with empty struct as values.
// Simulate a set using a map with empty structs as values, which take up zero bytes. This way, the lookup is both quick and memory-efficient
var allowedConfigPaths = map[string]struct{}{
	"config/dev.applications.yml":      {},
	"config/primo_ve.applications.yml": {},
	"config/prod.applications.yml":     {},
	"config/saas.applications.yml":     {},
}

const EnvSkipWhitelistCheck = "SKIP_WHITELIST_CHECK"

// Config struct to replace environment variables
type Config struct {
	Applications []*a.Application
}

// Check if any required App field is empty
func hasEmptyRequiredFields(app *a.Application) bool {
	return app.Name == "" || app.URL == "" || app.ExpectedStatusCode == 0
}

// Loop through all categories and applications, check if any required field is empty
func (list *Config) isConfigAnyRequiredFieldEmpty() bool {
	for _, app := range list.Applications {
		if hasEmptyRequiredFields(app) {
			return true
		}
	}
	return false
}

func loadConfig(yamlPath string) (*Config, error) {
	skipCheck, _ := strconv.ParseBool(os.Getenv(EnvSkipWhitelistCheck))
	if !skipCheck {
		if _, ok := allowedConfigPaths[filepath.Clean(yamlPath)]; !ok {
			return nil, errors.New("config file path is not allowed")
		}
	}
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
