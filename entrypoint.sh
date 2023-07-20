#!/bin/sh

# Capture the output of /aswa
aswa_output=$(if ! command -v "$1" >/dev/null 2>&1; then /aswa "$1"; else "$@" & fi)

echo "aswa_output: ${aswa_output}"

if echo "$aswa_output" | grep -q "Failure"; then
    cluster_name=$CLUSTER_INFO
    slack_message="Cluster ${cluster_name}: ${aswa_output}"
    curl -X POST -H 'Content-type: application/json' --data "{\"text\":\"${slack_message}\"}" "${SLACK_WEBHOOK_URL}"
fi



