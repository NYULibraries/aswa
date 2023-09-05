#!/bin/sh

# Debug: Print SLACK_WEBHOOK_URL
echo "SLACK_WEBHOOK_URL: ${SLACK_WEBHOOK_URL}"

# Capture the output of /aswa
aswa_output=$(if ! command -v "$1" >/dev/null 2>&1; then /aswa "$1"; else "$@"; fi)

# Debug: Print aswa_output
echo "aswa_output: ${aswa_output}"

if echo "$aswa_output" | grep -q "Failure"; then
    cluster_name=${CLUSTER_INFO:-"Unknown cluster:CLUSTER_INFO is not set"}
    slack_message="${cluster_name}: ${aswa_output}"

    # Debug: Print slack_message
    echo "slack_message: ${slack_message}"

    # Construct JSON payload using jq
    payload=$(echo '{}' | jq --arg text "$slack_message" '.text = $text')

    curl -X POST -H 'Content-type: application/json' --data "$payload" "${SLACK_WEBHOOK_URL}"
fi
