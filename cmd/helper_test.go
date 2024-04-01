package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
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

func TestGetPushgatewayUrl(t *testing.T) {
	tests := []struct {
		name              string
		envPushgatewayUrl string
		want              string
	}{
		{"envPushgatewayUrl is set", "test-pushgateway-url", "test-pushgateway-url"},
		{"envPushgatewayUrl is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv(envPushgatewayUrl, tt.envPushgatewayUrl)

			got := getPushgatewayUrl()

			assert.Equal(t, tt.want, got, "getPushgatewayUrl() should return correct pushgateway URL")

			os.Unsetenv(envPushgatewayUrl)
		})
	}

}

func TestPushMetrics(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  func(w http.ResponseWriter, r *http.Request)
		expectedError bool
	}{
		{
			name: "Successful push",
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedError: false,
		},
		{
			name: "Failed push with non-200 response",
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server for this test case
			server := httptest.NewServer(http.HandlerFunc(tt.mockResponse))
			defer server.Close()

			// Setup environment variable to mock the pushgateway URL
			os.Setenv(envPushgatewayUrl, server.URL)
			defer os.Unsetenv(envPushgatewayUrl)

			// Increment a test counter to simulate metrics that would be pushed
			incrementFailedTestsCounter("testApp")

			// Call PushMetrics to attempt to push the test counter to the mock server
			err := PushMetrics()

			// Assert that an error occurred only if one was expected
			if tt.expectedError {
				if err == nil {
					t.Errorf("TestPushMetrics %s: expected an error but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("TestPushMetrics %s: did not expect an error but got: %v", tt.name, err)
				}
			}
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
			// Setup mock HTTP server to simulate the pushgateway
			mockPushgateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer mockPushgateway.Close()

			// Set the PUSHGATEWAY_URL to the mock server's URL
			os.Setenv(envPushgatewayUrl, mockPushgateway.URL)
			defer os.Unsetenv(envPushgatewayUrl)

			// Convert MockApplications to real ones
			var appData []*a.Application
			for _, app := range tt.apps {
				appData = append(appData, toApplication(app))
			}

			// Mock the network calls (assuming there is a method to mock network calls for app.GetStatus)
			// mockNetworkCalls(appData)

			// Call the RunSyntheticTests function
			err := RunSyntheticTests(appData, tt.targetAppName)

			// Handle the assertions based on expected errors
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tt.wantErrMessage, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckDo(t *testing.T) {
	// Define a mock server to simulate the Pushgateway
	mockPushgateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // simulate a successful push to the Pushgateway
	}))
	defer mockPushgateway.Close()

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
		{"valid case with existing app", "../testdata/expect_valid.yml", "https://hooks.slack.com/test-url", "test-cluster", []string{"cmd", "specialcollections"}, false},
		{"missing yaml path", "", "https://hooks.slack.com/test-url", "test-cluster", []string{"cmd", "arg"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up environment variables and command line arguments
			os.Setenv(envYamlPath, tt.envYamlPath)
			os.Setenv(envSlackWebhookUrl, tt.envSlackUrl)
			os.Setenv(envClusterInfo, tt.envClusterInfo)
			os.Setenv(envPushgatewayUrl, mockPushgateway.URL)
			// Set environment variable to true for this test
			os.Setenv(c.EnvSkipWhitelistCheck, "true")
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
			os.Unsetenv(envPushgatewayUrl)
			os.Unsetenv(c.EnvSkipWhitelistCheck)
		})
	}
}
