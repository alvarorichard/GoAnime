#!/bin/bash

# Assuming the Go binary is in /usr/local/go/bin
GO_PATH="/usr/local/go/bin/go"

GOOS=$($GO_PATH env GOOS)
GOARCH=$($GO_PATH env GOARCH)

if [ "$(uname)" == "Darwin" ]; then
    GOOS=darwin
    GOARCH=amd64 # or arm64 for M1 Macs
fi

function compile(){
  # Using the full path for the Go executable
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
