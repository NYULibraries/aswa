package main

import (
	"errors"
	"log"
	"testing"
	"time"

	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
	"github.com/stretchr/testify/assert"
)

type mockApplication struct {
	mockName               string
	mockURL                string
	mockExpectedStatusCode int
	mockTimeout            time.Duration
	mockExpectedLocation   string
	mockError              error
}

type mockApplicationStatus struct {
	mockApplication      *mockApplication
	mockActualStatusCode int
}

func (m *mockApplicationStatus) postTestResult(test *a.Application, channel string, token string) error {
	appStatus := test.GetStatus()
	m.mockActualStatusCode = test.GetStatus().ActualStatusCode

	slackClient := NewSlackClient(token)
	err := slackClient.PostToSlack(appStatus.String(), channel)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func testRunTestsFunc(t *testing.T, appData []*c.Application, channel string, token string, mockError error) error {
	mockApp := &mockApplication{
		mockName:               "test",
		mockURL:                "test",
		mockExpectedStatusCode: 200,
		mockTimeout:            1 * time.Second,
		mockExpectedLocation:   "test",
		mockError:              mockError,
	}

	mockAppStatus := &mockApplicationStatus{
		mockApplication:      mockApp,
		mockActualStatusCode: 200,
	}

	for _, app := range appData {
		name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

		test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)

		err := mockAppStatus.postTestResult(test, channel, token)

		if err != nil {
			log.Println(err)
			return err
		}

		assert.Equal(t, appData[0], mockAppStatus.mockApplication)
		assert.Equal(t, mockAppStatus.mockActualStatusCode, appData[0].ExpectedStatusCode)
		assert.Equal(t, mockApp.mockError, err)
	}

	return nil
}

func TestRunTests(t *testing.T) {
	var tests = []struct {
		description string
		appData     []*c.Application
		channel     string
		token       string
		error       error
	}{
		{"Valid test run", []*c.Application{{Name: "test", URL: "test", ExpectedStatusCode: 200, Timeout: 1 * time.Second, ExpectedLocation: "test"}}, "test", "test", nil},
		{"Test Run Tests No Cmd Args", []*c.Application{{Name: "collections", URL: "www.collections.com", ExpectedStatusCode: 304, Timeout: 1 * time.Second}}, "collections", "www.collections.com", nil},
		{"Invalid Test Run Tests No Cmd Args", []*c.Application{{Name: "", URL: "", ExpectedStatusCode: 200, Timeout: 1 * time.Second}}, "", "", errors.New("application Name & Url not provided, aborting")},
		{"Test Run Tests Invalid Credentials", []*c.Application{{Name: "collections", URL: "www.collections.com", ExpectedStatusCode: 304, Timeout: 1 * time.Second}}, "collections", "invalid_token", errors.New("invalid slack credentials: invalid token")},
		{"Test Run Tests No Credentials", []*c.Application{{Name: "collections", URL: "www.collections.com", ExpectedStatusCode: 304, Timeout: 1 * time.Second}}, "", "", nil},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := testRunTestsFunc(t, test.appData, test.channel, test.token, test.error)
			if err != nil {
				log.Println(err)
			}
		})
	}
}
