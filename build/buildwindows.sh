#!/bin/bash

# Sai imediatamente se um comando falhar
set -e

# Variáveis
OUTPUT_DIR="../build"        # Diretório de saída para o binário e checksum
BINARY_NAME="goanime.exe"    # Nome do binário para Windows
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"
ZIP_NAME="goanime-windows.zip"
ZIP_PATH="$OUTPUT_DIR/$ZIP_NAME"
CHECKSUM_FILE="$ZIP_PATH.sha256"
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

echo "Compilando o binário goanime para Windows ($GOARCH)..."
# Para Windows, precisamos garantir que o go-winio esteja disponível
# Mas com CGO_ENABLED=0, o SQLite será compilado sem suporte nativo
CGO_ENABLED=0 GOOS=windows GOARCH=$GOARCH go build -o "$BINARY_PATH" -ldflags="-s -w" -trimpath -tags="windows" "$MAIN_PACKAGE"

echo "Compilação concluída: $BINARY_PATH"

# Verifica se o UPX está instalado
if command -v upx >/dev/null 2>&1; then
    echo "Comprimindo o binário com UPX..."
    upx --best --ultra-brute "$BINARY_PATH"
    echo "Compressão concluída."
else
    echo "UPX não encontrado. Pulando compressão."
fi

# Cria arquivo ZIP
echo "Criando arquivo ZIP..."
zip -j "$ZIP_PATH" "$BINARY_PATH"
echo "Arquivo ZIP criado: $ZIP_PATH"

# Gera o checksum SHA256 para o arquivo ZIP
echo "Gerando checksum SHA256 para o arquivo ZIP..."
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$ZIP_PATH" > "$CHECKSUM_FILE"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$ZIP_PATH" > "$CHECKSUM_FILE"
elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$ZIP_PATH" | awk '{print $2}' > "$CHECKSUM_FILE"
else
    echo "Nem sha256sum, shasum ou openssl estão disponíveis. Não é possível gerar o checksum."
    exit 1
fi
echo "Checksum gerado: $CHECKSUM_FILE"

echo "Script de build concluído com sucesso."
