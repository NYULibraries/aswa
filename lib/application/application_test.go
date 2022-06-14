package application

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	var tests = []struct {
		description string
		application *Application
		//url string
		//expectedStatusCode  int
		//timeout time.Duration
		expectedSuccess          bool
		expectedActualStatusCode int
		expectedActualLocation   string
	}{
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 600 * time.Millisecond, ""}, true, http.StatusMovedPermanently, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusFound, 600 * time.Millisecond, ""}, false, http.StatusMovedPermanently, ""},
		{"Success: correct error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusNotFound, 600 * time.Millisecond, ""}, true, http.StatusNotFound, ""},
		{"Success: success status code expected", &Application{"", "https://library.nyu.edu", http.StatusOK, 600 * time.Millisecond, ""}, true, http.StatusOK, ""},
		{"Failure: wrong status code expected", &Application{"", "https://httpstat.us/404", http.StatusOK, 600 * time.Millisecond, ""}, false, http.StatusNotFound, ""},
		{"Failure: application is down", &Application{"", "https://httpstat.us/500", http.StatusOK, 800 * time.Millisecond, ""}, false, http.StatusInternalServerError, ""},
		{"Success: timeout", &Application{"", "https://library.nyu.edu", http.StatusOK, 200 * time.Millisecond, ""}, true, http.StatusOK, ""},
		{"Failure: timeout", &Application{"", "httpstat.us/200?sleep=100", http.StatusOK, 1 * time.Millisecond, ""}, false, 0, ""},
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/"}, true, http.StatusMovedPermanently, "https://library.nyu.edu/"},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 600 * time.Millisecond, "http://library.nyu.edu/"}, false, http.StatusMovedPermanently, "https://library.nyu.edu/"},
		{"Failure: wrong redirect location expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 600 * time.Millisecond, "http://library.nyu.edu/"}, false, http.StatusMovedPermanently, "https://library.nyu.edu/"},
		{"Failure: wrong error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusFound, 800 * time.Millisecond, ""}, false, http.StatusNotFound, ""},
	}

	for _, test := range tests {
		t.Run(test.description, testGetStatusFunc(test.application, test.expectedSuccess, test.expectedActualStatusCode))
	}
}

func testGetStatusFunc(application *Application, expectedSuccess bool, expectedActualStatusCode int) func(*testing.T) {
	return func(t *testing.T) {
		status := application.GetStatus()
		assert.Equal(t, expectedSuccess, status.Success)
		assert.Equal(t, expectedActualStatusCode, status.ActualStatusCode)
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		description    string
		appStatus      *ApplicationStatus
		expectedOutput string
	}{
		{"Successful status", &ApplicationStatus{&Application{"", "https://library.nyu.edu", http.StatusOK, time.Second, ""}, true, http.StatusOK, ""}, "Success: URL https://library.nyu.edu resolved with 200"},
		{"Failed status", &ApplicationStatus{&Application{"", "https://library.nyu.edu", http.StatusOK, time.Second, ""}, false, http.StatusNotFound, ""}, "Failure: URL https://library.nyu.edu resolved with 404, expected 200, "},
		{"Successful status with location", &ApplicationStatus{&Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, time.Second, "https://library.nyu.edu/"}, true, http.StatusMovedPermanently, "https://library.nyu.edu/"}, "Success: URL http://library.nyu.edu resolved with 301, redirect location matched https://library.nyu.edu/"},
		{"Failed status with location", &ApplicationStatus{&Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, time.Second, "http://library.nyu.edu/"}, false, http.StatusMovedPermanently, "https://library.nyu.edu/"}, "Failure: URL http://library.nyu.edu resolved with 301, but redirect location https://library.nyu.edu/ did not match http://library.nyu.edu/"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, test.appStatus.String())
		})
	}
}
