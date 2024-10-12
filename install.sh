#!/bin/sh

set -e

# Determina o sistema operacional e arquitetura
OS="$(uname -s)"
ARCH="$(uname -m)"

# Mapeia o sistema operacional e arquitetura para os nomes usados nos binários
case "$OS" in
  Darwin)
    OS='darwin'
    ;;
  Linux)
    OS='linux'
    ;;
  *)
    echo "Sistema operacional não suportado: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH='amd64'
    ;;
  arm64|aarch64)
    ARCH='arm64'
    ;;
  *)
    echo "Arquitetura não suportada: $ARCH"
    exit 1
    ;;
esac

# URL do binário do GoAnime
VERSION="v1.0.5"
BINARY="goanime"
URL="https://github.com/alvarorichard/GoAnime/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}"

# Baixa o binário
echo "Baixando o GoAnime ${VERSION} para ${OS}/${ARCH}..."
curl -L "${URL}" -o "${BINARY}"
chmod +x "${BINARY}"

# Move o binário para /usr/local/bin
echo "Instalando o GoAnime..."
if [ "$(id -u)" -ne 0 ]; then
  sudo mv "${BINARY}" /usr/local/bin/goanime
  sudo ln -sf /usr/local/bin/goanime /usr/local/bin/go-anime
else
  mv "${BINARY}" /usr/local/bin/goanime
  ln -sf /usr/local/bin/goanime /usr/local/bin/go-anime
fi

echo "GoAnime instalado com sucesso!"
