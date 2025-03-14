#!/bin/sh

# Check if the output should be posted to Slack
IsOutputSlack=$(echo "${OUTPUT_SLACK:-false}")

# Capture the output of /aswa
aswa_output=$(if ! command -v "$1" >/dev/null 2>&1; then /aswa "$1"; else "$@"; fi)

if [ "$IsOutputSlack" = "true" ] && echo "$aswa_output" | grep -q "Failure"; then
    # Check if CLUSTER_INFO is set and process accordingly
    if [ -z "$CLUSTER_INFO" ]; then
        cluster_name="Unknown cluster: CLUSTER_INFO is not set"
    else
        cluster_name=$(echo "$CLUSTER_INFO" | tr '[:lower:]' '[:upper:]')
    fi

    slack_message="${cluster_name}: ${aswa_output}"

    # Construct JSON payload using jq
    payload=$(echo '{}' | jq --arg text "$slack_message" '.text = $text')

    # Send the constructed message to Slack webhook URL
    curl -X POST -H 'Content-type: application/json' --data "$payload" "${SLACK_WEBHOOK_URL}"
fi
