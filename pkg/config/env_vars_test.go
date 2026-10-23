package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
