#!/bin/sh

if ! command -v "$1" >/dev/null 2>&1; then
    /aswa "$1"
else
    exec "$@"
fi
