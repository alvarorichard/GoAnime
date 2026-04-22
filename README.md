<h4 align="center">
    <p>
        <b>English</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README_pt-br.md">Рortuguês</a>
    </p>
</h4>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="GoAnime Logo" />
</p>

<p align="center">
    <a href="alvarorichard/GoAnime/blob/master/LICENSE"><img src="https://img.shields.io/github/license/alvarorichard/GoAnime" alt="GitHub license"></a>
    <img src="https://img.shields.io/github/stars/alvarorichard/GoAnime" alt="GitHub stars">
    <img src="https://img.shields.io/github/last-commit/alvarorichard/GoAnime" alt="GitHub last commit">
    <img src="https://img.shields.io/github/forks/alvarorichard/GoAnime?style=social" alt="GitHub forks">
    <a href="https://github.com/alvarorichard/GoAnime/actions"><img src="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
    <img src="https://img.shields.io/github/contributors/alvarorichard/GoAnime" alt="GitHub contributors">
    <a href="https://discord.gg/FbQuf78D9G"><img src="https://img.shields.io/badge/Discord-Community-7289DA?logo=discord&logoColor=white" alt="Discord"></a>
</p>

# GoAnime

GoAnime is a simple text-based user interface (TUI) built in Go, allowing users
to search for anime and either play or download episodes directly in mpv. It
scrapes data from websites to provide a selection of anime and episodes, with
support for both subbed and dubbed content in English and Portuguese.

## Table of contents

1.  [Features](#features)
2.  [Prerequisites](#prerequisites)
3.  [Installation](#installation)
4.  [How to use](#how-to-use)
5.  [Advanced usage](#advanced-usage)
6.  [Community and mobile](#community-and-mobile)
7.  [Contributing](#contributing)

## Features

*   Search for anime, movies, and TV shows by name
*   Simultaneous multi-source searching by default across all active platforms
*   Support for subbed and dubbed content in English and Portuguese
*   Play online with quality selection or download episodes
*   Discord RPC integration to show what you're watching
*   Progress tracking: Resume playback and track watched episodes

*   Built-in upscaling (Anime4K) for better video quality

## Prerequisites

Before installing GoAnime, ensure you have the following dependency installed:
*   [mpv](https://mpv.io/) (Media player, latest version recommended)

## Installation

Choose the installation method that best fits your system.

### Universal installation

If you have Go installed on your system, you can install GoAnime via `go install`:

```bash
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

### macOS

Install `mpv` using Homebrew, then download and configure GoAnime:

```bash
brew install mpv

curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-apple-darwin
chmod +x goanime
sudo mv goanime /usr/local/bin/

sudo xattr -d com.apple.quarantine /usr/local/bin/goanime
```

### Linux

<details>
<summary><b>Debian / Ubuntu (and derivatives)</b></summary>

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

**Strongly Recommended:** Use our installer for the best experience.

1.  Download and run the [GoAnime Windows Installer](https://github.com/alvarorichard/GoAnime/releases/latest/download/GoAnimeInstaller.exe).
2.  Install `mpv` for Windows and ensure it is available in your system's path.

## How to use

Follow these steps for a simple, interactive watching experience:

1.  **Open your terminal.**
2.  **Start the app:** Type `goanime` and press `Enter`.
3.  **Search:** Provide the name of the anime you want to watch.
4.  **Select:** Navigate the resulting list using your arrow keys and press 
    `Enter` to pick an anime.
5.  **Watch:** Select an episode, choose your preferred streaming quality, and 
    the video will automatically launch in `mpv`.

## Advanced usage

### Direct search

To bypass the initial prompt, directly pass the anime name:

```bash
goanime "Naruto"
```



### Updating the app

Keep GoAnime updated to the newest features without manual downloads:

```bash
goanime --update
```

### Help

To view all available commands and flags:

```bash
goanime -h
```

## Community and mobile

Join our Discord for support, feedback, and updates:
[Join the Discord server](https://discord.gg/6nZ2SYv3)

A mobile version of GoAnime is also available for Android devices:
[GoAnime Mobile](https://github.com/alvarorichard/goanime-mobile)

## Contributing

Contributions to improve or enhance are always welcome.

See the [development guide](docs/Development.md).

Quick start:
1.  Fork the project and read the development guide.
2.  Create your feature branch from `dev` (`git checkout -b feature/foo`).
3.  Follow formatting standards (`go fmt`).
4.  Commit your changes (`git commit -m 'feat: add foo'`).
5.  Push to the branch (`git push origin feature/foo`).
6.  Open a pull request to the `dev` branch.

All changes must go through the `dev` branch first.
