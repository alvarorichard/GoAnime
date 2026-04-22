<h4 align="center">
    <p>
        <b>Рortuguês</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README.md">English</a>
    </p>
</h4>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="Imagem logo do GoAnime" />
</p>

<p align="center">
    <a href="alvarorichard/GoAnime/blob/master/LICENSE"><img src="https://img.shields.io/github/license/alvarorichard/GoAnime" alt="GitHub license"></a>
    <img src="https://img.shields.io/github/stars/alvarorichard/GoAnime" alt="GitHub stars">
    <img src="https://img.shields.io/github/last-commit/alvarorichard/GoAnime" alt="GitHub last commit">
    <img src="https://img.shields.io/github/forks/alvarorichard/GoAnime?style=social" alt="GitHub forks">
    <a href="https://github.com/alvarorichard/GoAnime/actions"><img src="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
    <img src="https://img.shields.io/github/contributors/alvarorichard/GoAnime" alt="GitHub contributors">
    <a href="https://discord.gg/FbQuf78D9G"><img src="https://img.shields.io/badge/Discord-Comunidade-7289DA?logo=discord&logoColor=white" alt="Discord"></a>
</p>

# GoAnime

O GoAnime é uma interface de usuário baseada em texto (TUI) simples,
desenvolvida em Go, que permite aos usuários procurar animes e reproduzir ou
baixar episódios diretamente no mpv. Ele coleta dados de sites para oferecer
conteúdo legendado e dublado em inglês e português.

## Índice

1.  [Recursos](#recursos)
2.  [Pré-requisitos](#pré-requisitos)
3.  [Instalação](#instalação)
4.  [Como usar](#como-usar)
5.  [Uso avançado](#uso-avançado)
6.  [Comunidade e mobile](#comunidade-e-mobile)
7.  [Contribuindo](#contribuindo)

## Recursos

*   Busca de animes, filmes e séries por nome
*   Pesquisa simultânea em todas as fontes ativas por padrão
*   Suporte a conteúdo legendado e dublado em inglês e português
*   Reprodução online com qualidade selecionável (1080p, 720p, etc.)
*   Download único ou em lote de múltiplos episódios
*   Integração com Discord RPC
*   Rastreamento de progresso (retomar reprodução e salvar histórico no SQLite)

*   Upscaling integrado (Anime4K) para melhorar a qualidade de vídeo

## Pré-requisitos

Para que o GoAnime funcione corretamente, instale:
*   [mpv](https://mpv.io/) (Reprodutor de mídia atualizado)

## Instalação

Escolha o método mais adequado para o seu sistema operacional.

### Instalação universal

Se você já possui o Go instalado, pode obter a versão mais recente diretamente:

```bash
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

### macOS

Primeiro, instale o `mpv` usando o Homebrew. Depois instale o GoAnime:

```bash
brew install mpv

curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-apple-darwin
chmod +x goanime
sudo mv goanime /usr/local/bin/

sudo xattr -d com.apple.quarantine /usr/local/bin/goanime
```

### Linux

<details>
<summary><b>Debian / Ubuntu (e derivados)</b></summary>

```bash
sudo apt update
sudo apt install mpv -y

curl -LO https://github.com/alvarorichard/Goanime/releases/latest/download/goanime-linux-amd64.tar.gz
tar -xzf goanime-linux-amd64.tar.gz
chmod +x goanime-linux-amd64
sudo mv goanime-linux-amd64 /usr/local/bin/goanime
```
</details>

<details>
<summary><b>Arch Linux / Manjaro (AUR)</b></summary>

```bash
yay -S goanime
```
</details>

<details>
<summary><b>Fedora</b></summary>

```bash
sudo dnf update
sudo dnf install mpv

curl -LO https://github.com/alvarorichard/Goanime/releases/latest/download/goanime-linux-amd64.tar.gz
tar -xzf goanime-linux-amd64.tar.gz
chmod +x goanime-linux-amd64
sudo mv goanime-linux-amd64 /usr/local/bin/goanime
```
</details>

### Windows

**Recomendado:** Use o aplicativo de instalação para uma melhor experiência.

1.  Baixe e execute o [Instalador do Windows](https://github.com/alvarorichard/GoAnime/releases/latest/download/GoAnimeInstaller.exe).
2.  Lembre-se também de instalar o `mpv` e adicioná-lo ao PATH do seu sistema.

## Como usar

Siga estes passos recomendados para um uso simples e interativo:

1.  **Abra o terminal.**
2.  **Inicie o aplicativo:** Digite `goanime` e aperte `Enter`.
3.  **Pesquise:** Escreva o nome do anime que deseja assistir.
4.  **Selecione:** Navegue pela lista de resultados com as setas do seu teclado
    e aperte `Enter` para prosseguir.
5.  **Assista:** Escolha o episódio, defina a qualidade e o vídeo será 
    executado imediatamente no reprodutor `mpv`.

## Uso avançado

### Busca direta

Para pesquisar direto da linha de comando, informe um título:

```bash
goanime "Naruto"
```



### Atualizando

Atualize o GoAnime com frequência para receber as novidades:

```bash
goanime --update
```

### Menu de ajuda

```bash
goanime -h
```

## Comunidade e mobile

Entre no nosso Discord para suporte, feedback e trocar ideias:
[Servidor do Discord](https://discord.gg/6nZ2SYv3)

Uma versão mobile do GoAnime está disponível para dispositivos Android:
[GoAnime Mobile](https://github.com/alvarorichard/goanime-mobile)

## Contribuindo

Contribuições são sempre bem-vindas. Antes de iniciar qualquer trabalho, 
certifique-se de ler o nosso [Guia de desenvolvimento](docs/Development.md).

Início rápido:
1.  Faça o fork do projeto.
2.  Crie sua branch a partir de `dev` (`git checkout -b feature/foo`).
3.  Padronize o código usando `go fmt`.
4.  Realize commits no formato Conventional Commits.
5.  Faça o push no github.
6.  Abra um pull request apontando para a branch `dev`.
