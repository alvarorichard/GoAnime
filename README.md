
<p align="center">
  <img src="https://i.imgur.com/rgkp8OS.png" alt="Imagem logo" />
</p>


# GoAnime 
GoAnime is a simple command-line interface (CLI) built in Go that allows users to search for an anime and play episodes directly in VLC. It scrapes data from the website  provide a selection of anime and episodes to the user.

## Prerequisites

* Go (at latest version)
*  VLC Media Player
* Sqlite3

## Dependencies
* PuerkitoBio/goquery
* manifoldco/promptui
* mattn/go-sqlite3
## how to install and run

```shell
git clone https://github.com/alvarorichard/GoAnime.git
```
```shell
cd GoAnime
sudo bash install.sh
```

## Usage
```go
go-anime
```

The program will prompt you to input the name of an anime. Enter the name of the anime you wish to watch.

 The program will present a list of anime which match your input. Navigate the list using the arrow keys and press enter to select an anime.

The program will then present a list of episodes for the selected anime. Again, navigate the list using the arrow keys and press enter to select an episode.

The selected episode will then play in VLC media player.

## Working in progress

- Fuzzy Search in GoAnime:
    - [x] First Step: Save all animes listed in search 
    - [x] Second Step: Fixing VLC bug in media show
    - [x] Third Step: Working in a fuzzy search
# Thanks 
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda) for help and improve this application
