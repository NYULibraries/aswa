package main

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
	}{
		{"Success: correct redirect expected", &Application{"http://library.nyu.edu", http.StatusMovedPermanently, 200 * time.Millisecond}, true, http.StatusMovedPermanently},
		{"Failure: wrong redirect expected", &Application{"http://library.nyu.edu", http.StatusFound, 200 * time.Millisecond}, false, http.StatusMovedPermanently},
		{"Success: correct error expected", &Application{"https://library.nyu.edu/nopageexistshere", http.StatusNotFound, 200 * time.Millisecond}, true, http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.description, testGetStatusFunc(test.application, test.expectedSuccess, test.expectedActualStatusCode))
	}
}

func testGetStatusFunc(application *Application, expectedSuccess bool, expectedActualStatusCode int) func(*testing.T) {
	return func(t *testing.T) {
		status := application.GetStatus()
		assert.Equal(t, status.Success, expectedSuccess)
		assert.Equal(t, status.ActualStatusCode, expectedActualStatusCode)
	}
}
