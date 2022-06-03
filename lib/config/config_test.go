package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const configTestPath = "./applications.yml"

func TestNewConfig(t *testing.T) {
	var tests = []struct {
		description string
		path        string
		expectedErr string
	}{
		{"Valid path", configTestPath, ""},
		{"Invalid path", "../../config/config_test.yml", "open ../../config/config_test.yml: no such file or directory"},
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
