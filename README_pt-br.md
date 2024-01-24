<h4 align="center">
    <p>
        <b>Рortuguês</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README.md">English</a>
    </p>
</h4>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="Imagem logo" />
</p>

[![GitHub license](https://img.shields.io/github/license/alvarorichard/GoAnime)](alvarorichard/GoAnime/blob/master/LICENSE) ![GitHub stars](https://img.shields.io/github/stars/alvarorichard/GoAnime) ![GitHub stars](https://img.shields.io/github/languages/count/alvarorichard/ZennityLang) ![GitHub stars](https://img.shields.io/github/languages/top/alvarorichard/GoAnime)  ![GitHub stars](https://img.shields.io/github/last-commit/alvarorichard/GoAnime) ![GitHub stars](https://img.shields.io/github/forks/alvarorichard/GoAnime?style=social)

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

### Instalação Universal (Só precisa do go instalado)
```shell
go install github.com/alvarorichard/GoAnime@latest
```

## Métodos de instalação manual

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
## Instalação somente para Windows

Para instalar o GoAnime no Windows usando o script PowerShell install.ps1, siga estes passos:

1. Abra o PowerShell como Administrador
2.Habilite a Execução de Scripts no PowerShell (se ainda não estiver habilitada):

No PowerShell, execute o seguinte comando para permitir a execução de scripts:

```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```
3. Execute o Script de Instalação:

Execute o script install.ps1:

```powershell
.\install.ps1
```



Lembre-se de adicionar o vlc ao caminho

Use o comando para adicionar o vlc a path:
```shell
set PATH=%PATH%;C:\Program Files\VideoLAN\VLC
```
ou siga este tutorial para adicionar o vlc a path

[Como adicionar o VLC ao PATH](https://www.vlchelp.com/add-vlc-command-prompt-windows/)

## Uso no Linux e macOS

```shell
go-anime
```

## Uso no Windows

```go
goanime
```

O programa solicitará que você insira o nome de um anime. Digite o nome do anime que deseja assistir.

O programa apresentará uma lista de animes que correspondem à sua entrada. Navegue pela lista usando as setas do teclado e pressione enter para selecionar um anime.

Em seguida, o programa apresentará uma lista de episódios do anime selecionado. Novamente, navegue pela lista usando as setas do teclado e pressione enter para selecionar um episódio.

O episódio selecionado será então reproduzido no VLC media player.

# Obrigado
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda) por ajudar e melhorar essa aplicação
