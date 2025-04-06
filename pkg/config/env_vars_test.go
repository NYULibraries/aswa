package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

// CaptureOutput captures the log output of a function
func CaptureOutput(f func()) string {
	var buf bytes.Buffer
	// Save the current flags and set log flags to exclude date and time
	savedFlags := log.Flags()
	log.SetFlags(0)
	log.SetOutput(&buf)
	defer func() {
		// Restore the original log flags and output
		log.SetFlags(savedFlags)
		log.SetOutput(os.Stderr)
	}()
	f()
	return buf.String()
}

func TestGetYamlPath(t *testing.T) {
	tests := []struct {
		name        string
		envYamlPath string
		want        string
	}{
		{"EnvYamlPath is set", "test-path", "test-path"},
		{"EnvYamlPath is not set", "", "config/dev.applications.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Setenv(EnvYamlPath, tt.envYamlPath)

			got := GetYamlPath()

			assert.Equal(t, tt.want, got, "getYamlPath() should return correct yaml path")

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

			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()
			os.Args = tt.osArgs

			got := GetCmdArg()

			assert.Equal(t, tt.want, got, "getCmdArg() should return correct command argument")
		})
	}
}

func TestGetSlackWebhookUrl(t *testing.T) {
	tests := []struct {
		name               string
		envSlackWebhookUrl string
		wantLogmessage     string
	}{
		{"EnvSlackWebhookUrl is set", "https://hooks.slack.com/test-url", ""},
		{"EnvSlackWebhookUrl is not set", "", "SLACK_WEBHOOK_URL is not set\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Setenv(EnvSlackWebhookUrl, tt.envSlackWebhookUrl)

			gotLogMessage := CaptureOutput(GetSlackWebhookUrl)

			assert.Equal(t, tt.wantLogmessage, gotLogMessage, "getSlackWebhookUrl() should return correct Slack webhook URL")

		})
	}
}

func TestGetClusterInfo(t *testing.T) {
	tests := []struct {
		name           string
		envClusterInfo string
		wantLogMessage string
	}{
		{"EnvClusterInfo is set with lower case", "test-cluster", ""},
		{"EnvClusterInfo is not set", "", "CLUSTER_INFO is not set\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up environment variable
			t.Setenv(EnvClusterInfo, tt.envClusterInfo)

			// Call function under test
			gotLogMessage := CaptureOutput(GetClusterInfo)

			// Assert that the function returns the expected result
			assert.Equal(t, tt.wantLogMessage, gotLogMessage, "getClusterInfo() should return correct cluster info")

		})
	}
}

func TestGetPromAggregationGatewayUrl(t *testing.T) {
	tests := []struct {
		name                         string
		envPromAggregationGatewayUrl string
		want                         string
	}{
		{"EnvPromAggregationGatewayUrl is set", "test-promaggregationgateway-url", "test-promaggregationgateway-url"},
		{"EnvPromAggregationGatewayUrl is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Setenv(EnvPromAggregationGatewayUrl, tt.envPromAggregationGatewayUrl)

			got := GetPromAggregationgatewayUrl()

			assert.Equal(t, tt.want, got, "getPushgatewayUrl() should return correct pushgateway URL")

		})
	}

}

func TestGetEnvironmentName(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		want    string
	}{
		{"EnvName is set", "test-env", "test-env"},
		{"EnvName is not set", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Setenv(EnvName, tt.envName)

			got := GetEnvironmentName()

			assert.Equal(t, tt.want, got, "GetEnvironmentName() should return correct environment name")

		})
	}
}
