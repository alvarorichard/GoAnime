<h4 align="center">
    <p>
        <b>Рortuguês</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README.md">English</a>
    </p>
</h4>

<p align="center">
  <img src="https://i.imgur.com/rgkp8OS.png" alt="Imagem logo" />
</p>

# GoAnime
GoAnime é uma interface de linha de comando (CLI) simples construída em Go que permite aos usuários pesquisar um anime e reproduzir episódios diretamente no VLC. Ele raspa dados do site para fornecer uma seleção de anime e episódios ao usuário.

## Pré-requisitos

* Go (na versão mais recente)
* VLC Media Player
* Sqlite3

## Dependências
* PuerkitoBio/goquery
* manifoldco/promptui
* mattn/go-sqlite3
* cavaliergopher/grab/v3
* fzf

## como instalar e executar

```shell
git clone https://github.com/alvarorichard/GoAnime.git
```
```shell
cd GoAnime
sudo bash install.sh
```
para instalar fzf no debian,ArchLinux ou Fedora

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

# Windows
```shell
winget install fzf
```
ou
```shell
choco install fzf
```
ou
```shell
scoop install fzf
```

Lembre-se de adicionar o vlc ao caminho

Use o comando para adicionar o vlc a path
```shell
set PATH=%PATH%;C:\Program Files\VideoLAN\VLC
```
ou siga este tutorial para adicionar o vlc a path

[Como adicionar o VLC ao PATH](https://www.vlchelp.com/add-vlc-command-prompt-windows/)

## Como usar

```shell
go-anime
```

O programa solicitará que você insira o nome de um anime. Digite o nome do anime que deseja assistir.

O programa apresentará uma lista de animes que correspondem à sua entrada. Navegue pela lista usando as setas do teclado e pressione enter para selecionar um anime.

Em seguida, o programa apresentará uma lista de episódios do anime selecionado. Novamente, navegue pela lista usando as setas do teclado e pressione enter para selecionar um episódio.

O episódio selecionado será então reproduzido no VLC media player.

# Obrigado
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda) por ajudar e melhorar essa aplicação