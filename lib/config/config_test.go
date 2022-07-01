package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		applications []*Application
		appName      string
		expected     bool
	}{
		{"Valid application", []*Application{{Name: "test", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: time.Second, ExpectedLocation: "test"}}, "test", true},
		{"Valid application", []*Application{{Name: "test"}}, "test", true},
		{"Invalid application", []*Application{{Name: "test"}}, "test2", false},
		{"Invalid application", []*Application{{"test", "test", 0, 0, ""}}, "test2", false},
		{"Empty application", []*Application{}, "test", false},
		{"Empty application", []*Application{{Name: "test"}}, "", false},
		{"Empty application", []*Application{}, "", false},
	}

	for _, test := range tests {
		t.Run(test.description, testContainAppFunc(test.applications, test.appName))
	}
}

func testContainAppFunc(applications []*Application, appName string) func(*testing.T) {
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

func TestExtractValuesFromConfig(t *testing.T) {
	var tests = []struct {
		description            string
		application            *Application
		appName                string
		ExpectedName           string
		ExpectedURL            string
		ExpectedStatusCode     int
		ExpectedTimeout        time.Duration
		ExpectedActualLocation string
	}{
		{"Valid application", &Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: time.Second, ExpectedLocation: "test"}, "test", "test", "http://test.com", 200, time.Second, "test"},
		{"Valid application", &Application{Name: "test", URL: "http://test1.com", ExpectedStatusCode: 200, Timeout: time.Second, ExpectedLocation: "test"}, "test", "test", "http://test1.com", 200, time.Second, "test"},
		{"Empty application", &Application{}, "", "", "", 0, 0, ""},
		{"Empty application", &Application{Name: "test"}, "", "test", "", 0, 0, ""},
		{"Empty application", &Application{Name: "test", URL: "http://test.com"}, "test", "test", "http://test.com", 0, 0, ""},
	}

	for _, test := range tests {
		t.Run(test.description, testExtractValuesFromConfigFunc(test.application))
	}
}

func testExtractValuesFromConfigFunc(app *Application) func(*testing.T) {
	return func(t *testing.T) {
		expectedName, expectedURL, expectedStatusCode, expectedTimeout, expectedActualLocation := ExtractValuesFromConfig(app)
		assert.Equal(t, expectedName, app.Name)
		assert.Equal(t, expectedURL, app.URL)
		assert.Equal(t, expectedStatusCode, app.ExpectedStatusCode)
		assert.Equal(t, expectedTimeout, app.Timeout)
		assert.Equal(t, expectedActualLocation, app.ExpectedLocation)
	}
}

func TestAnyRequiredField(t *testing.T) {
	var tests = []struct {
		description string
		application *Application
		expected    bool
	}{
		{"Valid application", &Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: 1 * time.Second, ExpectedLocation: "test"}, true},
		{"Valid application", &Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: time.Second, ExpectedLocation: "test"}, true},
		{"Invalid application", &Application{Name: "test"}, false},
		{"Invalid application", &Application{Name: "test"}, false},
		{"Invalid application", &Application{Name: "test", URL: "http://test.com"}, false},
		{"Valid application", &Application{Name: "test", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: time.Millisecond}, true},
		{"Invalid application", &Application{Name: "", URL: "http://test.com", ExpectedStatusCode: 200, Timeout: time.Second}, false},
		{"Invalid application", &Application{Name: "test", URL: "", ExpectedStatusCode: 200, Timeout: time.Second, ExpectedLocation: "test"}, false},
		{"Empty application", &Application{}, false},
		{"Empty application", &Application{Name: "test"}, false},
		{"Empty application", &Application{Name: "test", URL: "http://test.com"}, false},
		{"Empty application", &Application{Name: "", URL: "", ExpectedStatusCode: 200}, false},
		{"Empty application", &Application{Name: "test", URL: "http://test.com", Timeout: time.Second}, false},
		{"Empty application", &Application{Name: "", URL: "", Timeout: time.Second, ExpectedLocation: "test"}, false},
	}

	for _, test := range tests {
		t.Run(test.description, testAnyRequiredFieldFunc(test.application, test.expected))
	}
}

func testAnyRequiredFieldFunc(app *Application, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, expected, app.Name != "" && app.URL != "" && app.ExpectedStatusCode != 0)
	}
}
