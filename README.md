
<h4 align="center">
    <p>
        <b>English</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README_pt-br.md">Рortuguês</a>
    </p>
</h4>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="Imagem logo" />
</p>

[![GitHub license](https://img.shields.io/github/license/alvarorichard/GoAnime)](alvarorichard/GoAnime/blob/master/LICENSE) ![GitHub stars](https://img.shields.io/github/stars/alvarorichard/GoAnime) ![GitHub stars](https://img.shields.io/github/languages/count/alvarorichard/ZennityLang) ![GitHub stars](https://img.shields.io/github/languages/top/alvarorichard/GoAnime)  ![GitHub stars](https://img.shields.io/github/last-commit/alvarorichard/GoAnime) ![GitHub stars](https://img.shields.io/github/forks/alvarorichard/GoAnime?style=social)


# GoAnime 
GoAnime is a simple command-line interface (CLI) built in Go that allows users to search for an anime and play episodes directly in VLC. It scrapes data from  website to provide a selection of anime and episodes to the user, with a special focus and objective on offering animes that are both subtitled and dubbed in Portuguese.

## Prerequisites

* Go (at latest version)
*  VLC Media Player
* Sqlite3

## Dependencies
* PuerkitoBio/goquery
* manifoldco/promptui
* mattn/go-sqlite3
* cavaliergopher/grab/v3
* fzf 
## how to install and run

### Universal install (Only needs go installed  
```shell
go install github.com/alvarorichard/Goanime@latest
```

### Manual install methods
```shell
git clone https://github.com/alvarorichard/GoAnime.git
```
```shell
cd GoAnime
sudo bash install.sh
```
to install fzf in debian,ArchLinux or Fedora

debian :
```shell
sudo apt install fzf
```
ArchLinux :

```shell
sudo pacman -S fzf
```

Fedora: 

```shell
sudo dnf install fzf
```

macOS:
```shell
brew install fzf
```

### Windows 
```shell
winget install fzf
```
or 
```shell
choco install fzf
```
or
```shell
scoop install fzf
```

# Windows install only
To install GoAnime on Windows using the `install.ps1` PowerShell script, follow these steps:

1. Open PowerShell as Administrator

2.Enable PowerShell Script Execution (if not already enabled):


In the PowerShell window, execute the following command to allow the execution of scripts:

```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

3. Run the Install Script:

Execute the `install.ps1` script:

```powershell
.\install.ps1
```









Rember add vlc to path

use this command to add vlc to path
```shell
set PATH=%PATH%;C:\Program Files\VideoLAN\VLC
```
or follow this tutorial for add vlc to path 

[How to add vlc to path](https://www.vlchelp.com/add-vlc-command-prompt-windows/)



### Usage in Linux and macOS
```go
go-anime
```

### Usage in Windows

```go
goanime
```



The program will prompt you to input the name of an anime. Enter the name of the anime you wish to watch.

 The program will present a list of anime which match your input. Navigate the list using the arrow keys and press enter to select an anime.

The program will then present a list of episodes for the selected anime. Again, navigate the list using the arrow keys and press enter to select an episode.

The selected episode will then play in VLC media player.


# Thanks 
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda) for help and improve this application
