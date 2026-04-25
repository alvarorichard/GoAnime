# Source diagnostics test plan

Este roteiro ajuda a separar indisponibilidade da source de bug no GoAnime.

## Objetivo

- `SourceUnavailable`: 521, 522, 523, 524, 530, DNS error, timeout de conexao ou origem fora devem virar skip no health check.
- `BlockedOrChallenge`: 403, 429, 1020, captcha ou challenge devem virar skip no health check.
- `ParserBroken`: resposta 200 OK sem seletores, JSON ou resultados esperados deve falhar no health check.
- `DecryptBroken`: decrypt/API retornou formato invalido deve falhar no health check.
- `DownloadExpired`: link CDN extraido retornou 403/404 deve ser diagnosticado como link expirado.
- `InternalBug`: panic, nil pointer, loop infinito ou erro local deve falhar.

## Comandos locais

Rode estes comandos antes de abrir ou atualizar a PR:

```powershell
go test ./internal/scraper -count=1 -v
go test ./internal/player -count=1 -v
go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper
$env:CI='true'; go test ./... -count=1
go vet ./...
golangci-lint run --timeout=15m
gosec ./...
govulncheck ./...
git diff --check
```

## Health check live

O teste `TestSourceHealthLive` faz uma busca conhecida por provider:

- Anime/geral: `naruto`
- Filmes/series: `dexter`

Resultado esperado:

- Source offline, Cloudflare 521/522/523/524/530, DNS ou timeout: `t.Skip`.
- Captcha, challenge, 403/429/1020: `t.Skip`.
- 200 OK com parser quebrado ou zero resultados para query conhecida: `t.Fatal`.
- Decrypt quebrado ou erro interno: `t.Fatal`.

## App e logs

Mensagens esperadas:

- `FlixHQ temporariamente indisponivel: Cloudflare 521/origem fora`
- `SFlix bloqueou a requisicao: captcha/challenge`
- `Goyabu respondeu, mas o parser nao encontrou os dados esperados`
- `Download link de download expirou ou foi negado: HTTP 404`

Depois de 3 falhas consecutivas de origem/bloqueio, o circuit breaker pula a source por 10 minutos para evitar martelar servidor fora.

## Discord

O projeto ja possui Discord Rich Presence local, mas isso nao e a mesma coisa que alertas de saude do projeto. Para publicar diagnosticos em um canal do Discord com seguranca, use uma PR separada com:

- `DISCORD_WEBHOOK_URL` configurado como GitHub secret.
- Um job agendado ou manual que rode `go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper`.
- Um passo que envie apenas o resumo de sources `healthy`, `skipped` e `failed`, sem expor tokens, cookies ou URLs privadas.

Sem esse secret configurado, a opcao segura e manter as informacoes no log do CI e no output local.
