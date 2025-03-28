#!/bin/bash
set -euo pipefail
set -x

VERSION="$(git describe --tags --abbrev=0)"
COMMIT="$(git rev-parse HEAD)"
BUILDTIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

LDARGS="-X main.VERSION=$VERSION -X main.COMMIT=$COMMIT -X main.BUILDTIME=$BUILDTIME"

SRC="./cmd/meshtastic-go"
NAME="meshtastic-go"

env GOOS=darwin  GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_osx                 $SRC
env GOOS=linux   GOARCH=arm GOARM=5 go build -ldflags="$LDARGS" -o builds/${NAME}_linux_arm           $SRC
env GOOS=linux   GOARCH=arm64       go build -ldflags="$LDARGS" -o builds/${NAME}_linux_arm64         $SRC
env GOOS=linux   GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_linux_amd64         $SRC
env GOOS=linux   GOARCH=386         go build -ldflags="$LDARGS" -o builds/${NAME}_linux_x86           $SRC
env GOOS=windows GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_windows_amd64.exe   $SRC
