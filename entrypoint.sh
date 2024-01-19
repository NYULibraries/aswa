#!/bin/sh

# Capture the output of /aswa
aswa_output=$(if ! command -v "$1" >/dev/null 2>&1; then /aswa "$1"; else "$@"; fi)

if echo "$aswa_output" | grep -q "Failure"; then
    cluster_name=${CLUSTER_INFO:-"Unknown cluster:CLUSTER_INFO is not set"}
    slack_message="${cluster_name}: ${aswa_output}"

    # Construct JSON payload using jq
    payload=$(echo '{}' | jq --arg text "$slack_message" '.text = $text')

    curl -X POST -H 'Content-type: application/json' --data "$payload" "${SLACK_WEBHOOK_URL}"
fi
