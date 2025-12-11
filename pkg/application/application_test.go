package application

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatus(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the CSP header to the response
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

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
		case "/redir":
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
		case "/redir-wrong":
			w.Header().Set("Location", "/somewhere-else")
			w.WriteHeader(http.StatusFound)
		case "/login":
			_, _ = fmt.Fprint(w, "Log into your account")
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
		expectedCSPSuccess       bool
		expectedActualCSP        string
	}{
		{
			description: "Success: relative redirect with content follow",
			application: &Application{
				Name:               "primo-ve-search",
				URL:                mockServer.URL + "/redir",
				ExpectedStatusCode: http.StatusFound,
				Timeout:            2 * time.Second,
				ExpectedLocation:   "/login",
				ExpectedContent:    "Log into your account",
			},
			expectedSuccess:          true,
			expectedActualStatusCode: http.StatusFound,
			expectedActualLocation:   "/login",
			expectedContentSuccess:   true,
			expectedActualContent:    "Log into your account",
			expectedCSPSuccess:       true,
			expectedActualCSP:        "",
		},
		{
			description: "Failure: relative redirect mismatches expected",
			application: &Application{
				Name:               "primo-ve-search",
				URL:                mockServer.URL + "/redir-wrong",
				ExpectedStatusCode: http.StatusFound,
				Timeout:            2 * time.Second,
				ExpectedLocation:   "/login",
				ExpectedContent:    "Log into your account",
			},
			expectedSuccess:          false,
			expectedActualStatusCode: http.StatusFound,
			expectedActualLocation:   "/somewhere-else",
			expectedContentSuccess:   false,
			expectedActualContent:    "",
			expectedCSPSuccess:       true,
			expectedActualCSP:        "",
		},
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", "", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, "", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusFound, 800 * time.Millisecond, "", "", ""}, false, http.StatusMovedPermanently, "", true, "", true, ""},
		{"Success: 301 redirect with dynamic location (no expected_location)", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "", "", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, "", true, ""},
		{"Success: correct error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusNotFound, 600 * time.Millisecond, "", "", ""}, true, http.StatusNotFound, "", true, "", true, ""},
		{"Success: success status code expected", &Application{"", "https://library.nyu.edu", http.StatusOK, 800 * time.Millisecond, "", "", ""}, true, http.StatusOK, "", true, "", true, ""},
		{"Failure: wrong status code expected", &Application{"", mockServer.URL + "/wrongstatus", http.StatusOK, 800 * time.Millisecond, "", "", ""}, false, http.StatusNotFound, "", true, "", true, ""},
		{"Failure: application is down", &Application{"", mockServer.URL + "/500", http.StatusOK, 800 * time.Millisecond, "", "", ""}, false, http.StatusInternalServerError, "", true, "", true, ""},
		{"Success: timeout", &Application{"", "https://library.nyu.edu", http.StatusOK, 200 * time.Millisecond, "", "", ""}, true, http.StatusOK, "", true, "", true, ""},
		{"Failure: timeout", &Application{"", mockServer.URL + "/slowresponse", http.StatusOK, 1 * time.Millisecond, "", "", ""}, false, 0, "", false, "", false, ""},
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", "", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, "", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", "", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, "", true, ""},
		{"Failure: wrong redirect location expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", "", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, "", true, ""},
		{"Failure: wrong error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusFound, 800 * time.Millisecond, "", "", ""}, false, http.StatusNotFound, "", true, "", true, ""},
		{"Success: expected content found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Herman Melville", ""}, true, http.StatusOK, "", true, "Herman Melville", true, ""},
		{"Failure: expected content not found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Jules Verne - 20,000 Leagues Under the Sea", ""}, true, http.StatusOK, "", false, "", true, ""},
		{"Success: expected CSP header found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "", "default-src 'self'"}, true, http.StatusOK, "", true, "", true, "default-src 'self'"},
		{"Failure: expected CSP header not found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "", "default-src 'none'"}, true, http.StatusOK, "", true, "", false, ""},
	}

	for _, test := range tests {
		t.Run(test.description, testGetStatusFunc(test.application, test.expectedSuccess, test.expectedActualStatusCode, test.expectedContentSuccess, test.expectedActualContent, test.expectedCSPSuccess, test.expectedActualCSP))
	}
}

func testGetStatusFunc(application *Application, expectedSuccess bool, expectedActualStatusCode int, expectedContentSuccess bool, actualContent string, expectedCSPSuccess bool, actualCSP string) func(*testing.T) {
	return func(t *testing.T) {
		status := application.GetStatus()
		assert.Equal(t, expectedSuccess, status.StatusOk)
		assert.Equal(t, expectedActualStatusCode, status.ActualStatusCode)
		assert.Equal(t, expectedContentSuccess, status.StatusContentOk)
		assert.Equal(t, expectedCSPSuccess, status.StatusCSPOk)
		if expectedContentSuccess {
			assert.Contains(t, status.ActualContent, actualContent)
		}
		if expectedCSPSuccess {
			assert.Contains(t, status.ActualCSP, actualCSP)
		}
	}
}

// This is a custom type that implements the http.RoundTripper interface: https://pkg.go.dev/net/http#RoundTripper
// It will be used to simulate a network error for your tests.
type errorTransport struct{}

// The RoundTrip method is what's called by http.Client when it makes a request.
// By always returning an error here, we can simulate network issues.
func (t *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}

func TestCreateApplicationStatus(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/successful":
			if r.Method == http.MethodGet {
				_, _ = fmt.Fprint(w, "Successful Request")
			}
		case "/failed":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer mockServer.Close()

	// Define an application for testing
	app := Application{
		URL:                mockServer.URL + "/successful",
		ExpectedStatusCode: http.StatusOK,
		Timeout:            500 * time.Millisecond,
		ExpectedLocation:   "",
		ExpectedContent:    "Successful Request",
	}

	appWithoutContent := Application{
		URL:                mockServer.URL + "/successful",
		ExpectedStatusCode: http.StatusOK,
		Timeout:            500 * time.Millisecond,
		ExpectedLocation:   "",
		ExpectedContent:    "",
	}

	var tests = []struct {
		description         string
		app                 Application
		statusOk            bool
		statusContentOk     bool
		expectedApplication *AppCheckStatus
	}{
		{
			description:     "GET request with network error",
			app:             app,
			statusOk:        false,
			statusContentOk: false,
			expectedApplication: &AppCheckStatus{
				Application:      &app,
				StatusOk:         false,
				StatusContentOk:  false,
				ActualStatusCode: 0,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
		{
			description:     "HEAD request with network error",
			app:             app,
			statusOk:        false,
			statusContentOk: false,
			expectedApplication: &AppCheckStatus{
				Application:      &app,
				StatusOk:         false,
				StatusContentOk:  false,
				ActualStatusCode: 0,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
		{
			description:     "Successful GET request",
			app:             app,
			statusOk:        true,
			statusContentOk: true,
			expectedApplication: &AppCheckStatus{
				Application:      &app,
				StatusOk:         true,
				StatusContentOk:  true,
				StatusCSPOk:      true,
				ActualStatusCode: http.StatusOK,
				ActualLocation:   "",
				ActualContent:    "Successful Request",
			},
		},
		{
			description:     "Successful HEAD request",
			app:             appWithoutContent,
			statusOk:        true,
			statusContentOk: true,
			expectedApplication: &AppCheckStatus{
				Application:      &appWithoutContent,
				StatusOk:         true,
				StatusContentOk:  true,
				StatusCSPOk:      true,
				ActualStatusCode: http.StatusOK,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Here we create the http.Client to use in the tests.
			// If the test is supposed to simulate a network error, use the errorTransport defined above.
			// Otherwise, use a regular http.Client.
			var client *http.Client
			if test.description == "GET request with network error" || test.description == "HEAD request with network error" {
				client = &http.Client{
					Transport: &errorTransport{},
					Timeout:   test.app.Timeout,
				}
			} else {
				client = &http.Client{
					Timeout: test.app.Timeout,
				}
			}

			var resp *http.Response
			var err error
			var actualContent string
			var statusContentOk bool
			var statusCode int

			if test.app.IsGet() {
				statusCode, _, actualContent, statusContentOk, err = performGetRequest(test.app, client)
				if err == nil {
					resp = &http.Response{StatusCode: statusCode, Header: http.Header{}}
				}
			} else {
				resp, err = performHeadRequest(test.app, client)
				statusContentOk = err == nil
			}

			if err != nil {
				statusContentOk = false
			}

			result := createApplicationStatus(test.app, resp, err, actualContent, statusContentOk)
			assert.Equal(t, test.expectedApplication, result)
		})
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		description    string
		appStatus      *AppCheckStatus
		expectedOutput string
	}{
		{description: "Successful status", appStatus: &AppCheckStatus{Application: &Application{"", "https://library.nyu.edu", http.StatusOK, time.Second, "", "", ""}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200}, expectedOutput: "Success: URL https://library.nyu.edu resolved with 200"},
		{description: "Failed status", appStatus: &AppCheckStatus{Application: &Application{URL: "https://library.nyu.edu", ExpectedStatusCode: http.StatusOK, Timeout: time.Second}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 404}, expectedOutput: "Failure: URL https://library.nyu.edu resolved with 404, expected 200"},
		{description: "Successful status with location", appStatus: &AppCheckStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "https://library.nyu.edu/"}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Success: URL http://library.nyu.edu resolved with 301, redirect location matched https://library.nyu.edu/"},
		{description: "Failed status with location", appStatus: &AppCheckStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "http://library.nyu.edu/"}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Failure: URL http://library.nyu.edu resolved with 301, but redirect location https://library.nyu.edu/ did not match http://library.nyu.edu/"},
		{description: "Successful status with expected content", appStatus: &AppCheckStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Example Domain", ""}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Success: URL https://example.com resolved with 200\nSuccess: ExpectedContent Example Domain matched ActualContent Example Domain"},
		{description: "Failed status with unexpected content", appStatus: &AppCheckStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Wrong Content", ""}, StatusOk: true, StatusContentOk: false, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Success: URL https://example.com resolved with 200\nFailure: Expected content Wrong Content did not match Actual Content"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, test.appStatus.String())
		})
	}
}

func TestCompareContent(t *testing.T) {
	var tests = []struct {
		description string
		actual      string
		expected    string
		wantBool    bool
		wantStr     string
	}{
		{description: "Expected content is found at the beginning of the actual string", actual: "hello world", expected: "hello", wantBool: true, wantStr: "hello"},
		{description: "Expected content is found at the end of the actual string", actual: "hello world", expected: "world", wantBool: true, wantStr: "world"},
		{description: "Expected content is not found in the actual string", actual: "hello world", expected: "earth", wantBool: false, wantStr: "hello world"},
		{description: "Expected content is found in the middle of the actual string", actual: "hello beautiful world", expected: "beautiful", wantBool: true, wantStr: "beautiful"},
		{description: "Actual content is empty", actual: "", expected: "world", wantBool: false, wantStr: ""},
		{description: "Expected content is empty", actual: "hello world", expected: "", wantBool: true, wantStr: ""},
		{description: "Both actual content and expected content are empty", actual: "", expected: "", wantBool: true, wantStr: ""},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			gotBool, gotStr := compareContent(tt.actual, tt.expected)
			assert.Equal(t, tt.wantBool, gotBool)
			assert.Equal(t, tt.wantStr, gotStr)
		})
	}
}

func TestCompareLocations(t *testing.T) {
	var tests = []struct {
		description string
		actual      string
		expected    string
		wantBool    bool
	}{
		{description: "Identical locations", actual: "New York", expected: "New York", wantBool: true},
		{description: "Different locations", actual: "New York", expected: "San Francisco", wantBool: false},
		{description: "Empty expected location", actual: "New York", expected: "", wantBool: false},
		{description: "Empty actual location", actual: "", expected: "San Francisco", wantBool: false},
		{description: "Both locations empty", actual: "", expected: "", wantBool: true},
		{description: "Relative path without leading slash matches", actual: "/mng/login", expected: "mng/login", wantBool: true},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.wantBool, compareLocations(tt.actual, tt.expected), tt.description)
		})
	}
}

func TestCompareStatusCodes(t *testing.T) {
	var tests = []struct {
		description string
		actual      int
		expected    int
		want        bool
	}{
		{description: "Identical status codes", actual: 200, expected: 200, want: true},
		{description: "Different status codes", actual: 200, expected: 404, want: false},
		{description: "Zero actual status code", actual: 0, expected: 200, want: false},
		{description: "Zero expected status code", actual: 200, expected: 0, want: false},
		{description: "Both status codes zero", actual: 0, expected: 0, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.want, compareStatusCodes(tt.actual, tt.expected), tt.description)
		})
	}
}

func TestStringWithEnvVarsSuccessAndContent(t *testing.T) {

	originalIsPrimoVE := IsPrimoVE
	originalDebugMode := DebugMode
	defer func() {
		IsPrimoVE = originalIsPrimoVE
		DebugMode = originalDebugMode
	}()

	var tests = []struct {
		description    string
		appStatus      *AppCheckStatus
		isPrimoVE      bool
		debugMode      bool
		expectedOutput string
	}{
		{
			description: "Successful status, IsPrimoVE=false, DebugMode=false",
			isPrimoVE:   false,
			debugMode:   false,
			appStatus: &AppCheckStatus{
				Application: &Application{Name: "TestApp", URL: "https://example.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second * 5, ExpectedContent: "Example Content"},
				StatusOk:    true, StatusContentOk: true,
				ActualStatusCode: 200, ActualContent: "Example Content"},
			expectedOutput: "Success: URL https://example.com resolved with 200\nSuccess: ExpectedContent Example Content matched ActualContent Example Content",
		},
		{
			description: "Successful status, IsPrimoVE=true, DebugMode=true, with detailed content",
			isPrimoVE:   true,
			debugMode:   true,
			appStatus: &AppCheckStatus{
				Application: &Application{Name: "TestApp", URL: "https://example.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second * 5, ExpectedContent: "Example Content"},
				StatusOk:    true, StatusContentOk: true,
				ActualStatusCode: 200, ActualContent: "Example Content"},
			expectedOutput: "Success: URL https://example.com resolved with 200\nSuccess: ExpectedContent Example Content matched ActualContent Example Content",
		},
		{
			description: "Failed status with unexpected content, DebugMode=false",
			isPrimoVE:   false,
			debugMode:   false,
			appStatus: &AppCheckStatus{
				Application: &Application{Name: "TestApp", URL: "https://example.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second * 5, ExpectedContent: "Expected Content"},
				StatusOk:    true, StatusContentOk: false,
				ActualStatusCode: 200, ActualContent: "Some Actual Content"},
			expectedOutput: "Success: URL https://example.com resolved with 200\nFailure: Expected content Expected Content did not match Actual Content",
		},
		{
			description: "Failed status with unexpected content, DebugMode=true, with detailed content",
			isPrimoVE:   true,
			debugMode:   true,
			appStatus: &AppCheckStatus{
				Application: &Application{Name: "TestApp", URL: "https://example.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second * 5, ExpectedContent: "Expected Content"},
				StatusOk:    true, StatusContentOk: false,
				ActualStatusCode: 200, ActualContent: "Some Actual Content"},

			expectedOutput: "Success: URL https://example.com resolved with 200\nFailure: Expected content Expected Content did not match Actual Content Some Actual Content",
		},
		{
			description: "Failed status with unexpected content, DebugMode=true, with detailed content",
			isPrimoVE:   false,
			debugMode:   true,
			appStatus: &AppCheckStatus{
				Application: &Application{Name: "TestApp", URL: "https://example.com", ExpectedStatusCode: http.StatusOK, Timeout: time.Second * 5, ExpectedContent: "Expected Content"},
				StatusOk:    true, StatusContentOk: false,
				ActualStatusCode: 200, ActualContent: "Some Actual Content"},

			expectedOutput: "Success: URL https://example.com resolved with 200\nFailure: Expected content Expected Content did not match Actual Content",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			IsPrimoVE = test.isPrimoVE
			DebugMode = test.debugMode

			assert.Equal(t, test.expectedOutput, test.appStatus.String())
		})
	}
}

type hop struct {
	path     string
	status   int
	location string
	body     string
}

func newRedirectServer(hops ...hop) *httptest.Server {

	table := make(map[string]hop, len(hops))
	for _, h := range hops {
		table[h.path] = h
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := table[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if h.location != "" {
			w.Header().Set("Location", h.location)
		}
		w.WriteHeader(h.status)
		if h.body != "" {
			_, _ = w.Write([]byte(h.body))
		}
	}))
}

func TestGetStatus_RelativeRedirect_AndContentSuccess(t *testing.T) {
	// / → 302 Location:/mng/login
	// /mng/login → 302 Location:/discovery/search
	// /discovery/search → 200 body contains the expected substring
	const expectedSnippet = "<primo-explore>"

	srv := newRedirectServer(
		hop{path: "/", status: http.StatusFound, location: "/mng/login"},
		hop{path: "/mng/login", status: http.StatusFound, location: "/discovery/search"},
		hop{path: "/discovery/search", status: http.StatusOK, body: expectedSnippet},
	)
	t.Cleanup(srv.Close)

	app := &Application{
		Name:               "primo-ve-search",
		URL:                srv.URL + "/",
		ExpectedStatusCode: http.StatusFound,
		Timeout:            2 * time.Second,
		ExpectedLocation:   "/mng/login",
		ExpectedContent:    expectedSnippet,
	}

	status := app.GetStatus()

	require.NotNil(t, status, "GetStatus must return a non-nil status")
	require.Equalf(t, http.StatusFound, status.ActualStatusCode,
		"source status mismatch for %s", app.URL)
	require.Equalf(t, "/mng/login", status.ActualLocation,
		"first-hop Location mismatch for %s", app.URL)

	assert.Truef(t, status.StatusOk, "redirect check should pass (status=%d location=%q)",
		status.ActualStatusCode, status.ActualLocation)
	assert.Truef(t, status.StatusContentOk, "content check should pass; got content=%q",
		status.ActualContent)
	assert.Equalf(t, expectedSnippet, status.ActualContent, "matched snippet should be returned")
}

func TestGetStatus_RelativeRedirect_AndContent_Failure_ContentMismatch(t *testing.T) {
	// Same redirect chain, but FINAL page omits the expected substring.
	const expectedSnippet = "Log into your account"
	const deliveredBody = "<primo-explore>Welcome page without login text</primo-explore>"

	srv := newRedirectServer(
		hop{path: "/", status: http.StatusFound, location: "/mng/login"},
		hop{path: "/mng/login", status: http.StatusFound, location: "/discovery/search"},
		hop{path: "/discovery/search", status: http.StatusOK, body: deliveredBody},
	)
	t.Cleanup(srv.Close)

	app := &Application{
		Name:               "primo-ve-search",
		URL:                srv.URL + "/",
		ExpectedStatusCode: http.StatusFound,
		Timeout:            2 * time.Second,
		ExpectedLocation:   "/mng/login",
		ExpectedContent:    expectedSnippet, // NOT present on final page
	}

	status := app.GetStatus()

	require.NotNil(t, status, "GetStatus must return a non-nil status")
	require.Equalf(t, http.StatusFound, status.ActualStatusCode,
		"source status mismatch for %s", app.URL)
	require.Equalf(t, "/mng/login", status.ActualLocation,
		"first-hop Location mismatch for %s", app.URL)

	assert.Truef(t, status.StatusOk, "redirect check should pass (status=%d location=%q)",
		status.ActualStatusCode, status.ActualLocation)
	assert.Falsef(t, status.StatusContentOk, "content check should fail; expected substring %q", expectedSnippet)
	assert.NotEmpty(t, status.ActualContent, "body should be preserved on mismatch")
	assert.Equalf(t, deliveredBody, status.ActualContent,
		"on mismatch we should capture the full final-page body")
}
