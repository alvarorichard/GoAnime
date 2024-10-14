#!/bin/sh

set -e

# URL do binário do GoAnime
URL="https://github.com/alvarorichard/GoAnime/releases/download/v1.0.5/goanime"

# Baixa o binário
echo "Baixando o GoAnime v1.0.5..."
curl -L "${URL}" -o "goanime"
chmod +x "goanime"

# Move o binário para /usr/local/bin
echo "Instalando o GoAnime..."
if [ "$(id -u)" -ne 0 ]; then
  sudo mv "goanime" /usr/local/bin/goanime
  sudo ln -sf /usr/local/bin/goanime /usr/local/bin/go-anime
else
  mv "goanime" /usr/local/bin/goanime
  ln -sf /usr/local/bin/goanime /usr/local/bin/go-anime
fi

echo "GoAnime instalado com sucesso!"
