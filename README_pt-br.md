<h4 align="center">
    <p>
        <b>Рortuguês</b> |
        <a href="https://github.com/alvarorichard/GoAnime/blob/main/README.md">English</a>
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
GoAnime é uma interface de linha de comando (CLI) simples, desenvolvida em Go, que permite aos usuários procurar um anime e reproduzir ou baixar episódios diretamente no MPV. Ele coleta dados de sites para fornecer uma seleção de animes e episódios ao usuário, com um foco e objetivo especial em oferecer animes que são legendados e dublados em português


# Demo 
https://github.com/alvarorichard/GoAnime/assets/88117897/ffec6ad7-6ac1-464d-b048-c80082119836



## Pré-requisitos

* Go (na versão mais recente)
* mpv (na versão mais recente)

## Dependências
* PuerkitoBio/goquery
* manifoldco/promptui
* cavaliergopher/grab/v3
* ktr0731/go-fuzzyfinder


### Instalação Universal (Só precisa do go instalado e recomendado para a maioria dos usuarios)
```shell
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

## Métodos de instalação manual

```shell
git clone https://github.com/alvarorichard/GoAnime.git
```
```shell
cd GoAnime
sudo bash install.sh
```
## Instalação no Arch Linux (AUR)

Para usuários do Arch Linux, o GoAnime está disponível no AUR. Você pode instalá-lo usando um auxiliar de AUR como `paru` ou `yay`:

Usando `paru`:

  ```shell
  paru -S goanime
  ```
Usando `yay`:

  ```shell
  yay -S goanime
  ```


# Instalação apenas para Windows
Para instalar o GoAnime no Windows usando o script PowerShell `install.ps1`, siga estes passos:

1. Abra o PowerShell como Administrador.
2. Habilite a Execução de Scripts no PowerShell (se ainda não estiver habilitada):
```shell
   Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

3. Execute o Script de Instalação:
```shell
   .\install.ps1
```
### Recomendações Adicionais para os Usuários

Para uma experiência de configuração mais suave, é recomendado instalar `mpv` e `yt-dlp` usando o Scoop, pois isso os adiciona automaticamente ao PATH do seu sistema. Siga estes passos para instalar essas ferramentas:

 Instale o Scoop (se não estiver instalado):
  ```shell
  Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
  ```
 Instale `mpv` e `yt-dlp` usando o Scoop:
   ```shell
   scoop install mpv yt-dlp
   ``` 
Este método garante que mpv e yt-dlp sejam adicionados ao seu PATH automaticamente, eliminando a necessidade de configuração manual.

Lembre-se de adicionar mpv ao PATH usando este comando:
   
   ```shell
   set PATH=%PATH%;C:\Program Files\mpv
   ```
Ou siga este tutorial para adicionar mpv ao PATH:

[Como adicionar mpv ao PATH](https://thewiki.moe/tutorials/mpv/)


## Uso no Linux e macOS

```shell
go-anime
```

## Uso no Windows

```go
goanime
```

### Uso Avançado
Você também pode usar parâmetros para procurar e reproduzir anime diretamente. Aqui estão alguns exemplos:

* Para procurar e reproduzir um anime diretamente, use o seguinte comando:
```shell
goanime  "nome do anime"
```
Você pode usar a opção `-h` ou `--help` para exibir informações de ajuda sobre como usar o comando goanime.

```shell
goanime -h
```



O programa solicitará que você insira o nome de um anime. Digite o nome do anime que deseja assistir.

O programa apresentará uma lista de animes que correspondem à sua entrada. Navegue pela lista usando as setas do teclado e pressione enter para selecionar um anime.

Em seguida, o programa apresentará uma lista de episódios do anime selecionado. Novamente, navegue pela lista usando as setas do teclado e pressione enter para selecionar um episódio.

O episódio selecionado será então reproduzido no MPV

# Agradecimento
[@KitsuneSemCalda](https://github.com/KitsuneSemCalda)  e [@the-eduardo](https://github.com/the-eduardo) por ajudar e melhorar essa aplicação


# Alternativas

Se você estiver procurando por mais opções, confira este projeto alternativo do meu amigo [@KitsuneSemCalda](https://github.com/KitsuneSemCalda) chamado [Animatic-v2](https://github.com/KitsuneSemCalda/Animatic-v2), que foi inspirado no GoAnime.

## Contribuindo

Contribuições para melhorar ou aprimorar são sempre bem-vindas. Por favor, siga o processo padrão de pull request para contribuições.

1. Faça um fork do projeto
2. Crie sua branch de funcionalidade
3. Faça commit das suas alterações
4. Faça push para a sua branch
5. Abra um pull request.
