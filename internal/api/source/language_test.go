package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeLanguage(t *testing.T) {
	for scope, want := range map[string]string{
		"ptbr": "PT", "pt-br": "PT", "PTBR": "PT", "english": "EN", "English": "EN",
		"": "", "movietv": "", "hianime": "", "plutotv": "", "pluto": "", "animefire": "", "all": "",
	} {
		assert.Equal(t, want, ScopeLanguage(scope), "%q", scope)
		assert.Equal(t, want, ScopeLanguage(" "+scope+"	"), "%q, with spaces around it", scope)
	}
}

func TestWithLanguageCarriesOnlyALanguageItKnows(t *testing.T) {
	base := context.Background()
	assert.Empty(t, LanguageFrom(base), "nothing asked for")
	assert.Equal(t, "PT", LanguageFrom(WithLanguage(base, "PT")))
	assert.Equal(t, "EN", LanguageFrom(WithLanguage(base, " en ")), "however it is written")
	for _, lang := range []string{"", "FR", "PT-BR", "english"} {
		ctx := WithLanguage(base, lang)
		assert.Empty(t, LanguageFrom(ctx), "%q is not a language a tab asks for", lang)
		assert.Equal(t, base, ctx, "%q: the context is returned as it came", lang)
	}
}

func TestTheLanguageSurvivesADerivedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(WithLanguage(context.Background(), "EN"))
	defer cancel()
	assert.Equal(t, "EN", LanguageFrom(ctx), "the fan-out derives timeouts from the search's context")
	assert.Equal(t, "PT", LanguageFrom(WithLanguage(ctx, "PT")), "the innermost asking wins")
}
