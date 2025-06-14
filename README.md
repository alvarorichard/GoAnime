<h4 align="center">
    <p>
        <b>English</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README_pt-br.md">Рortuguês</a>
    </p>
</h4>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="Imagem logo" />
</p>

[![GitHub license](https://img.shields.io/github/license/alvarorichard/GoAnime)
](alvarorichard/GoAnime/blob/master/LICENSE)
![GitHub stars](https://img.shields.io/github/stars/alvarorichard/GoAnime)
![GitHub stars](https://img.shields.io/github/last-commit/alvarorichard/GoAnime)
![GitHub stars](https://img.shields.io/github/forks/alvarorichard/GoAnime?style=social)
[![Build Status](https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml/badge.svg)](https://github.com/alvarorichard/GoAnime/actions)
![GitHub contributors](https://img.shields.io/github/contributors/alvarorichard/GoAnime)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/9923765cb2854ae39af6b567996aad43)](https://app.codacy.com/gh/alvarorichard/GoAnime/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Build Status](https://app.travis-ci.com/alvarorichard/GoAnime.svg?branch=main)](https://app.travis-ci.com/alvarorichard/GoAnime)

# GoAnime

GoAnime is a simple command-line interface (CLI) built in Go, allowing users to search for anime and either play or download episodes directly in Mpv. It scrapes data from websites to provide a selection of anime and episodes to the user, with a special focus and objective on offering animes that are both subtitled and dubbed in Portuguese.

## Features

- Search for anime by name
- Browse episodes
- Skip anime Intro
- Play online with quality selection
- Download single episodes
- Discord RPC about the anime
- Batch download multiple episodes
- Resume playback from where you left off (in builds with SQLite support)
- Track watched episodes (in builds with SQLite support)

> **Note:** GoAnime can be built with or without SQLite support for tracking anime progress.  
> [See the build options documentation](docs/BUILD_OPTIONS.md) for more details.

# Demo

<https://github.com/alvarorichard/GoAnime/assets/88117897/ffec6ad7-6ac1-464d-b048-c80082119836>

## Prerequisites

- Go (at latest version)

- Mpv(at latest version)

## how to install and run

### Universal install (Only needs go installed and recommended for most users)  

```shell
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

### Manual install methods

```shell
git clone https://github.com/alvarorichard/GoAnime.git
```

```shell
cd GoAnime
```

```shell
go run cmd/goanime/main.go
```

## Linux

<details>
<summary>Arch Linux / Manjaro (AUR-based systems)</summary>

Using Yay:

```bash
yay -S goanime
```

or using Paru:

```bash
paru -S goanime
```

Or, to manually clone and install:

```bash
git clone https://aur.archlinux.org/goanime.git
cd goanime
makepkg -si
sudo pacman -S mpv
```

</details>

<details>
<summary>Debian / Ubuntu (and derivatives)</summary>

```bash
sudo apt update
sudo apt install mpv

# For x86_64 systems:
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-linux

chmod +x goanime
sudo mv goanime /usr/bin/
goanime
```

</details>

<details>
<summary>Fedora Installation</summary>

```bash
sudo dnf update
sudo dnf install mpv

# For x86_64 systems:
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-linux

chmod +x goanime
sudo mv goanime /usr/bin/
goanime
```

</details>

<details>
<summary>openSUSE Installation</summary>

```bash
sudo zypper refresh
sudo zypper install mpv

# For x86_64 systems:
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-linux

chmod +x goanime
sudo mv goanime /usr/bin/
goanime
```

</details>

## Windows

<details>
<summary>Windows Installation</summary>

> **Strongly Recommended:** Use the installer for the best experience on Windows.

Option 1: Using the installer (Recommended)

- Download and run the [Windows Installer](https://github.com/alvarorichard/GoAnime/releases/latest/download/GoAnimeInstaller.exe)

Option 2: Standalone executable

- Download the appropriate executable for your system from the [latest release](https://github.com/alvarorichard/GoAnime/releases/latest)

</details>

## macOS

<details>
<summary>macOS Installation</summary>

First, install mpv using Homebrew:

```bash
# Install Homebrew if you haven't already
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install mpv
brew install mpv

# Download and install GoAnime
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-apple-darwin

chmod +x goanime
sudo mv goanime /usr/local/bin/
goanime
```

Alternative installation using MacPorts:

```bash
# Install mpv using MacPorts
sudo port install mpv

# Download and install GoAnime
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-apple-darwin

chmod +x goanime
sudo mv goanime /usr/local/bin/
goanime
```

</details>

### Additional Setup Steps

# NixOS install (Flakes)

## Temporary Run

```shell
nix github:alvarorichard/GoAnime
```

## Install

Add in your `flake.nix`:

```nix
 inputs.goanime.url = "github:alvarorichard/GoAnime";
```

Pass inputs to your modules using ``specialArgs`` and Then in ``configuration.nix``:

```nix
environment.systemPackages = [
  inputs.goanime.packages.${pkgs.system}.GoAnime
];
```

### Usage in Linux and macOS

```go
go-anime
```

### Usage in Windows

```go
goanime
```

### Advanced Usage

You can also use parameters to search for and play anime directly. Here are some examples:

- To search for and play an anime directly, use the following command:

```shell
goanime  "anime name"
```

You can use the `-h` or `--help` option to display help information about how to use the `goanime` command.

```shell
goanime -h
```

The program will prompt you to input the name of an anime. Enter the name of the anime you wish to watch.

 The program will present a list of anime which match your input. Navigate the list using the arrow keys and press enter to select an anime.

The program will then present a list of episodes for the selected anime. Again, navigate the list using the arrow keys and press enter to select an episode.

The selected episode will then play in mpv media player.

# Thanks

[@KitsuneSemCalda](https://github.com/KitsuneSemCalda),[@RushikeshGaikwad](https://github.com/Wraient) and [@the-eduardo](https://github.com/the-eduardo) for help and improve this application

# Alternatives

If you're looking for more options, check out this alternative project by my friend [@KitsuneSemCalda](https://github.com/KitsuneSemCalda) called [Animatic-v2](https://github.com/KitsuneSemCalda/Animatic-v2), which was inspired by GoAnime.

## Contributing

Contributions to improve or enhance are always welcome. Before contributing, please read our comprehensive development guide for detailed information about our workflow, coding standards, and project structure.

📖 **[Development Guide](docs/Development.md)** - Essential reading for contributors

**Quick Start for Contributors:**

1. Fork the Project
2. Read the [Development Guide](docs/Development.md) thoroughly
3. Create your Feature Branch from `dev` (never from `main`)
4. Follow our coding standards (use `go fmt`, add meaningful comments)
5. Ensure all tests pass and add tests for new features
6. Commit your Changes using conventional commit format
7. Push to the Branch
8. Open a Pull Request to the `dev` branch

**Important:** Never commit directly to the `main` branch. All changes must go through the `dev` branch first.
