$ErrorActionPreference = "Stop"

# Configura caminhos absolutos
$SCRIPT_DIR = $PSScriptRoot
$ROOT_DIR = Split-Path -Parent $SCRIPT_DIR
$OUTPUT_DIR = Join-Path $ROOT_DIR "build"
$BINARY_NAME = "goanime.exe"
$BINARY_PATH = Join-Path $OUTPUT_DIR $BINARY_NAME
$ZIP_NAME = "goanime-windows.zip"
$ZIP_PATH = Join-Path $OUTPUT_DIR $ZIP_NAME
$CHECKSUM_FILE = "$ZIP_PATH.sha256"
$MAIN_PACKAGE = Join-Path $ROOT_DIR "cmd\goanime"

# Detecta arquitetura
$ARCH = $env:PROCESSOR_ARCHITECTURE
if ($ARCH -eq "AMD64") {
    $GOARCH = "amd64"
} elseif ($ARCH -eq "ARM64") {
    $GOARCH = "arm64"
} else {
    Write-Host "Arquitetura não suportada: $ARCH"
    exit 1
}

# Cria diretório de saída
New-Item -ItemType Directory -Force -Path $OUTPUT_DIR | Out-Null

Write-Host "Compilando binário para Windows ($GOARCH)..."
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = $GOARCH

# Executa a compilação
try {
    go build -o $BINARY_PATH -ldflags="-s -w" -trimpath -tags="windows" $MAIN_PACKAGE
    if (-not (Test-Path $BINARY_PATH)) {
        throw "Binário não gerado"
    }
    Write-Host "Compilação concluída: $BINARY_PATH"
}
catch {
    Write-Host "ERRO na compilação: $_"
    exit 1
}

# UPX (opcional)
if (Get-Command upx -ErrorAction SilentlyContinue) {
    Write-Host "Comprimindo com UPX..."
    upx --best --ultra-brute $BINARY_PATH
    Write-Host "Compressão concluída."
}
else {
    Write-Host "UPX não encontrado. Pulando compressão."
}

# Cria ZIP
Write-Host "Criando ZIP..."
try {
    if (Test-Path $ZIP_PATH) {
        Remove-Item $ZIP_PATH -Force
    }
    $compressParams = @{
        Path             = $BINARY_PATH
        DestinationPath  = $ZIP_PATH
        CompressionLevel = "Optimal"
    }
    Compress-Archive @compressParams -Force -ErrorAction Stop
    Write-Host "ZIP criado: $ZIP_PATH"
}
catch {
    Write-Host "ERRO ao criar ZIP: $_"
    exit 1
}

# Checksum
Write-Host "Gerando checksum SHA256..."
try {
    $hash = Get-FileHash -Path $ZIP_PATH -Algorithm SHA256 -ErrorAction Stop
    $hash.Hash.ToLower() | Out-File -FilePath $CHECKSUM_FILE -Encoding ASCII
    Write-Host "Checksum gerado: $CHECKSUM_FILE"
}
catch {
    Write-Host "ERRO ao gerar checksum: $_"
    exit 1
}

Write-Host "Build concluído com sucesso!"