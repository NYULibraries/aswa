package metrics

import (
	dto "github.com/prometheus/client_model/go"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/NYULibraries/aswa/pkg/config"
)

func TestIncrementFailedTestsCounter(t *testing.T) {
	failedTests.Reset()
	appInfo.Reset()
	t.Setenv(c.EnvName, "saas")

	SetAppInfo("testApp", "https://example.com/app")
	IncrementFailedTestsCounter("testApp")

	counterMetric := &dto.Metric{}
	if err := failedTests.WithLabelValues("saas", "testApp").Write(counterMetric); err != nil {
		t.Fatalf("failed to read counter metric: %v", err)
	}

	if counterMetric.GetCounter().GetValue() != 1 {
		t.Fatalf("expected counter value 1, got %v", counterMetric.GetCounter().GetValue())
	}

	infoMetric := &dto.Metric{}
	if err := appInfo.WithLabelValues("saas", "testApp", "https://example.com/app").Write(infoMetric); err != nil {
		t.Fatalf("failed to read app info metric: %v", err)
	}

	if infoMetric.GetGauge().GetValue() != 1 {
		t.Fatalf("expected app info gauge value 1, got %v", infoMetric.GetGauge().GetValue())
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
			failedTests.Reset()
			appInfo.Reset()

			// Setup mock HTTP server for this test case
			server := httptest.NewServer(http.HandlerFunc(tt.mockResponse))
			defer server.Close()

			// Setup environment variable to mock the pushgateway URL
			t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)
			t.Setenv(c.EnvName, "saas")

			// Record app metadata and increment a test counter to simulate metrics that would be pushed.
			SetAppInfo("testApp", "https://example.com/app")
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
