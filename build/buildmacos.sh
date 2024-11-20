#!/bin/bash

# Sai imediatamente se um comando falhar
set -e

# Variáveis
OUTPUT_DIR="../build"
BINARY_NAME="goanime"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"
CHECKSUM_FILE="$OUTPUT_DIR/$BINARY_NAME.sha256"
MAIN_PACKAGE="../cmd/goanime"

# Detecta a arquitetura
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
    GOARCH=amd64
elif [ "$ARCH" == "arm64" ]; then
    GOARCH=arm64
else
    echo "Arquitetura não suportada: $ARCH"
    exit 1
fi

# Cria o diretório de saída se não existir
mkdir -p "$OUTPUT_DIR"

echo "Compilando o binário goanime para macOS ($GOARCH)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=$GOARCH go build -o "$BINARY_PATH" -ldflags="-s -w" -trimpath "$MAIN_PACKAGE"

echo "Compilação concluída: $BINARY_PATH"

# Gera o checksum SHA256
echo "Gerando checksum SHA256..."
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$BINARY_PATH" > "$CHECKSUM_FILE"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$BINARY_PATH" > "$CHECKSUM_FILE"
else
    echo "Nem sha256sum nem shasum estão disponíveis. Não é possível gerar o checksum."
    exit 1
fi
echo "Checksum gerado: $CHECKSUM_FILE"

echo "Script de build concluído com sucesso."
