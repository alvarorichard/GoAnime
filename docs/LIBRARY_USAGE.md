# GoAnime Library Integration Guide

## 📦 O que foi criado

Foi criada uma estrutura completa em `pkg/goanime` que expõe as funcionalidades de scraping e busca do GoAnime como uma biblioteca pública para ser usada em outros projetos Go.

## 🎯 Estrutura Criada

```
pkg/
├── README.md                                 # Documentação principal da biblioteca
├── PACKAGE_INFO.md                          # Informações sobre a estrutura do pacote
└── goanime/                                 # Pacote principal
    ├── client.go                            # Cliente principal da API
    ├── client_test.go                       # Testes unitários e de integração
    ├── doc.go                               # Documentação do pacote
    ├── README.md                            # Guia completo de uso
    ├── types/                               # Tipos públicos
    │   ├── anime.go                         # Tipos Anime, Episode, etc.
    │   └── source.go                        # Enum Source e helpers
    └── examples/                            # Exemplos de uso
        ├── search/main.go                   # Exemplo: busca básica
        ├── episodes/main.go                 # Exemplo: listar episódios
        ├── stream/main.go                   # Exemplo: obter URL de stream
        └── source_specific/main.go          # Exemplo: busca em fonte específica
```

## 🚀 Como Usar em Outros Projetos

### 1. Instalação

```bash
go get github.com/alvarorichard/Goanime
```

### 2. Uso Básico

```go
package main

import (
    "fmt"
    "log"
    "github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
    // Criar cliente
    client := goanime.NewClient()
    
    // Buscar anime
    results, err := client.SearchAnime("Naruto", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Exibir resultados
    for _, anime := range results {
        fmt.Printf("%s [%s]\n", anime.Name, anime.Source)
    }
}
```

### 3. Busca em Fonte Específica

```go
import "github.com/alvarorichard/Goanime/pkg/goanime/types"

client := goanime.NewClient()

// Buscar apenas no AllAnime
source := types.SourceAllAnime
results, err := client.SearchAnime("One Piece", &source)
```

### 4. Obter Episódios

```go
// Após buscar um anime...
source, _ := types.ParseSource(anime.Source)
episodes, err := client.GetAnimeEpisodes(anime.URL, source)

for _, ep := range episodes {
    fmt.Printf("Episódio %s: %s\n", ep.Number, ep.Title.English)
}
```

### 5. Obter URL de Stream

```go
// Após obter episódios...
streamURL, headers, err := client.GetStreamURL(episode.URL, source)

fmt.Printf("URL: %s\n", streamURL)
for key, value := range headers {
    fmt.Printf("Header %s: %s\n", key, value)
}
```

## 📚 API Disponível

### Client

- **`NewClient()`** - Cria um novo cliente
- **`SearchAnime(query, source)`** - Busca anime por nome
- **`GetAnimeEpisodes(animeURL, source)`** - Obtém episódios de um anime
- **`GetStreamURL(episodeURL, source)`** - Obtém URL de streaming
- **`GetAvailableSources()`** - Lista fontes disponíveis

### Types

#### `types.Anime`
- `Name` - Nome do anime
- `URL` - URL do anime na fonte
- `ImageURL` - URL da imagem de capa
- `Episodes` - Lista de episódios
- `AnilistID` - ID do AniList
- `MalID` - ID do MyAnimeList
- `Source` - Nome da fonte
- `Details` - Metadados estendidos

#### `types.Episode`
- `Number` - Número do episódio
- `URL` - URL do episódio
- `Title` - Título do episódio
- `Duration` - Duração em segundos
- `IsFiller` - Se é episódio filler
- `IsRecap` - Se é episódio recap
- `SkipTimes` - Timestamps para pular OP/ED

#### `types.Source`
- `SourceAllAnime` - Fonte AllAnime
- `SourceAnimeFire` - Fonte AnimeFire

## 🧪 Testes

```bash
# Executar todos os testes
go test ./pkg/goanime/...

# Apenas testes unitários (sem integração)
go test -short ./pkg/goanime/...

# Com verbose
go test -v ./pkg/goanime/...
```

**Resultado:** ✅ Todos os testes passando

## 🔨 Compilar Exemplos

```bash
# Exemplo de busca
go build -o search ./pkg/goanime/examples/search/

# Exemplo de episódios
go build -o episodes ./pkg/goanime/examples/episodes/

# Exemplo de stream
go build -o stream ./pkg/goanime/examples/stream/

# Exemplo de fonte específica
go build -o source ./pkg/goanime/examples/source_specific/
```

## ✅ Verificações

- ✅ Código compila sem erros
- ✅ Todos os testes passam
- ✅ Linting sem problemas (`golangci-lint run ./pkg/...`)
- ✅ Exemplos funcionais
- ✅ Documentação completa
- ✅ API type-safe

## 📖 Documentação

1. **[pkg/README.md](pkg/README.md)** - Visão geral e início rápido
2. **[pkg/goanime/README.md](pkg/goanime/README.md)** - Documentação detalhada da API
3. **[pkg/goanime/examples/](pkg/goanime/examples/)** - Exemplos práticos de uso
4. **[pkg/PACKAGE_INFO.md](pkg/PACKAGE_INFO.md)** - Informações sobre o pacote

## 🎓 Exemplos de Integração

### Integração com MPV

```go
streamURL, headers, _ := client.GetStreamURL(episode.URL, source)

args := []string{streamURL}
for key, value := range headers {
    args = append(args, fmt.Sprintf("--http-header-fields=%s: %s", key, value))
}

cmd := exec.Command("mpv", args...)
cmd.Run()
```

### Integração com HTTP Client

```go
streamURL, headers, _ := client.GetStreamURL(episode.URL, source)

req, _ := http.NewRequest("GET", streamURL, nil)
for key, value := range headers {
    req.Header.Set(key, value)
}

resp, _ := http.DefaultClient.Do(req)
defer resp.Body.Close()
```

### Filtrar Episódios Filler

```go
episodes, _ := client.GetAnimeEpisodes(anime.URL, source)

mainEpisodes := make([]*types.Episode, 0)
for _, ep := range episodes {
    if !ep.IsFiller {
        mainEpisodes = append(mainEpisodes, ep)
    }
}
```

## 🔒 Segurança

- Todas as funções exportadas são seguras para uso concorrente
- URLs são validadas antes do uso
- Tratamento de erros apropriado em todas as operações
- Headers HTTP são tratados de forma segura

## 📝 Notas Importantes

1. **URLs de Stream expiram** - Obtenha novamente quando necessário
2. **Rate Limiting** - A biblioteca já lida com isso automaticamente
3. **Headers** - Alguns streams requerem headers específicos (retornados por `GetStreamURL`)
4. **Metadados** - Nem todos os animes têm todos os metadados disponíveis

## 🤝 Contribuindo

Ao contribuir com a biblioteca pública:

1. Mantenha a API simples e limpa
2. Mantenha compatibilidade retroativa
3. Adicione testes para novos recursos
4. Atualize a documentação
5. Siga as convenções Go

## 📄 Licença

MIT License - veja [LICENSE](../LICENSE)

## 🔗 Links Úteis

- Repositório principal: https://github.com/alvarorichard/GoAnime
- Documentação completa: [pkg/goanime/README.md](pkg/goanime/README.md)
- Exemplos: [pkg/goanime/examples/](pkg/goanime/examples/)
- Issues: https://github.com/alvarorichard/GoAnime/issues

---

**Criado em:** 19 de Novembro de 2025  
**Status:** ✅ Pronto para uso  
**Versão:** 1.0.0
