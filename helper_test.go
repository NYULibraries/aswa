package main

import (
	"errors"
	"log"
	"testing"
	"time"

	a "github.com/NYULibraries/aswa/lib/application"
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

func testRunSyntheticTestsFunc(t *testing.T, appData []*a.Application, channel string, token string, mockError error) error {
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
		err := mockAppStatus.postTestResult(app, channel, token)

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

func TestRunSyntheticTests(t *testing.T) {
	var tests = []struct {
		description string
		appData     []*a.Application
		channel     string
		token       string
		cmdArg      string
		error       error
	}{
		{"Valid test run with cmdArgs", []*a.Application{{Name: "test", URL: "test", ExpectedStatusCode: 200, Timeout: 1 * time.Second, ExpectedLocation: "test"}}, "test", "test", "test", nil},
		{"Valid test run without cmdArgs", []*a.Application{{Name: "test", URL: "test", ExpectedStatusCode: 200, Timeout: 1 * time.Second, ExpectedLocation: "test"}}, "test", "test", "", nil},
		{"Invalid Test Run Tests No Cmd Args", []*a.Application{{Name: "", URL: "", ExpectedStatusCode: 200, Timeout: 1 * time.Second}}, "", "", "", errors.New("application Name & Url not provided, aborting")},
		{"Run with invalid slack credentials", []*a.Application{{Name: "collections", URL: "www.collections.com", ExpectedStatusCode: 304, Timeout: 1 * time.Second}}, "collections", "invalid_token", "", errors.New("invalid slack credentials: invalid token")},
		{"Run with no slack credentials", []*a.Application{{Name: "collections", URL: "www.collections.com", ExpectedStatusCode: 304, Timeout: 1 * time.Second}}, "", "", "", nil},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := testRunSyntheticTestsFunc(t, test.appData, test.channel, test.token, test.error)
			if err != nil {
				log.Println(err)
			}
		})
	}
}
