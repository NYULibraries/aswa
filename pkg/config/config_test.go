package config

import (
	"net/http"
	"testing"
	"time"

	a "github.com/NYULibraries/aswa/pkg/application"
	"github.com/stretchr/testify/assert"
)

const configTestPath = "../../config/prod.applications.yml"

func TestNewConfig(t *testing.T) {
	var tests = []struct {
		description string
		path        string
		expectedErr string
	}{
		{"Valid prod config", configTestPath, ""},
		{"Valid dev config", "../../config/dev.applications.yml", ""},
		{"Valid saas config", "../../config/saas.applications.yml", ""},
		{"Valid primo_ve config", "../../config/primo_ve.applications.yml", ""},
		{"Valid testdata config", "../../testdata/expect_valid.yml", ""},
		{"Missing required fields", "../../testdata/expect_invalid.yml", "config file is missing one or more required fields"},
		{"Wrong type for timeout", "../../testdata/expect_timeout_wrong_type.yml", "cannot unmarshal !!int `600` into time.Duration"},
		{"Nonexistent file in config dir", "../../config/does_not_exist.yml", "no such file or directory"},
		{"Nonexistent config.yml", "../../config/config.yml", "no such file or directory"},
		{"Nonexistent testdata file", "../../testdata/test.yml", "no such file or directory"},
		{"Nonexistent relative file", "./prod.applications.yml", "no such file or directory"},
		{"Empty path", "", "no such file or directory"},
	}

	for _, test := range tests {
		t.Run(test.description, testNewConfigFunc(test.path, test.expectedErr))
	}
}

func testNewConfigFunc(path string, expectedErr string) func(*testing.T) {
	return func(t *testing.T) {
		// Set environment variable to true for this test
		t.Setenv(EnvSkipWhitelistCheck, "true")
		_, err := NewConfig(path)

		if expectedErr == "" {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, expectedErr)
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
		{"Invalid application", []*a.Application{{Name: "test", URL: "test", ExpectedStatusCode: 0, Timeout: 0, ExpectedLocation: "", ExpectedContent: "", ExpectedCSP: ""}}, "test2", false},
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
