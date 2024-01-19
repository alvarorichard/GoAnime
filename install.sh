#!/bin/bash

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

function compile(){
  GOOS=$GOOS GOARCH=$GOARCH go build main.go
  if [ "$(uname)" == "Darwin" ]; then
    # Se o sistema é macOS, ajuste GOOS e GOARCH conforme necessário
    GOOS=darwin
    GOARCH=amd64 # ou arm64 para Macs com chip M1
  fi
  GOOS=$GOOS GOARCH=$GOARCH go build main.go
}
# add bin to path macOS only
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
fi
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

function compile(){
  GOOS=$GOOS GOARCH=$GOARCH go build main.go
  if [ "$(uname)" == "Darwin" ]; then
    # Se o sistema é macOS, ajuste GOOS e GOARCH conforme necessário
    GOOS=darwin
    GOARCH=amd64 # ou arm64 para Macs com chip M1
  fi
  GOOS=$GOOS GOARCH=$GOARCH go build main.go
}
# add bin to path macOS only
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
fi
