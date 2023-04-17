package config

import (
	a "github.com/NYULibraries/aswa/lib/application"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
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
		{"Valid path with valid yaml but missing required fields", "../../testdata/test.yml", "config file is missing required fields"},
		{"Invalid path", "../../config/config_test.yml", "open ../../config/config_test.yml: no such file or directory"},
		{"Invalid path with valid yaml", "./applications.yml", "yaml: unmarshal errors:"},
		{"Empty path", "", "open : no such file or directory"},
		{"Invalid yaml", "./invalid.yml", "yaml: unmarshal errors"},
		{"Valid yaml", configTestPath, ""},
		{"Another valid yaml", "../../testdata/expect_valid.yml", ""},
		{"Invalid yaml", "../../testdata/expect_invalid.yml", "config file is missing required fields"},
		{"Wrong type timeout yaml", "../../testdata/expect_timeout_wrong_type.yml", "yaml: unmarshal errors:\n  line 5: cannot unmarshal !!int `600` into time.Duration"},
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

func TestContainApp(t *testing.T) {
	var tests = []struct {
		description  string
		applications []*a.Application
		appName      string
		expected     bool
	}{
		{"Valid application", []*a.Application{{Name: "test", URL: "http://test.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second, ExpectedLocation: "test"}}, "test", true},
		{"Valid application", []*a.Application{{Name: "test"}}, "test", true},
		{"Invalid application", []*a.Application{{Name: "test"}}, "test2", false},
		{"Invalid application", []*a.Application{{"test", "test", 0, 0, "", ""}}, "test2", false},
		{"Empty application", []*a.Application{}, "test", false},
		{"Empty application", []*a.Application{{Name: "test"}}, "", false},
		{"Empty application", []*a.Application{}, "", false},
	}

	for _, test := range tests {
		t.Run(test.description, testContainAppFunc(test.applications, test.appName))
	}
}

func testContainAppFunc(applications []*a.Application, appName string) func(*testing.T) {
	return func(t *testing.T) {
		for _, app := range applications {
			if app.Name == appName {
				assert.True(t, ContainApp(applications, appName))
			} else {
				assert.False(t, ContainApp(applications, appName))
			}
		}
	}
}

func TestIsConfigAnyRequiredFieldEmpty(t *testing.T) {
	var tests = []struct {
		description string
		application *a.Application
		expected    bool
	}{
		{"Valid application", &a.Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: http.StatusOK, Timeout: 1 * time.Second, ExpectedLocation: "test"}, true},
		{"Valid application", &a.Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second, ExpectedLocation: "test"}, true},
		{"Invalid application", &a.Application{Name: "test"}, false},
		{"Invalid application", &a.Application{Name: "test"}, false},
		{"Invalid application", &a.Application{Name: "test", URL: "http://test.com"}, false},
		{"Valid application", &a.Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Millisecond}, true},
		{"Invalid application", &a.Application{Name: "", URL: "http://test.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second}, false},
		{"Invalid application", &a.Application{Name: "test", URL: "", ExpectedStatusCode: http.StatusOK, Timeout: time.Second, ExpectedLocation: "test"}, false},
		{"Empty application", &a.Application{}, false},
		{"Empty application", &a.Application{Name: "test"}, false},
		{"Empty application", &a.Application{Name: "test", URL: "http://test.com"}, false},
		{"Empty application", &a.Application{Name: "", URL: "", ExpectedStatusCode: http.StatusOK}, false},
		{"Empty application", &a.Application{Name: "test", URL: "http://test.com", Timeout: time.Second}, false},
		{"Empty application", &a.Application{Name: "", URL: "", Timeout: time.Second, ExpectedLocation: "test"}, false},
	}

	for _, test := range tests {
		t.Run(test.description, testIsAnyRequiredFieldFunc(test.application, test.expected))
	}
}

func testIsAnyRequiredFieldFunc(app *a.Application, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, expected, app.Name != "" && app.URL != "" && app.ExpectedStatusCode != 0)
	}
}
