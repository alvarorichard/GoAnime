
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
[![Maintainability](https://api.codeclimate.com/v1/badges/c12b2dc43cc5c0fd3d45/maintainability)](https://codeclimate.com/github/alvarorichard/GoAnime/maintainability)


# GoAnime 
GoAnime is a simple command-line interface (CLI) built in Go, allowing users to search for anime and either play or download episodes directly in Mpv. It scrapes data from websites to provide a selection of anime and episodes to the user, with a special focus and objective on offering animes that are both subtitled and dubbed in Portuguese.




## Prerequisites

* Go (at latest version)



* Mpv(at latest version)
* yt-dlp(at latest version)


## Dependencies
* PuerkitoBio/goquery
* manifoldco/promptui
* cavaliergopher/grab/v3
* ktr0731/go-fuzzyfinder

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
sudo bash install.sh
```

## Arch Linux Install (AUR)

For Arch Linux users, GoAnime is available in the AUR. You can install it using an AUR helper like `paru` or `yay`:
  
Using `paru`:

  ```shell
  paru -S goanime
  ```
Using `yay`:

  ```shell
  yay -S goanime
  ```





# Windows install only
To install GoAnime on Windows using the `install.ps1` PowerShell script, follow these steps:

1. Open PowerShell as Administrator

2. Enable PowerShell Script Execution (if not already enabled):


In the PowerShell window, execute the following command to allow the execution of scripts:

```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

3.Run the Install Script:

Execute the `install.ps1` script:

```powershell
.\install.ps1
```

### Additional Recommendations for Users

For a smoother setup experience, it's recommended to install `mpv` and `yt-dlp` using Scoop as it automatically adds them to your system's PATH. Follow these steps to install these tools:

1. Install Scoop (if not installed):

Open PowerShell and execute:

```powershell
iwr -useb get.scoop.sh | iex
```

2. Install `mpv` and `yt-dlp` using Scoop:

```powershell
scoop install mpv yt-dlp
```
This method ensures that mpv and yt-dlp are added to your PATH automatically, eliminating the need for manual configuration.


Rember add mpv to path

use this command to add mpv to path
```shell
set PATH=%PATH%;C:\Program Files\mpv
```
or follow this tutorial for add mpv to path 

[How to add mpv to path](https://thewiki.moe/tutorials/mpv/)



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

* To search for and play an anime directly, use the following command:
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
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda)   and [@the-eduardo](https://github.com/the-eduardo) for help and improve this application

# Alternatives

If you're looking for more options, check out this alternative project by my friend [@KitsuneSemCalda](https://github.com/KitsuneSemCalda) called [Animatic-v2 ](https://github.com/KitsuneSemCalda/Animatic-v2), which was inspired by GoAnime.

## Contributing

Contributions to improve or enhance are always welcome. Please adhere to the standard pull request process for contributions.


1. Fork the Project
2. Create your Feature Branch
3. Commit your Changes
4. Push to the Branch
5. Open a Pull Request.

