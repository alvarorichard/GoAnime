// ===========================================================================
// goyabu_blogger_fix_test.go — Regressão para o bug Goyabu/Blogger batchexecute
//
// Problema detectado: 2026-04-23
//   Diiver reportou via Discord que episódios do Goyabu (fonte PT-BR) não
//   reproduziam. O log de debug mostrava o erro
//   "no video URL found in batchexecute response" em todas as 3 tentativas
//   do extrator de URL do Blogger.
//
// Causa raiz (player/scraper.go — antes do fix de 2026-04-23):
//   O parser do batchexecute assumia que o array de streams estava fixo no
//   índice data[2] do payload inner JSON. Quando o Google alterou o índice
//   (ou retornou data com menos de 3 elementos), o código executava `continue`
//   silenciosamente sem extrair nenhuma URL. Não havia fallback de regex.
//
//   Trecho bugado (removido em 2026-04-23):
//     if len(data) < 3 { continue }        // falha se data tiver 0–2 elementos
//     streams, ok := data[2].([]any)       // índice hardcoded — quebra com mudança do Google
//
// Correção aplicada: 2026-04-23
//   1. O parser itera todos os índices de data[], identificando streams como
//      o primeiro elemento que seja um array de arrays.
//   2. Fallback regex extrai URLs *.googlevideo.com diretamente do corpo bruto
//      caso o parsing estruturado não produza resultado.
//
// Funções testadas:
//   - parseBatchexecuteResponse (real — player/scraper.go, fix 2026-04-23)
//   - parseBatchexecuteResponseLegacy (lógica anterior ao fix, inlinada aqui)
// ===========================================================================

package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constantes de URL usadas nos fixtures de teste
// ---------------------------------------------------------------------------

const (
	googleVideoURL720p = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=22&mime=video%2Fmp4&source=blogger"
	googleVideoURL360p = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=18&mime=video%2Fmp4&source=blogger"
	googleVideoNoMIME  = "https://rr3---sn-q4f7l.googlevideo.com/videoplayback?expire=9999999999&itag=22&source=blogger"
)

// ---------------------------------------------------------------------------
// parseBatchexecuteResponseLegacy — lógica ANTERIOR ao fix de 2026-04-23
//
// Mantida aqui para demonstrar o bug: só verifica data[2] e não tem fallback
// regex. Qualquer resposta com streams fora do índice 2 retorna erro.
// ---------------------------------------------------------------------------
func parseBatchexecuteResponseLegacy(body []byte) (string, error) {
	var videoURL string
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.Contains(line, "wrb.fr") {
			continue
		}
		var outer []any
		if err := json.Unmarshal([]byte(line), &outer); err != nil {
			continue
		}
		for _, entry := range outer {
			arr, ok := entry.([]any)
			if !ok || len(arr) < 3 {
				continue
			}
			if fmt.Sprint(arr[0]) != "wrb.fr" || fmt.Sprint(arr[1]) != "WcwnYd" {
				continue
			}
			var data []any
			if err := json.Unmarshal(fmt.Append(nil, arr[2]), &data); err != nil {
				continue
			}
			// BUG (antes de 2026-04-23): índice hardcoded; falha se data tiver
			// menos de 3 elementos ou se o Google mover streams para outro índice.
			if len(data) < 3 {
				continue
			}
			streams, ok := data[2].([]any)
			if !ok {
				continue
			}
			for _, s := range streams {
				stream, ok := s.([]any)
				if !ok || len(stream) < 1 {
					continue
				}
				u, ok := stream[0].(string)
				if !ok {
					continue
				}
				if strings.Contains(u, "mime=video%2Fmp4") || strings.Contains(u, "mime=video/mp4") {
					videoURL = u
					break
				}
			}
			break
		}
		if videoURL != "" {
			break
		}
	}
	if videoURL == "" {
		return "", errors.New("no video URL found in batchexecute response")
	}
	return videoURL, nil
}

// ---------------------------------------------------------------------------
// buildBatchexecuteBody constrói um corpo de resposta batchexecute realista.
//
// O formato real do Google é:
//   )]}'\n
//   [["wrb.fr","WcwnYd","<inner_json_string>",null,...,"generic"]]\n
//
// onde <inner_json_string> é o payload inner JSON serializado como string.
// streams é posicionado em data[dataIdx] do inner payload.
// ---------------------------------------------------------------------------
func buildBatchexecuteBody(dataIdx int, streamURLs []string) []byte {
	streams := make([]any, len(streamURLs))
	for i, u := range streamURLs {
		streams[i] = []any{u, float64(360)}
	}

	data := make([]any, dataIdx+1)
	data[dataIdx] = streams

	innerJSON, _ := json.Marshal(data)

	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", string(innerJSON), nil, nil, nil, "generic"},
	})

	return []byte(")]}'\n" + string(outerLine) + "\n")
}

// ---------------------------------------------------------------------------
// Testes: formato clássico (streams em data[2]) — ambos os parsers funcionam
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_StreamsAtIndex2_FormatoClassico(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoURL720p, googleVideoURL360p})

	// Parser antigo (pré-fix) funciona porque streams estão em data[2].
	legacyURL, err := parseBatchexecuteResponseLegacy(body)
	require.NoError(t, err, "lógica antiga deve funcionar com streams em data[2]")
	assert.Equal(t, googleVideoURL720p, legacyURL)

	// Parser corrigido também funciona.
	fixedURL, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL720p, fixedURL)
}

// ---------------------------------------------------------------------------
// Testes: simulação do bug 2026-04-23 — streams em índice != 2
// ---------------------------------------------------------------------------

// TestParseBatchexecuteResponse_Bug_StreamsEmIndex0 demonstra o bug:
// quando o Google retorna streams em data[0], o parser antigo falha enquanto
// o parser corrigido (fix 2026-04-23) encontra a URL.
func TestParseBatchexecuteResponse_Bug_StreamsEmIndex0(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(0, []string{googleVideoURL720p})

	// BUG simulado: parser antigo (pré 2026-04-23) falha com streams em data[0].
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: parser antigo NÃO encontra streams em data[0]")

	// SOLUÇÃO: parser corrigido encontra streams em qualquer índice.
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: parser deve encontrar streams em data[0]")
	assert.Equal(t, googleVideoURL720p, url)
}

// TestParseBatchexecuteResponse_Bug_StreamsEmIndex1 cobre o caso em que o
// Google retorna apenas dois elementos em data (índice 0 e 1).
func TestParseBatchexecuteResponse_Bug_StreamsEmIndex1(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(1, []string{googleVideoURL360p})

	// BUG simulado: parser antigo falha (data tem 2 elementos, len(data) < 3).
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: parser antigo falha quando len(data) < 3")

	// SOLUÇÃO: parser corrigido encontra streams em data[1].
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: deve encontrar streams em data[1]")
	assert.Equal(t, googleVideoURL360p, url)
}

// TestParseBatchexecuteResponse_Bug_DataComUmElemento cobre o caso extremo
// onde data possui apenas um elemento (data[0] = streams).
func TestParseBatchexecuteResponse_Bug_DataComUmElemento(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(0, []string{googleVideoURL360p})

	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "BUG 2026-04-23: len(data)==1 < 3, parser antigo descarta")

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL360p, url)
}

// ---------------------------------------------------------------------------
// Testes: seleção de qualidade
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_Prefere720p(t *testing.T) {
	t.Parallel()

	// Resposta com 360p listado antes do 720p.
	body := buildBatchexecuteBody(2, []string{googleVideoURL360p, googleVideoURL720p})

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL720p, url, "deve preferir 720p (itag=22) sobre 360p (itag=18)")
}

func TestParseBatchexecuteResponse_Apenas360p(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoURL360p})

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Equal(t, googleVideoURL360p, url)
}

// ---------------------------------------------------------------------------
// Testes: fallback regex (fix 2026-04-23)
// ---------------------------------------------------------------------------

// TestParseBatchexecuteResponse_FallbackRegex verifica que, quando o parsing
// estruturado não encontra streams, o regex ainda captura URLs googlevideo.com
// presentes no corpo bruto da resposta.
func TestParseBatchexecuteResponse_FallbackRegex(t *testing.T) {
	t.Parallel()

	// Corpo com URL googlevideo.com mas sem estrutura wrb.fr/WcwnYd válida.
	body := []byte(`)]}'\n` +
		`[["wrb.fr","WcwnYd","{}",null,null,null,"generic"]]` + "\n" +
		`debug: ` + googleVideoURL720p + ` extra`)

	// Parser antigo (sem fallback regex) falha.
	_, err := parseBatchexecuteResponseLegacy(body)
	assert.Error(t, err, "parser antigo não tem fallback regex")

	// Parser corrigido encontra via regex.
	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err, "fix 2026-04-23: fallback regex deve encontrar URL googlevideo.com")
	assert.Contains(t, url, ".googlevideo.com")
}

// TestParseBatchexecuteResponse_FallbackRegex_StreamsSemMIME cobre streams
// que existem na estrutura mas não têm mime=video%2Fmp4, enquanto o corpo
// bruto contém uma URL googlevideo.com válida.
func TestParseBatchexecuteResponse_FallbackRegex_StreamsSemMIME(t *testing.T) {
	t.Parallel()

	body := buildBatchexecuteBody(2, []string{googleVideoNoMIME})
	// Adiciona URL válida em texto livre no final do corpo.
	body = append(body, []byte("\ninfo: "+googleVideoURL360p)...)

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	assert.Contains(t, url, ".googlevideo.com")
}

// ---------------------------------------------------------------------------
// Testes: casos de erro / resposta inválida
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_CorpoVazio(t *testing.T) {
	t.Parallel()

	_, err := parseBatchexecuteResponse([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no video URL found")
}

func TestParseBatchexecuteResponse_JSONInvalido(t *testing.T) {
	t.Parallel()

	body := []byte(")]}'\n[this is not valid json]\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_SemLinhaWrbFr(t *testing.T) {
	t.Parallel()

	body := []byte(")]}'\n[[\"other.method\",\"SomeRPC\",\"[]\",null]]\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_DataSemArrayDeArrays(t *testing.T) {
	t.Parallel()

	// data contém apenas strings e números, nenhum array de arrays.
	innerJSON := `["string_value", 42, null]`
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", innerJSON, nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

func TestParseBatchexecuteResponse_StreamsVazios(t *testing.T) {
	t.Parallel()

	// data[2] é um array vazio — não atende à condição len(s) > 0.
	innerJSON := `[null, null, []]`
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", innerJSON, nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	_, err := parseBatchexecuteResponse(body)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Testes: múltiplos índices — parser deve usar o primeiro válido encontrado
// ---------------------------------------------------------------------------

func TestParseBatchexecuteResponse_UsaPrimeiroIndiceValido(t *testing.T) {
	t.Parallel()

	// data[1] = streams com 360p; data[3] = streams com 720p.
	// Parser deve retornar o primeiro válido encontrado (data[1], 360p).
	streams360 := []any{[]any{googleVideoURL360p, float64(360)}}
	streams720 := []any{[]any{googleVideoURL720p, float64(720)}}

	data := []any{nil, streams360, nil, streams720}
	innerJSON, _ := json.Marshal(data)
	outerLine, _ := json.Marshal([][]any{
		{"wrb.fr", "WcwnYd", string(innerJSON), nil, nil, nil, "generic"},
	})
	body := []byte(")]}'\n" + string(outerLine) + "\n")

	url, err := parseBatchexecuteResponse(body)
	require.NoError(t, err)
	// Streams em data[1] (360p) devem ser encontrados primeiro.
	assert.Equal(t, googleVideoURL360p, url)
}
