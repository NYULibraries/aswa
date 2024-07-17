package utils

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

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

			os.Setenv(EnvYamlPath, tt.envYamlPath)

			got := GetYamlPath(log.New(os.Stdout, "", 0))

			assert.Equal(t, tt.want, got, "getYamlPath() should return correct yaml path")

			os.Unsetenv(EnvYamlPath)
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
		want               string
	}{
		{"EnvSlackWebhookUrl is set", "https://hooks.slack.com/test-url", "https://hooks.slack.com/test-url"},
		{"EnvSlackWebhookUrl is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv(EnvSlackWebhookUrl, tt.envSlackWebhookUrl)

			got := GetSlackWebhookUrl()

			assert.Equal(t, tt.want, got, "getSlackWebhookUrl() should return correct Slack webhook URL")

			os.Unsetenv(EnvSlackWebhookUrl)
		})
	}
}

func TestGetClusterInfo(t *testing.T) {
	tests := []struct {
		name           string
		envClusterInfo string
		want           string
	}{
		{"EnvClusterInfo is set", "test-cluster", "test-cluster"},
		{"EnvClusterInfo is not set", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up environment variable
			os.Setenv(EnvClusterInfo, tt.envClusterInfo)

			// Call function under test
			got := GetClusterInfo()

			// Assert that the function returns the expected result
			assert.Equal(t, tt.want, got, "getClusterInfo() should return correct cluster info")

			// Unset environment variable for next test
			os.Unsetenv(EnvClusterInfo)
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

			os.Setenv(EnvPromAggregationGatewayUrl, tt.envPromAggregationGatewayUrl)

			got := GetPromAggregationgatewayUrl()

			assert.Equal(t, tt.want, got, "getPushgatewayUrl() should return correct pushgateway URL")

			os.Unsetenv(EnvPromAggregationGatewayUrl)
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

			os.Setenv(EnvName, tt.envName)

			got := GetEnvironmentName()

			assert.Equal(t, tt.want, got, "GetEnvironmentName() should return correct environment name")

			os.Unsetenv(EnvName)
		})
	}
}
