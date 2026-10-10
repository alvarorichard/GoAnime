<p align="center">
  <a href="README.md">English</a> | <b>Português</b>
</p>

<p align="center">
  <img src="https://github.com/alvarorichard/GoAnime/assets/102667323/49600255-d5a2-4405-81d1-a08cebae569a" alt="GoAnime" />
</p>

<p align="center">
  <a href="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml"><img src="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/alvarorichard/GoAnime/releases/latest"><img src="https://img.shields.io/github/v/release/alvarorichard/GoAnime" alt="Última versão"></a>
  <a href="https://aur.archlinux.org/packages/goanime"><img src="https://img.shields.io/aur/version/goanime" alt="AUR"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/alvarorichard/GoAnime" alt="Licença"></a>
</p>

O GoAnime é um aplicativo de terminal para buscar, assistir e baixar animes,
filmes e séries em português e inglês. Ele pesquisa em várias fontes ao mesmo
tempo, reproduz o episódio escolhido no [mpv](https://mpv.io/) e guarda o
ponto em que você parou.

<p align="center">
  <img src="docs/demo/demo.gif" alt="Buscando um título no GoAnime, escolhendo nos resultados e assistindo" width="800">
</p>

É um único binário para Linux, macOS e Windows, e só precisa do mpv para
funcionar.

## Conteúdo

- [Recursos](#recursos)
- [Instalação](#instalação)
- [Uso](#uso)
- [Fontes](#fontes)
- [Configuração](#configuração)
- [Solução de problemas](#solução-de-problemas)
- [Usando o GoAnime como biblioteca](#usando-o-goanime-como-biblioteca)
- [Contribuindo](#contribuindo)
- [Comunidade](#comunidade)
- [Aviso legal](#aviso-legal)
- [Licença](#licença)

## Recursos

- Pesquisa em todas as fontes ativas em paralelo e passa para outra fonte
  quando uma falha.
- Animes legendados e dublados em português e inglês, além de filmes e séries
  em português.
- Reprodução com escolha de qualidade, ou download de um episódio, de um
  intervalo ou da série inteira.
- Downloads salvos com nomes compatíveis com o Plex (`Anime - S01E01.mp4`).
- Retoma a reprodução e lembra os episódios assistidos.
- Pula aberturas e encerramentos automaticamente.
- Discord Rich Presence.
- Upscaling com Anime4K para vídeos e imagens baixados.
- Atualiza a si mesmo com `goanime --update`.

## Instalação

O GoAnime usa o [mpv](https://mpv.io/) para reproduzir vídeos. Instale o mpv
antes, a menos que use o instalador do Windows, que já o inclui.

Há binários prontos para todas as plataformas na
[página de releases](https://github.com/alvarorichard/GoAnime/releases/latest).

### Windows

Baixe e execute o `GoAnime-Installer-<versão>.exe` da
[última release](https://github.com/alvarorichard/GoAnime/releases/latest).
Ele instala o GoAnime e o mpv e adiciona os dois ao `PATH`.

Se preferir não usar o instalador, baixe o `goanime-windows-amd64.zip`,
extraia e coloque o `goanime.exe` e o `mpv.exe` em uma pasta do seu `PATH`.

### macOS

```bash
brew install mpv
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-darwin-universal
chmod +x goanime
sudo mv goanime /usr/local/bin/
sudo xattr -d com.apple.quarantine /usr/local/bin/goanime
```

O binário é universal e roda tanto em Macs com Apple Silicon quanto com Intel.

### Arch Linux

O GoAnime está disponível no [AUR](https://aur.archlinux.org/packages/goanime):

```bash
yay -S goanime
```

### Debian, Ubuntu, Fedora e outras distribuições

Instale o mpv pelo gerenciador de pacotes (`sudo apt install mpv` ou
`sudo dnf install mpv`) e depois:

```bash
curl -LO https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-linux-amd64.tar.gz
tar -xzf goanime-linux-amd64.tar.gz
sudo install goanime-linux-amd64 /usr/local/bin/goanime
```

Em ARM64, troque `amd64` por `arm64`.

### A partir do código-fonte

Com Go 1.27 ou mais recente:

```bash
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

O histórico de progresso usa SQLite e depende de CGO, então é preciso ter um
compilador C disponível na hora do build. Sem ele, o GoAnime compila e
funciona normalmente, mas não registra o progresso. Os binários das releases
já vêm com SQLite. Mais detalhes em
[docs/BUILD_OPTIONS.md](docs/BUILD_OPTIONS.md).

## Uso

Rode `goanime` sem argumentos para abrir a busca interativa. Digite um nome,
escolha um resultado com as setas, selecione o episódio e ele abre no mpv.
Também dá para passar o nome direto:

```bash
goanime "one piece"
```

Use espaços no nome, não hífens.

Os downloads vão para `~/.local/goanime/downloads/anime/`, a menos que você
use `-o`:

```bash
goanime -d "one piece" 1                  # episódio 1
goanime -d -r "naruto" 1-12               # episódios 1 a 12
goanime -d -a "one piece"                 # todos os episódios
goanime -d --quality 720p "frieren" 1     # uma qualidade específica
goanime -d -o ~/Anime "bleach" 10         # outra pasta
```

Filmes e séries usam `-dm`:

```bash
goanime -dm "Interestelar"                # um filme
goanime -dm -r "Dark" 1 1-5               # temporada 1, episódios 1 a 5
goanime -dm -a "Dark"                     # todas as temporadas
```

Outras opções:

```bash
goanime --source animefire "jujutsu kaisen"   # buscar em uma única fonte
goanime --upscale --upscale-hq video.mp4      # upscale de um arquivo com Anime4K
goanime --update                              # atualizar para a última versão
goanime --debug "naruto"                      # gerar um log de depuração
```

Rode `goanime -h` para ver todas as opções.

## Fontes

| Fonte     | Conteúdo         | Idioma                        | Observações                                   |
| --------- | ---------------- | ----------------------------- | --------------------------------------------- |
| HiAnime   | Animes           | Inglês, legendado e dublado   | Legendado por padrão                          |
| AnimeFire | Animes           | Português (Brasil)            |                                               |
| Goyabu    | Animes           | Português (Brasil)            | Desativada por padrão, veja abaixo            |
| StartFlix | Filmes, séries   | Português (Brasil)            | Dublado e legendado                           |
| TopCine   | Filmes, séries   | Português (Brasil)            | Apenas catálogo; a reprodução usa o StartFlix |

O Goyabu fica atrás de uma verificação do Cloudflare e vem desativado. Para
ativá-lo, use `GOANIME_ENABLED_SOURCES=goyabu`.

O GoAnime depende de sites de terceiros, e esses sites mudam sem aviso. Quando
isso acontece, a fonte correspondente pode parar de funcionar até ser
atualizada. Um [workflow agendado](.github/workflows/source-health.yml) testa
todas as fontes diariamente, então a falha costuma aparecer lá antes de ser
reportada.

## Configuração

O GoAnime não precisa de arquivo de configuração. Algumas variáveis de
ambiente cobrem casos menos comuns:

| Variável                        | Descrição                                                         |
| ------------------------------- | ----------------------------------------------------------------- |
| `GOANIME_ENABLED_SOURCES`       | Fontes a ativar, separadas por vírgula, como `goyabu`             |
| `GOANIME_DISABLED_SOURCES`      | Fontes a ignorar, separadas por vírgula                           |
| `GOANIME_HIANIME_AUDIO`         | `sub` (padrão) ou `dub`                                           |
| `GOANIME_STRICT_SOURCE`         | `1` para mostrar um erro em vez de adivinhar uma fonte desconhecida |
| `SSL_CERT_FILE`, `SSL_CERT_DIR` | Usar um bundle de CA próprio, veja [Solução de problemas](#solução-de-problemas) |

## Solução de problemas

### Reportando um problema

Rode o mesmo comando de novo com `--debug`. Ao iniciar, o GoAnime mostra o
caminho de um arquivo de log. Anexe esse arquivo à sua
[issue](https://github.com/alvarorichard/GoAnime/issues/new), junto com o
comando que você usou e o seu sistema operacional.

### O mpv não é encontrado

Confira se `mpv --version` funciona no mesmo terminal em que você roda o
GoAnime. No Windows, abra um terminal novo depois de instalar, para que o
`PATH` atualizado seja carregado.

### Erros de TLS atrás de proxy corporativo ou CA própria

Se todas as requisições falham com erro de certificado, normalmente em uma
rede que inspeciona TLS com uma CA raiz própria, aponte o GoAnime para esse
bundle:

```bash
export SSL_CERT_FILE=/caminho/para/ca-corporativa.pem   # um único bundle PEM
export SSL_CERT_DIR=/caminho/para/ca-certificates.d      # ou um diretório de PEMs
```

Desde o Go 1.27 essas variáveis também valem no Windows e no macOS, não só no
Linux: quando uma delas está definida, o GoAnime verifica os certificados com
o verificador do próprio Go usando o seu bundle, em vez do repositório de
certificados do sistema. Para voltar a usar o repositório do sistema sem
remover as variáveis, rode com `GODEBUG=x509sslcertoverrideplatform=0`.

## Usando o GoAnime como biblioteca

O código de busca e scraping é exposto como um pacote Go público:

```go
client := goanime.NewClient()
results, err := client.SearchAnime("One Piece", nil)
```

Veja [pkg/goanime](pkg/goanime) para a API e
[pkg/goanime/examples](pkg/goanime/examples) para programas completos.

## Contribuindo

Relatos de bugs, correções e novas fontes são bem-vindos.

A maior parte da manutenção do GoAnime é manter as fontes funcionando conforme
os sites mudam, e esse também é o jeito mais fácil de começar a contribuir: o
[workflow de saúde das fontes](https://github.com/alvarorichard/GoAnime/actions/workflows/source-health.yml)
mostra quais estão falhando, e cada fonte é um pacote independente em
`internal/scraper/providers`. Para adicionar uma nova, siga o
[docs/ADDING_A_SOURCE.md](docs/ADDING_A_SOURCE.md), que usa o Goyabu como
implementação de referência.

Para compilar e testar:

```bash
git clone https://github.com/alvarorichard/GoAnime.git
cd GoAnime
git checkout dev
go build ./cmd/goanime
go test -short ./...
```

O desenvolvimento acontece na branch `dev`. Crie sua branch a partir da `dev`
e abra os pull requests para a `dev`; a `main` só recebe releases. Estilo de
código, linters e o restante do fluxo estão descritos em
[docs/Development.md](docs/Development.md).

## Comunidade

- [Discord](https://discord.gg/FbQuf78D9G) para dúvidas, sugestões e anúncios
  de novas versões.
- [GoAnime Mobile](https://github.com/alvarorichard/goanime-mobile), a versão
  para Android.
- [goanime.online](https://goanime.online/)

## Aviso legal

O GoAnime não hospeda, envia nem distribui nenhuma mídia. Ele acessa conteúdo
que sites de terceiros disponibilizam publicamente e não tem vínculo com
nenhum deles. Cada usuário é responsável por cumprir as leis do seu país.

## Licença

O GoAnime é distribuído sob a [Licença MIT](LICENSE).
