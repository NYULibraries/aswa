package application

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			_, _ = fmt.Fprint(w, "<html><body><h1>Herman Melville</h1></body></html>")
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/timeout":
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/wrongstatus":
			w.WriteHeader(http.StatusNotFound)
		case "/500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/slowresponse":
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockServer.Close()

	var tests = []struct {
		description              string
		application              *Application
		expectedSuccess          bool
		expectedActualStatusCode int
		expectedActualLocation   string
		expectedContentSuccess   bool
		expectedActualContent    string
	}{
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusFound, 800 * time.Millisecond, "", ""}, false, http.StatusMovedPermanently, "", true, ""},
		{"Success: correct error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusNotFound, 600 * time.Millisecond, "", ""}, true, http.StatusNotFound, "", true, ""},
		{"Success: success status code expected", &Application{"", "https://library.nyu.edu", http.StatusOK, 800 * time.Millisecond, "", ""}, true, http.StatusOK, "", true, ""},
		{"Failure: wrong status code expected", &Application{"", mockServer.URL + "/wrongstatus", http.StatusOK, 800 * time.Millisecond, "", ""}, false, http.StatusNotFound, "", true, ""},
		{"Failure: application is down", &Application{"", mockServer.URL + "/500", http.StatusOK, 800 * time.Millisecond, "", ""}, false, http.StatusInternalServerError, "", true, ""},
		{"Success: timeout", &Application{"", "https://library.nyu.edu", http.StatusOK, 200 * time.Millisecond, "", ""}, true, http.StatusOK, "", true, ""},
		{"Failure: timeout", &Application{"", mockServer.URL + "/slowresponse", http.StatusOK, 1 * time.Millisecond, "", ""}, false, 0, "", false, ""},
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect location expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusFound, 800 * time.Millisecond, "", ""}, false, http.StatusNotFound, "", true, ""},
		{"Success: expected content found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Herman Melville"}, true, http.StatusOK, "", true, "Herman Melville"},
		{"Failure: expected content not found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Jules Verne - 20,000 Leagues Under the Sea"}, true, http.StatusOK, "", false, ""},
	}

	for _, test := range tests {
		t.Run(test.description, testGetStatusFunc(test.application, test.expectedSuccess, test.expectedActualStatusCode, test.expectedContentSuccess, test.expectedActualContent))
	}
}

func testGetStatusFunc(application *Application, expectedSuccess bool, expectedActualStatusCode int, expectedContentSuccess bool, actualContent string) func(*testing.T) {
	return func(t *testing.T) {
		status := application.GetStatus()
		assert.Equal(t, expectedSuccess, status.StatusOk)
		assert.Equal(t, expectedActualStatusCode, status.ActualStatusCode)
		assert.Equal(t, expectedContentSuccess, status.StatusContentOk)
		if expectedContentSuccess {
			assert.Contains(t, status.ActualContent, actualContent)
		}
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		description    string
		appStatus      *ApplicationStatus
		expectedOutput string
	}{
		{description: "Successful status", appStatus: &ApplicationStatus{Application: &Application{"", "https://library.nyu.edu", http.StatusOK, time.Second, "", ""}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200}, expectedOutput: "Success: URL https://library.nyu.edu resolved with 200"},
		{description: "Failed status", appStatus: &ApplicationStatus{Application: &Application{URL: "https://library.nyu.edu", ExpectedStatusCode: http.StatusOK, Timeout: time.Second}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 404}, expectedOutput: "Failure: URL https://library.nyu.edu resolved with 404, expected 200, "},
		{description: "Successful status with location", appStatus: &ApplicationStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "https://library.nyu.edu/"}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Success: URL http://library.nyu.edu resolved with 301, redirect location matched https://library.nyu.edu/"},
		{description: "Failed status with location", appStatus: &ApplicationStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "http://library.nyu.edu/"}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Failure: URL http://library.nyu.edu resolved with 301, but redirect location https://library.nyu.edu/ did not match http://library.nyu.edu/"},
		{description: "Successful status with expected content", appStatus: &ApplicationStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Example Domain"}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Success: URL https://example.com resolved with 200\nSuccess: ExpectedContent Example Domain matched ActualContent Example Domain"},
		{description: "Failed status with unexpected content", appStatus: &ApplicationStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Wrong Content"}, StatusOk: false, StatusContentOk: false, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Failure: URL https://example.com resolved with 200, expected 200, \nFailure: Expected content Wrong Content did not match ActualContent Example Domain"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, test.appStatus.String())
		})
	}
}
