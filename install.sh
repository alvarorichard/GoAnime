#!/bin/bash

# Find the full path of the Go executable
GO_PATH=$(which go)

# Check if Go is installed
if [ -z "$GO_PATH" ]; then
    echo "Go is not installed or not in the PATH"
    exit 1
fi

# Get GOOS and GOARCH
GOOS=$($GO_PATH env GOOS)
GOARCH=$($GO_PATH env GOARCH)

# Determine the installation directory
if [ -w /usr/local/bin ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="/usr/bin"
fi

if [ "$(uname)" == "Darwin" ]; then
    GOOS=darwin
    GOARCH=amd64 # or arm64 for M1 Macs
fi

function compile(){
  # Use the full path for the Go executable
  GOOS=$GOOS GOARCH=$GOARCH $GO_PATH build main.go
}

function install_macos(){
  mv main /usr/local/bin/goanime
  ln -sf /usr/local/bin/goanime /usr/local/bin/go-anime
}

function install_others(){
  mv main /usr/bin/goanime
  ln -sf /usr/bin/goanime /usr/bin/go-anime
}

function start(){
  compile
  if [ "$(uname)" == "Darwin" ]; then
    install_macos
  else
    install_others
  fi
}

if [ "$EUID" -eq 0 ]; then
  start
else
  echo "Este programa deve ser rodado como sudo"
  # Optionally add an English translation
  echo "This program must be run as sudo"
fi
