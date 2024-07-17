package utils

import (
	"log"
	"os"
)

// Constants for environment variables
const (
	EnvClusterInfo               = "CLUSTER_INFO"
	EnvName                      = "ENV"
	EnvPromAggregationGatewayUrl = "PROM_AGGREGATION_GATEWAY_URL"
	EnvSlackWebhookUrl           = "SLACK_WEBHOOK_URL"
	EnvYamlPath                  = "YAML_PATH"
)

// GetClusterInfo retrieves the cluster info from environment variables.
func GetClusterInfo() string {
	clusterInfo := os.Getenv(EnvClusterInfo)
	if clusterInfo == "" {
		log.Println("CLUSTER_INFO is not set")
		return ""
	}
	return clusterInfo
}

// GetCmdArg retrieves the command line argument without using the flag package.
func GetCmdArg() string {
	if len(os.Args) == 1 {
		return ""
	}
	return os.Args[1]
}

// GetEnvironmentName retrieves the environment name from environment variables, defaults to 'dev' if not set
func GetEnvironmentName() string {
	env := os.Getenv(EnvName)
	if env == "" {
		return "dev"
	}
	return env
}

// GetPromAggregationgatewayUrl retrieves the pag url from environment variables.
func GetPromAggregationgatewayUrl() string {
	promAggregationGatewayUrl := os.Getenv(EnvPromAggregationGatewayUrl)
	if promAggregationGatewayUrl == "" {
		log.Println("PROM_AGGREGATION_GATEWAY_URL is not set")
		return ""
	}
	return promAggregationGatewayUrl
}

// GetSlackCredentials retrieves Slack credentials from environment variables.
func GetSlackWebhookUrl() string {
	slackWebhookUrl := os.Getenv(EnvSlackWebhookUrl)
	if slackWebhookUrl == "" {
		log.Println("SLACK_WEBHOOK_URL is not set")
		return ""
	}
	return slackWebhookUrl
}

// GetYamlPath retrieves the YAML path from the environment variable.
func GetYamlPath(logger *log.Logger) string {
	yamlPath := os.Getenv(EnvYamlPath)
	if yamlPath == "" {
		logger.Println("Environment variable for YAML path not found, using default")
		yamlPath = "config/dev.applications.yml"
	}
	return yamlPath
}
