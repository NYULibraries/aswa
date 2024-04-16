package metrics

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetPromAggregationGatewayUrl(t *testing.T) {
	tests := []struct {
		name                         string
		envPromAggregationGatewayUrl string
		want                         string
	}{
		{"envPromAggregationGatewayUrl is set", "test-promaggregationgateway-url", "test-promaggregationgateway-url"},
		{"envPromAggregationGatewayUrl is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv(EnvPromAggregationGatewayUrl, tt.envPromAggregationGatewayUrl)

			got := getPromAggregationgatewayUrl()

			assert.Equal(t, tt.want, got, "getPushgatewayUrl() should return correct pushgateway URL")

			os.Unsetenv(EnvPromAggregationGatewayUrl)
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
			os.Setenv(EnvPromAggregationGatewayUrl, server.URL)
			defer os.Unsetenv(EnvPromAggregationGatewayUrl)

			// Increment a test counter to simulate metrics that would be pushed
			IncrementFailedTestsCounter("testApp")

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