#!/usr/bin/env bash

set -euo pipefail

GOCACHE="$PWD/go-build"

GOOS=linux    GOARCH=amd64 go build -i -ldflags='-s -w' -o bin/build build/main.go
GOOS=linux    GOARCH=amd64 go build -i -ldflags='-s -w' -o bin/detect detect/main.go