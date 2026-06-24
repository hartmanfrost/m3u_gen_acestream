package m3u

import (
	"bytes"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"m3u_gen_acestream/acestream"
	"m3u_gen_acestream/config"
	"m3u_gen_acestream/util/logger"
)

func itemNames(srs []acestream.SearchResult) []string {
	var out []string
	for _, sr := range srs {
		for _, it := range sr.Items {
			out = append(out, it.Name)
		}
	}
	return out
}

// TestFilterByNameOrLanguages verifies the OR semantics: a channel is kept when its name matches
// nameRxFilter OR its language is in languagesFilter, covering the sparse-metadata cases where
// each signal alone would miss a real RU/UA channel.
func TestFilterByNameOrLanguages(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.DebugLevel, &buf)

	input := []acestream.SearchResult{{Items: []acestream.Item{
		{Name: "Discovery HD [RU]", Languages: []string{}},    // name tag, language unset -> keep via name
		{Name: "IZ.RU", Languages: []string{"rus"}},           // latin name, language rus -> keep via language
		{Name: "DAZN Baloncesto", Languages: []string{"spa"}}, // neither -> drop
		{Name: "Россия 1 HD", Languages: nil},                 // cyrillic name, language unset -> keep via name
		{Name: "GLAS", Languages: []string{"ukr"}},            // latin name, language ukr -> keep via language
	}}}

	pl := config.Playlist{
		OutputPath:              "test",
		NameRxFilter:            []string{`[Ѐ-ӿ]`, `\[RU\]`},
		LanguagesFilter:         []string{"rus", "ukr"},
		NameRxOrLanguagesFilter: lo.ToPtr(true),
	}

	got := itemNames(filterByNameOrLanguages(log, input, pl))
	assert.ElementsMatch(t,
		[]string{"Discovery HD [RU]", "IZ.RU", "Россия 1 HD", "GLAS"}, got,
		"OR must keep every channel matched by name OR by language, dropping only the Spanish one")
}

// TestFilterByNameOrLanguagesDegrades confirms that with only one of the two filters set the
// combined filter behaves like that single filter (the other side contributes no keeps).
func TestFilterByNameOrLanguagesDegrades(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.DebugLevel, &buf)
	input := []acestream.SearchResult{{Items: []acestream.Item{
		{Name: "Россия 1 HD", Languages: nil},
		{Name: "IZ.RU", Languages: []string{"rus"}},
		{Name: "DAZN", Languages: []string{"spa"}},
	}}}

	nameOnly := config.Playlist{OutputPath: "t", NameRxFilter: []string{`[Ѐ-ӿ]`}, NameRxOrLanguagesFilter: lo.ToPtr(true)}
	assert.ElementsMatch(t, []string{"Россия 1 HD"}, itemNames(filterByNameOrLanguages(log, input, nameOnly)))

	langOnly := config.Playlist{OutputPath: "t", LanguagesFilter: []string{"rus"}, NameRxOrLanguagesFilter: lo.ToPtr(true)}
	assert.ElementsMatch(t, []string{"IZ.RU"}, itemNames(filterByNameOrLanguages(log, input, langOnly)))
}
