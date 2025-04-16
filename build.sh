#!/bin/bash
set -euo pipefail
set -x

VERSION="$(git describe --tags --dirty --always --abbrev=7)"
COMMIT="$(git describe --always --dirty --abbrev=7)"
BUILDDATE="$(date -u +%Y-%m-%d)"

LDARGS="-X main.VERSION=$VERSION -X main.COMMIT=$COMMIT -X main.BUILDDATE=$BUILDDATE"

SRC="./cmd/meshtastic-go"
NAME="meshtastic-go"

env GOOS=darwin  GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_osx                 $SRC
env GOOS=linux   GOARCH=arm GOARM=5 go build -ldflags="$LDARGS" -o builds/${NAME}_linux_arm           $SRC
env GOOS=linux   GOARCH=arm64       go build -ldflags="$LDARGS" -o builds/${NAME}_linux_arm64         $SRC
env GOOS=linux   GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_linux_amd64         $SRC
env GOOS=linux   GOARCH=386         go build -ldflags="$LDARGS" -o builds/${NAME}_linux_x86           $SRC
env GOOS=windows GOARCH=amd64       go build -ldflags="$LDARGS" -o builds/${NAME}_windows_amd64.exe   $SRC
