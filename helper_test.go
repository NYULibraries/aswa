package main

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	a "github.com/NYULibraries/aswa/lib/application"
	"github.com/stretchr/testify/assert"
)

type MockApplication struct {
	Name               string
	URL                string
	ExpectedStatusCode int
	Timeout            time.Duration
	ExpectedLocation   string
	Status             *MockApplicationStatus
}

type MockApplicationStatus struct {
	StatusOk        bool
	StatusContentOk bool
}

func (m *MockApplication) GetStatus() *MockApplicationStatus {
	return m.Status
}

func TestGetYamlPath(t *testing.T) {
	tests := []struct {
		name        string
		envYamlPath string
		want        string
	}{
		{"envYamlPath is set", "test-path", "test-path"},
		{"envYamlPath is not set", "", "config/dev.applications.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv(envYamlPath, tt.envYamlPath)

			got := getYamlPath(log.New(os.Stdout, "", 0))

			assert.Equal(t, tt.want, got, "getYamlPath() should return correct yaml path")

			os.Unsetenv(envYamlPath)
		})
	}
}

func TestGetCmdArg(t *testing.T) {
	tests := []struct {
		name   string
		osArgs []string
		want   string
	}{
		{"No argument", []string{"test"}, ""},
		{"With argument", []string{"test", "arg1"}, "arg1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Args = tt.osArgs

			got := getCmdArg()

			assert.Equal(t, tt.want, got, "getCmdArg() should return correct command argument")
		})
	}
}

func TestGetSlackWebhookUrl(t *testing.T) {
	tests := []struct {
		name               string
		envSlackWebhookUrl string
		want               string
	}{
		{"envSlackWebhookUrl is set", "https://hooks.slack.com/test-url", "https://hooks.slack.com/test-url"},
		{"envSlackWebhookUrl is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv(envSlackWebhookUrl, tt.envSlackWebhookUrl)

			got := getSlackWebhookUrl()

			assert.Equal(t, tt.want, got, "getSlackWebhookUrl() should return correct Slack webhook URL")

			os.Unsetenv(envSlackWebhookUrl)
		})
	}
}

func TestGetClusterInfo(t *testing.T) {
	tests := []struct {
		name           string
		envClusterInfo string
		want           string
	}{
		{"envClusterInfo is set", "test-cluster", "test-cluster"},
		{"envClusterInfo is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up environment variable
			os.Setenv(envClusterInfo, tt.envClusterInfo)

			// Call function under test
			got := getClusterInfo()

			// Assert that the function returns the expected result
			assert.Equal(t, tt.want, got, "getClusterInfo() should return correct cluster info")

			// Unset environment variable for next test
			os.Unsetenv(envClusterInfo)
		})
	}
}

func TestRunSyntheticTests(t *testing.T) {
	// Convert MockApplication to a.Application
	toApplication := func(ma *MockApplication) *a.Application {
		return &a.Application{
			Name:               ma.Name,
			URL:                ma.URL,
			ExpectedStatusCode: ma.ExpectedStatusCode,
			Timeout:            ma.Timeout,
			ExpectedLocation:   ma.ExpectedLocation,
		}
	}

	tests := []struct {
		name           string
		targetAppName  string
		apps           []*MockApplication
		wantErr        bool
		wantErrMessage string
	}{
		{
			"Valid synthetic test run",
			"test",
			[]*MockApplication{
				{
					Name:               "test",
					URL:                "test",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "test",
					Status:             &MockApplicationStatus{StatusOk: true, StatusContentOk: true},
				},
			},
			false,
			"",
		},
		{
			"Synthetic test run with nonexistent app",
			"nonexistent",
			[]*MockApplication{
				{
					Name:               "test",
					URL:                "test",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "test",
					Status:             &MockApplicationStatus{StatusOk: true, StatusContentOk: true},
				},
			},
			true,
			"app 'nonexistent' not found in config file",
		},
		{
			"Synthetic test run with failing app status",
			"test",
			[]*MockApplication{
				{
					Name:               "test",
					URL:                "test",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "test",
					Status:             &MockApplicationStatus{StatusOk: false, StatusContentOk: true},
				},
			},
			false,
			"",
		},
		{
			"Synthetic test run with failing app content status",
			"test",
			[]*MockApplication{
				{
					Name:               "test",
					URL:                "test",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "test",
					Status:             &MockApplicationStatus{StatusOk: true, StatusContentOk: false},
				},
			},
			false,
			"",
		},
		{
			"Synthetic test run with multiple failing apps",
			"",
			[]*MockApplication{
				{
					Name:               "app1",
					URL:                "app1",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "app1",
					Status:             &MockApplicationStatus{StatusOk: false, StatusContentOk: false},
				},
				{
					Name:               "app2",
					URL:                "app2",
					ExpectedStatusCode: http.StatusOK,
					Timeout:            1 * time.Second,
					ExpectedLocation:   "app2",
					Status:             &MockApplicationStatus{StatusOk: false, StatusContentOk: false},
				},
			},
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appData []*a.Application
			for _, app := range tt.apps {
				appData = append(appData, toApplication(app))
			}

			err := RunSyntheticTests(appData, tt.targetAppName)

			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
				assert.Equal(t, tt.wantErrMessage, err.Error(), "Expected error message does not match")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}

func TestCheckDo(t *testing.T) {
	tests := []struct {
		name           string
		envYamlPath    string
		envSlackUrl    string
		envClusterInfo string
		cmdArgs        []string
		wantErr        bool
	}{
		// Add test cases here
		{"valid case, but missing app", "test-path", "https://hooks.slack.com/test-url", "test-cluster", []string{"cmd", "arg"}, true},
		{"valid case with existing app", "testdata/expect_valid.yml", "https://hooks.slack.com/test-url", "test-cluster", []string{"cmd", "specialcollections"}, false},
		{"missing yaml path", "", "https://hooks.slack.com/test-url", "test-cluster", []string{"cmd", "arg"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up environment variables and command line arguments
			os.Setenv(envYamlPath, tt.envYamlPath)
			os.Setenv(envSlackWebhookUrl, tt.envSlackUrl)
			os.Setenv(envClusterInfo, tt.envClusterInfo)
			os.Args = tt.cmdArgs

			// Initialize Check struct with a logger that outputs to stdout.
			logger := log.New(os.Stdout, "", 0)
			ch := &Check{Logger: logger}

			// Call function under test
			err := ch.Do()

			// Use assertions to check for expected error
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}

			// Unset environment variables
			os.Unsetenv(envYamlPath)
			os.Unsetenv(envSlackWebhookUrl)
			os.Unsetenv(envClusterInfo)
		})
	}
}
