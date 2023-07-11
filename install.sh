GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

function compile(){
  GOOS=$GOOS GOARCH=$GOARCH go build main.go
}

function start(){
  compile
  mv main /usr/local/bin/goanime
  ln -sf /usr/local/bin/goanime /usr/bin/go-anime
}

if [ "$EUID" -eq 0 ]; then
  start
else
  echo "Este programa deve ser rodado como sudo"
fi
