package cmd

import (
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
			// Setup mock HTTP server to simulate the pag
			mockPromAggregationGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer mockPromAggregationGateway.Close()

			// Set the PROM_AGGREGATION_GATEWAY_URL to the mock server's URL
			t.Setenv(c.EnvPromAggregationGatewayUrl, mockPromAggregationGateway.URL)

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

func TestDoCheck(t *testing.T) {
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
			t.Setenv(c.EnvYamlPath, tt.envYamlPath)
			t.Setenv(c.EnvSlackWebhookUrl, tt.envSlackUrl)
			t.Setenv(c.EnvClusterInfo, tt.envClusterInfo)
			t.Setenv(c.EnvPromAggregationGatewayUrl, mockPushgateway.URL)
			// Set environment variable to true for this test
			t.Setenv(c.EnvSkipWhitelistCheck, "true")
			os.Args = tt.cmdArgs

			// Call function under test
			err := DoCheck()

			// Use assertions to check for expected error
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}
