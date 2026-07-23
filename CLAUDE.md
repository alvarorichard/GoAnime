# CLAUDE.md — Instruções de Execução de Testes GoAnime

> **LEIA ISTO PRIMEIRO.** Instruções obrigatórias para execução de testes por fases.
> **Docs:** `TEST_PLAN.md`, `TEST_STRATEGY.md`, `TEST_STAGES.md`, `TEST_PLAN_FUNCTIONS.md`

---

## REGRA ABSOLUTA — 1 TESTE POR FUNÇÃO

**CADA função listada em `TEST_PLAN_FUNCTIONS.md` DEVE ter seu próprio teste dedicado.**
Isso é um requisito obrigatório de qualidade. Não agrupe múltiplas funções em um único teste.
Cada `func X()` no código → gera um `func TestX()` no arquivo de teste.

Exceções permitidas (apenas estas):
- `main()` de exemplos e cmd
- Funções que requerem hardware (GPU para FFmpeg)
- Funções de TUI interativo (`Find`, `getUserInput`) → testar lógica interna indiretamente

---

## REGRA #2 — UMA FASE POR SESSÃO, MÁXIMO POSSÍVEL

1. Abra `TEST_STAGES.md`, encontre a próxima fase ⬜
2. Execute a fase INTEIRA — sem limites artificiais de arquivos ou testes
3. Se a fase for grande demais para uma sessão, faça o máximo e marque como 🔄
4. A próxima sessão continua de onde parou

---

## WORKFLOW POR FASE

### Passo 1: Identificar a fase
- Abrir `TEST_STAGES.md` → primeira fase ⬜
- Anunciar: "Executando FASE X — [nome]"

### Passo 2: Ler código-fonte
- Abrir cada arquivo `.go` listado na fase
- Ler assinaturas, tipos de retorno, dependências
- Identificar se a função é exportada ou não

### Passo 3: Verificar testes existentes
- Checar se existe `*_test.go` no pacote
- Reutilizar helpers/mocks/fixtures existentes
- **NÃO duplicar** mocks

### Passo 4: Escrever testes — REGRAS OBRIGATÓRIAS

```
1. UM teste por função — TestNomeDaFuncao_Cenario
2. Table-driven com subcases para cobrir múltiplos cenários DA MESMA função
3. t.Parallel() em TODOS (exceto globals)
4. t.Helper() em funções auxiliares de teste
5. require para precondições, assert para verificações
6. testify: "github.com/stretchr/testify/assert" e "require"
7. httptest.Server para TODO mock HTTP — NUNCA rede real
8. t.TempDir() para arquivos temporários
9. t.Cleanup() para liberar recursos
```

### Passo 5: Rodar e verificar
```bash
go test ./PACOTE/ -v -race -count=1
go test ./PACOTE/ -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1
```

### Passo 6: Atualizar `TEST_STAGES.md`
- Mudar ⬜ → ✅ (completa) ou 🔄 (parcial)

### Passo 7: Relatório final da sessão
```
═══════════════════════════════════════
FASE X — [NOME]
Testes criados: N
Testes passando: N/N
Funções cobertas: N/N da fase
Cobertura: XX.X% → XX.X%
Próxima: FASE Y — [nome]
═══════════════════════════════════════
```

---

## PADRÕES DE TESTE

### Unitário Puro
```go
func TestFuncName(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name string
        input TYPE
        want  TYPE
    }{
        {"normal", input1, expected1},
        {"edge", input2, expected2},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got := FuncName(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Com Mock HTTP
```go
func TestFuncName(t *testing.T) {
    t.Parallel()
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, `{"status":"ok"}`)
    }))
    t.Cleanup(srv.Close)
    client := NewClient()
    client.baseURL = srv.URL
    result, err := client.DoSomething()
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Lógica de Estado
```go
func TestComponent_ConcurrentAccess(t *testing.T) {
    t.Parallel()
    comp := NewComponent()
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() { defer wg.Done(); comp.DoOp() }()
    }
    wg.Wait()
}
```

---

## MOCKS EXISTENTES — REUTILIZAR

| Mock | Localização |
|---|---|
| `MockScraper` | `internal/scraper/unified_test.go` |
| `mockHLSCDN` | `internal/downloader/hls/download_test.go` |
| `mockCDN` | `internal/downloader/hls/hls_test.go` |
| `createTestManager` | `internal/scraper/unified_test.go` |

---

## FIXTURES (Fases 7-10)
```go
func loadFixture(t *testing.T, path string) string {
    t.Helper()
    data, err := os.ReadFile(path)
    require.NoError(t, err)
    return string(data)
}
```
Criar em `internal/scraper/testdata/SCRAPER_NAME/`

---

## ERROS A EVITAR

| Erro | Correção |
|---|---|
| HTTP real em teste | `httptest.Server` SEMPRE |
| Singleton pollution | `t.Cleanup()` para resetar |
| Path relativa | `filepath.Join("testdata", ...)` |
| Depender de ordem | `t.Parallel()` + dados independentes |
| Esquecer `-race` | SEMPRE rodar com `-race` |
| Pular função sem teste | **PROIBIDO** — cada função DEVE ter teste |

---

## COMO O USUÁRIO VAI PEDIR

- "Continue os testes" → próxima fase ⬜
- "Fase X" → fase específica
- "Status" → checklist de `TEST_STAGES.md`
- "Cobertura" → `go test ./... -short -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1`

---

## VERIFICAÇÃO FINAL (após FASE 18)
```bash
go test ./... -coverprofile=coverage.out -covermode=atomic -race
go tool cover -func=coverage.out | tail -1
go tool cover -func=coverage.out | awk '$NF == "0.0%"' | grep -v "examples\|cmd/goanime" | wc -l
```
Meta: **≥ 70.0%** · **≤ 26 funções a 0%** (apenas TUI/IPC/main/hardware não-testáveis)

---

## HISTÓRICO DE COBERTURA (2026-05-24) — META ATINGIDA

- ✅ FASES 1–18: **TODAS COMPLETAS** → **75.7% cobertura total** (full suite com -race) · **2 funções non-main a 0%**
- 2 restantes = `tui.Find` (TTY fuzzyfinder) + `upscaler.Close` (cover-tool quirk) — ambas intratáveis sem TTY/hardware

| Fase | Pacotes | Funcs alvo | Stmts | Status |
|---|---|---:|---:|:---:|
| 15 | api + util | 57 | +600 | ✅ |
| 16 | playback + handlers + discord + upscaler + updater | 55 | +900 | ✅ |
| 17 | scraper + providers + downloader + SDK + misc | 53 | +600 | ✅ (2026-05-23) |
| 18 | types + SDK client + api success paths + exec mock + appflow + download | 28 | +858 | ✅ (2026-05-24) |
| **TOTAL** | | **193** | **+2958** | ✅ |

**Resultado final FASE 18 (2026-05-24):** 62.9% → **75.7%** · non-main 0%: 26 → **2** · Meta ≥70% SUPERADA.

**Baseline atual do CI (2026-07-23):** **66.8%** com
`go test -short -race -count=1 -covermode=atomic ./...`. O gate começa em
**66.0%** para impedir regressões imediatas; a meta de projeto continua sendo
**≥70.0%** e o piso deve subir junto com novos testes.

**Mapeamento funcional completo:** ver `TEST_PLAN_FUNCTIONS.md` (193 funções listadas por arquivo:linha:nome, agrupadas por fase).

**Atenção ao bug do grep:** Para listar funções 0% use `awk '$NF == "0.0%"'`, NÃO `grep "0.0%"` (este último também matches `100.0%`, `80.0%`, `70.0%`, etc.).
