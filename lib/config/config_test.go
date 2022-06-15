package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const configTestPath = "../../config/applications.yml"

func TestNewConfig(t *testing.T) {
	var tests = []struct {
		description string
		path        string
		expectedErr string
	}{
		{"Valid path", configTestPath, ""},
		{"Valid path with valid yaml", configTestPath, ""},
		{"Valid path with invalid yaml", "../../config/config.yml", "yaml: unmarshal errors:"},
		{"Valid path with valid yaml but missing required fields", "../../testdata/config.yml", "config file is missing required fields"},
		{"Invalid path", "../../config/config_test.yml", "open ../../config/config_test.yml: no such file or directory"},
		{"Invalid path with valid yaml", "./applications.yml", "yaml: unmarshal errors:"},
		{"Empty path", "", "open : no such file or directory"},
		{"Invalid yaml", "./invalid.yml", "yaml: unmarshal errors"},
	}

	for _, test := range tests {
		t.Run(test.description, testNewConfigFunc(test.path, test.expectedErr))
	}
}

func testNewConfigFunc(path string, expectedErr string) func(*testing.T) {
	return func(t *testing.T) {
		_, err := NewConfig(path)
		if expectedErr == "" {
			assert.Nil(t, err)
		} else {
			assert.Error(t, err, expectedErr)
		}
	}
}
