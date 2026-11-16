package api

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests, for Pluto TV. Same runner as the other sources (a mutant
// breaks one decision in the code; the guard tests it names must fail on it),
// and four guards over the method itself:
//
//   - every mutant is well formed: its snippet is in the code exactly once,
//     inside the function it says it breaks, and every guard test it names
//     exists (TestPlutoTVMutantsAreWellFormed);
//   - the check that says so is itself checked against tables with each
//     defect in turn (TestTheMutantCheckSeesEveryDefect);
//   - every function of the feature has a mutant, or a written reason it has
//     none, and a reason is dropped once it stops being true
//     (TestPlutoTVFunctionsAreAllMutated);
//   - every function of the feature is fully covered when its own tests run
//     (TestPlutoTVFunctionsAreFullyCovered).
//
// The fixtures' own mutation table, which asks the same of the data the tests
// are fitted to, is in internal/scraper/providers/plutotv/meta_test.go.

const (
	ptPkg           = "./internal/scraper/providers/plutotv"
	ptAPIPkg        = "./internal/api"
	ptProvidersPkg  = "./internal/api/providers"
	ptScraperPkg    = "./internal/scraper"
	ptHandlersPkg   = "./internal/handlers"
	ptTUIPkg        = "./internal/tui"
	ptDownloadPkg   = "./internal/download"
	ptTypesPkg      = "./pkg/goanime/types"
	ptModelsPkg     = "./internal/models"
	ptClientFile    = "internal/scraper/providers/plutotv/client.go"
	ptAPIFile       = "internal/api/plutotv.go"
	ptProvidersFile = "internal/api/providers/source_providers.go"
	ptTaggingFile   = "internal/api/providers/tagging.go"
	ptNamingFile    = "internal/api/providers/naming/naming.go"
	ptManagerFile   = "internal/scraper/manager.go"
	ptHealthFile    = "internal/scraper/source_health.go"
	ptEnhancedFile  = "internal/api/enhanced.go"
	ptContinueFile  = "internal/handlers/continue.go"
	ptHomeFile      = "internal/handlers/searchhome.go"
	ptTryFile       = "internal/tui/try.go"
	ptWorkflowFile  = "internal/download/workflow.go"
	ptTypesFile     = "pkg/goanime/types/source.go"
	ptMediaFile     = "internal/models/media.go"
)

// plutoMutant is a mutant that says which function it breaks, so the table can
// be held to "every function has one" and "this one is where it says".
type plutoMutant struct {
	mutant
	file string
	// fn is the function the snippet sits in; "var:name" or "const:name" for a
	// package-level declaration.
	fn string
}

func pm(name, file, fn, from, to string, kills map[string][]string) plutoMutant {
	return plutoMutant{mutant: mutant{name: name, edits: []codeEdit{{file, from, to}}, kills: kills}, file: file, fn: fn}
}

func guard(pkg string, tests ...string) map[string][]string { return map[string][]string{pkg: tests} }

func and(maps ...map[string][]string) map[string][]string {
	out := map[string][]string{}
	for _, m := range maps {
		for pkg, tests := range m {
			out[pkg] = append(out[pkg], tests...)
		}
	}
	return out
}

var plutoMutants = []plutoMutant{
	// ---- identifiers
	pm("Shared builds a client per call", ptClientFile, "Shared",
		"sharedOnce.Do(func() { sharedClient = NewClient() })", "sharedClient = NewClient()",
		guard(ptPkg, "TestSharedIsOneClient")),
	pm("the client names no User-Agent of its own", ptClientFile, "NewClient",
		"userAgent: netx.UserAgent,", `userAgent: "",`,
		guard(ptPkg, "TestSearchSendsWhatTheServiceExpects", "TestRegressionTheClientAlwaysNamesAUserAgentOnTheMediaHost")),
	pm("the test client trusts no server of its own", ptClientFile, "NewClientForTest",
		"c.allow = map[string]bool{u.Host: true}", "c.allow = nil",
		guard(ptPkg, "TestTheTestClientTrustsOnlyItsOwnServerBesidesPluto")),
	pm("the client id loses its version bits", ptClientFile, "newClientID",
		"b[6] = (b[6] & 0x0f) | 0x40", "b[6] = b[6] & 0x0f",
		guard(ptPkg, "TestNewClientID")),
	pm("a series URL is spelled another way", ptClientFile, "SeriesURL",
		`return siteBase + "/on-demand/series/" + id`, `return siteBase + "/series/" + id`,
		guard(ptPkg, "TestParsePageURL", "TestTitleURLFollowsTheKind")),
	pm("a movie URL is spelled another way", ptClientFile, "MovieURL",
		`return siteBase + "/on-demand/movies/" + id`, `return siteBase + "/movies/" + id`,
		guard(ptPkg, "TestParsePageURL", "TestIsMovieURL")),
	pm("an episode URL is spelled another way", ptClientFile, "EpisodeURL",
		`return SeriesURL(seriesID) + "/episode/" + episodeID`, `return SeriesURL(seriesID) + "/ep/" + episodeID`,
		guard(ptPkg, "TestParsePageURL")),
	pm("a movie's URL is a series'", ptClientFile, "URL",
		"\tif t.Kind == KindMovie {\n\t\treturn MovieURL(t.ID)\n\t}", "\tif false {\n\t\treturn MovieURL(t.ID)\n\t}",
		guard(ptPkg, "TestTitleURLFollowsTheKind")),
	pm("a series has a year", ptClientFile, "Year",
		"\tif t.Kind != KindMovie {\n\t\treturn \"\"\n\t}\n", "",
		guard(ptPkg, "TestTitleYear")),
	pm("English is tagged Portuguese", ptClientFile, "languageTag",
		"case \"EN\":\n\t\treturn \"[English]\"", "case \"EN\":\n\t\treturn \"[PT-BR]\"",
		guard(ptPkg, "TestLanguageTag")),
	pm("a movie is tagged a series", ptClientFile, "ToAnimeModel",
		`mediaType, typeTag = models.MediaTypeMovie, "[Movie]"`, `mediaType, typeTag = models.MediaTypeMovie, "[TV]"`,
		guard(ptPkg, "TestToAnimeModel")),
	pm("the language tag comes before the type tag", ptClientFile, "ToAnimeModel",
		`fmt.Sprintf("%s %s %s", typeTag, languageTag(t.Language), t.Name)`, `fmt.Sprintf("%s %s %s", languageTag(t.Language), typeTag, t.Name)`,
		and(guard(ptPkg, "TestToAnimeModel"), guard(ptScraperPkg, "TestPlutoTVAdapterSearchConvertsTitlesAndSkipsChannels"))),
	pm("an http URL is a Pluto TV page", ptClientFile, "parsePageURL",
		`u.User != nil || u.Scheme != "https" || strings.TrimPrefix`, `u.User != nil || strings.TrimPrefix`,
		guard(ptPkg, "TestParsePageURL")),
	pm("any host ending in pluto.tv is Pluto TV", ptClientFile, "parsePageURL",
		`strings.TrimPrefix(strings.ToLower(u.Host), "www.") != "pluto.tv"`, `!strings.HasSuffix(strings.ToLower(u.Host), "pluto.tv")`,
		guard(ptPkg, "TestParsePageURL")),
	pm("a movie has episodes", ptClientFile, "parsePageURL",
		`if m == nil || (m[1] == "movies" && m[3] != "") {`, `if m == nil {`,
		guard(ptPkg, "TestParsePageURL")),
	pm("a movie page reads as a series", ptClientFile, "parsePageURL",
		"\t\tref.kind = KindMovie\n", "\t\tref.kind = KindSeries\n",
		guard(ptPkg, "TestParsePageURL")),
	pm("a movie is not a movie", ptClientFile, "IsMovieURL",
		"ref.kind == KindMovie", "ref.kind == KindSeries",
		guard(ptPkg, "TestIsMovieURL")),
	pm("a series plays by its own id", ptClientFile, "playID",
		"if r.episodeID != \"\" {", "if false {",
		guard(ptPkg, "TestPlayID", "TestStreamPlaysAnEpisodeByTheEpisodesIDNotTheSeries")),
	pm("a series names something to play", ptClientFile, "playID",
		"return r.id, r.kind == KindMovie", "return r.id, true",
		guard(ptPkg, "TestPlayID", "TestStreamOfASeriesIsRefused")),

	// ---- trust
	pm("a look-alike domain is Pluto's", ptClientFile, "trusted",
		`strings.HasSuffix(host, "."+domain)`, `strings.HasSuffix(host, domain)`,
		guard(ptPkg, "TestTrusted")),
	pm("http is trusted", ptClientFile, "trusted",
		"\tif u.Scheme != \"https\" {\n\t\treturn false\n\t}\n\thost :=", "\thost :=",
		guard(ptPkg, "TestTrusted")),
	pm("a URL with credentials is trusted", ptClientFile, "trusted",
		`err != nil || u.Host == "" || u.User != nil {`, `err != nil || u.Host == "" {`,
		guard(ptPkg, "TestTrusted")),
	pm("the test server may be reached over any scheme", ptClientFile, "trusted",
		`return u.Scheme == "http" || u.Scheme == "https"`, `return true`,
		guard(ptPkg, "TestTheTestClientTrustsOnlyItsOwnServerBesidesPluto")),

	// ---- session
	pm("a session is never reused", ptClientFile, "currentSession",
		"if c.sess != nil && time.Now().Before(c.sess.expires) {", "if false {",
		guard(ptPkg, "TestOneSessionServesManyCalls")),
	pm("a session never runs out", ptClientFile, "currentSession",
		"if c.sess != nil && time.Now().Before(c.sess.expires) {", "if c.sess != nil {",
		guard(ptPkg, "TestASessionThatRanOutIsReplaced")),
	pm("a new session is not kept", ptClientFile, "currentSession",
		"\tc.sess = s\n", "",
		guard(ptPkg, "TestOneSessionServesManyCalls")),
	pm("a rejected session is kept", ptClientFile, "dropSession",
		"\t\tc.sess = nil\n", "",
		guard(ptPkg, "TestA401RenewsTheSessionOnceAndRetries", "TestStreamIsRenewedAfterA401")),
	pm("the service is told ads are welcome", ptClientFile, "boot",
		`"serverSideAds":       {"false"},`, `"serverSideAds":       {"true"},`,
		guard(ptPkg, "TestBootOpensAWebSessionWithAnIdentityOfItsOwn")),
	pm("the session names no device", ptClientFile, "boot",
		`"clientID":            {c.clientID},`, `"clientID":            {"anonymous"},`,
		guard(ptPkg, "TestBootOpensAWebSessionWithAnIdentityOfItsOwn")),
	pm("a session without a token is accepted", ptClientFile, "boot",
		"\tif s.token == \"\" {\n\t\treturn nil, netx.NewParserError(SourceName, \"session\", \"the session answer carries no token\", nil)\n\t}\n", "",
		guard(ptPkg, "TestBootFailures")),
	pm("a service host that is not Pluto's is accepted", ptClientFile, "boot",
		"if !c.trusted(host) {", "if !c.trusted(host) && false {",
		guard(ptPkg, "TestBootFailures")),
	pm("the stitcher params keep their edge separators", ptClientFile, "boot",
		`strings.Trim(strings.TrimSpace(answer.StitcherParams), "&?")`, `strings.TrimSpace(answer.StitcherParams)`,
		guard(ptPkg, "TestStitcherParamsAreUsedWithoutTheirEdgeSeparators")),
	pm("a long refresh is not capped", ptClientFile, "boot",
		"refresh > 0 && refresh < lifetime", "refresh > 0",
		guard(ptPkg, "TestTheSessionLifetimeFollowsTheServiceButNeverOutlastsTheCap")),
	pm("a negative refresh expires the session", ptClientFile, "boot",
		"refresh > 0 && refresh < lifetime", "refresh < lifetime",
		guard(ptPkg, "TestTheSessionLifetimeFollowsTheServiceButNeverOutlastsTheCap")),
	pm("the refresh is read in minutes", ptClientFile, "boot",
		"time.Duration(answer.RefreshInSec) * time.Second", "time.Duration(answer.RefreshInSec) * time.Minute",
		guard(ptPkg, "TestTheSessionLifetimeFollowsTheServiceButNeverOutlastsTheCap")),
	pm("a 401 is retried without end", ptClientFile, "withSession",
		"status == http.StatusUnauthorized && attempt == 0", "status == http.StatusUnauthorized",
		guard(ptPkg, "TestASecond401IsAnErrorNotALoop")),
	pm("a 401 is retried on the same session", ptClientFile, "withSession",
		"\t\t\tc.dropSession(s)\n\t\t\tcontinue", "\t\t\tcontinue",
		guard(ptPkg, "TestA401RenewsTheSessionOnceAndRetries", "TestStreamIsRenewedAfterA401")),

	// ---- search
	pm("a channel is a title", ptClientFile, "Search",
		"(kind != KindMovie && kind != KindSeries) || ", "",
		and(guard(ptPkg, "TestSearchKeepsOnlyWhatCanBePlayed", "TestRegressionSearchOfARealAnswerKeepsOnlyTitles"),
			guard(ptScraperPkg, "TestPlutoTVAdapterSearchConvertsTitlesAndSkipsChannels"))),
	pm("a title with an id that is no id is kept", ptClientFile, "Search",
		"|| !idRe.MatchString(hit.ID) ||", "||",
		guard(ptPkg, "TestSearchDropsWhatItCannotAddress")),
	pm("a title without a name is kept", ptClientFile, "Search",
		`|| name == "" ||`, "||",
		guard(ptPkg, "TestSearchDropsWhatItCannotAddress")),
	pm("a title listed twice is kept twice", ptClientFile, "Search",
		" || seen[hit.ID] {", " {",
		guard(ptPkg, "TestSearchDropsWhatItCannotAddress")),
	pm("a blank query is searched", ptClientFile, "Search",
		"\tif query == \"\" {\n\t\treturn nil, nil\n\t}\n", "",
		guard(ptPkg, "TestSearchOfNothingAsksNothing")),
	pm("a query is searched untrimmed", ptClientFile, "Search",
		"\tquery = strings.TrimSpace(query)\n", "",
		guard(ptPkg, "TestSearchOfNothingAsksNothing", "TestSearchSendsWhatTheServiceExpects")),
	pm("the search asks for five hits", ptClientFile, "Search",
		`"limit": {fmt.Sprint(searchLimit)}`, `"limit": {"5"}`,
		guard(ptPkg, "TestSearchSendsWhatTheServiceExpects")),

	// ---- seasons
	pm("a movie is listed as a series", ptClientFile, "Show",
		"if ref.kind != KindSeries {", "if false {",
		guard(ptPkg, "TestShowRefusals")),
	pm("another series' answer is accepted", ptClientFile, "Show",
		"if answer.ID != ref.id {", "if false {",
		guard(ptPkg, "TestShowRefusals")),
	pm("an episode seen twice is listed twice", ptClientFile, "Show",
		"if !idRe.MatchString(e.ID) || seen[e.ID] {", "if !idRe.MatchString(e.ID) {",
		guard(ptPkg, "TestShowSkipsWhatCannotBePlayed")),
	pm("an empty season is offered", ptClientFile, "Show",
		"\t\tif len(season.Episodes) == 0 {\n\t\t\tcontinue\n\t\t}\n", "",
		guard(ptPkg, "TestShowSkipsWhatCannotBePlayed")),
	pm("unnumbered episodes keep their zero", ptClientFile, "Show",
		"season.Episodes[i].Number <= 0 || (i > 0 && season.Episodes[i].Number <= season.Episodes[i-1].Number)", "season.Episodes[i].Number < 0",
		guard(ptPkg, "TestShowNumbersEpisodesTheServiceLeftUnnumbered")),
	pm("a duration is read in seconds", ptClientFile, "Show",
		"time.Duration(e.Duration) * time.Millisecond", "time.Duration(e.Duration) * time.Second",
		and(guard(ptPkg, "TestShowFillsEveryFieldOfAnEpisode", "TestRegressionSeasonsOfARealAnswer"), guard(ptAPIPkg, "TestPlutoTVEpisodesCarryWhatThePlayerNeeds"))),
	pm("episodes are listed backwards", ptClientFile, "Show",
		"season.Episodes[i].Number < season.Episodes[j].Number", "season.Episodes[i].Number > season.Episodes[j].Number",
		guard(ptPkg, "TestShowOrdersSeasonsAndEpisodes")),
	pm("seasons are listed backwards", ptClientFile, "Show",
		"show.Seasons[i].Number < show.Seasons[j].Number", "show.Seasons[i].Number > show.Seasons[j].Number",
		guard(ptPkg, "TestShowOrdersSeasonsAndEpisodes")),
	pm("only the first ten episodes are asked for", ptClientFile, "Show",
		"offset=1000&page=1", "offset=10&page=1",
		guard(ptPkg, "TestShowAsksForTheWholeSeriesAtOnce")),
	pm("an episode keeps its name untrimmed", ptClientFile, "Show",
		"Name:        strings.TrimSpace(e.Name),", "Name:        e.Name,",
		guard(ptPkg, "TestShowFillsEveryFieldOfAnEpisode")),
	pm("an episode forgets its season", ptClientFile, "Show",
		"Season:      s.Number,", "Season:      0,",
		guard(ptPkg, "TestShowFillsEveryFieldOfAnEpisode")),
	pm("a series without episodes is listed", ptClientFile, "Show",
		"if len(show.Seasons) == 0 {", "if false {",
		guard(ptPkg, "TestShowRefusals")),

	// ---- stream
	pm("five rungs are tried", ptClientFile, "Stream",
		"if i == 3 {", "if i == 5 {",
		guard(ptPkg, "TestStreamTriesAtMostThreeRungsAndThenPlaysTheStitchedMaster")),
	pm("only the best rung is tried", ptClientFile, "Stream",
		"if i == 3 {", "if i == 1 {",
		guard(ptPkg, "TestStreamTriesTheNextRungWhenOneCannotBeUsed", "TestStreamTriesAtMostThreeRungsAndThenPlaysTheStitchedMaster")),
	pm("the quality is ignored", ptClientFile, "Stream",
		"for i, v := range orderVariants(variants, quality) {", `for i, v := range orderVariants(variants, "best") {`,
		guard(ptPkg, "TestStreamPicksTheRungQualityAsksFor")),
	pm("a rendition on another host is fetched", ptClientFile, "Stream",
		"\t\tif !c.trusted(v.URL) {\n\t\t\tcontinue\n\t\t}\n", "",
		guard(ptPkg, "TestStreamNeverFollowsAHostThatIsNotPlutos")),
	pm("a title clip on another host is fetched", ptClientFile, "Stream",
		"if !ok || !c.trusted(contentURL) {", "if !ok {",
		guard(ptPkg, "TestStreamNeverFollowsAHostThatIsNotPlutos")),
	pm("a protected master is played", ptClientFile, "Stream",
		"\tif isProtected(master) {\n\t\treturn Stream{}, ErrProtected\n\t}\n", "",
		guard(ptPkg, "TestStreamRefusesWhatIsProtected")),
	pm("a protected rendition is played", ptClientFile, "Stream",
		"\t\tif isProtected(playlist) {\n\t\t\treturn Stream{}, ErrProtected\n\t\t}\n", "",
		guard(ptPkg, "TestStreamRefusesWhatIsProtected")),
	pm("a protected title is traded for another rung", ptClientFile, "Stream",
		"\t\t\tif errors.Is(err, ErrProtected) {\n\t\t\t\t// The title itself is keyed; another rung of it is too.\n\t\t\t\treturn Stream{}, err\n\t\t\t}\n", "",
		guard(ptPkg, "TestStreamRefusesWhatIsProtected")),
	pm("a caller who left still gets the next rendition", ptClientFile, "Stream",
		"\t\t\tif ctx.Err() != nil {\n\t\t\t\treturn Stream{}, ctx.Err()\n\t\t\t}\n\t\t\tutil.Debug(\"Pluto TV rendition", "\t\t\tutil.Debug(\"Pluto TV rendition",
		guard(ptPkg, "TestStreamStopsWhenTheCallerGivesUp")),
	pm("a caller who left still gets the next content playlist", ptClientFile, "Stream",
		"\t\t\tif ctx.Err() != nil {\n\t\t\t\treturn Stream{}, ctx.Err()\n\t\t\t}\n\t\t\tif errors.Is(err, ErrProtected)", "\t\t\tif errors.Is(err, ErrProtected)",
		guard(ptPkg, "TestStreamStopsWhenTheCallerGivesUpOnTheContentPlaylist")),
	pm("no playlist is an error, not the stitched master", ptClientFile, "Stream",
		"return Stream{URL: masterURL}, nil", `return Stream{}, errors.New("no playlist")`,
		guard(ptPkg, "TestStreamFallsBackToTheStitchedPlaylistWhenTheTitleIsSeveralClips", "TestStreamTriesAtMostThreeRungsAndThenPlaysTheStitchedMaster")),
	pm("the title's own playlist is not called ad-free", ptClientFile, "Stream",
		"Stream{URL: contentURL, AdFree: true}", "Stream{URL: contentURL, AdFree: false}",
		guard(ptPkg, "TestStreamReturnsTheTitlesOwnPlaylist")),
	pm("the token goes into the URL as it is", ptClientFile, "Stream",
		"url.QueryEscape(s.token)", "s.token",
		guard(ptPkg, "TestStreamAsksTheStitcherTheWayTheSiteDoes")),
	pm("the stitcher is not asked to pass the token", ptClientFile, "Stream",
		"&jwt=%s&masterJWTPassthrough=true\",", "&jwt=%s\",",
		guard(ptPkg, "TestStreamAsksTheStitcherTheWayTheSiteDoes", "TestStreamReturnsTheTitlesOwnPlaylist")),
	pm("a series URL is played", ptClientFile, "Stream",
		"id, ok := ref.playID()", "id, _ := ref.playID()\n\tok := true",
		guard(ptPkg, "TestStreamOfASeriesIsRefused")),
	pm("a master without rungs is played", ptClientFile, "Stream",
		"if len(variants) == 0 {", "if false {",
		guard(ptPkg, "TestStreamErrors")),

	// ---- playlists
	pm("a key system is no reason to refuse", ptClientFile, "isProtected",
		`method != "NONE" && method != "AES-128"`, `method != "NONE"`,
		guard(ptPkg, "TestIsProtected")),
	pm("an unkeyed playlist is protected", ptClientFile, "isProtected",
		`method != "NONE" && method`, `method`,
		guard(ptPkg, "TestIsProtected")),
	pm("a method is read case-sensitively", ptClientFile, "isProtected",
		"strings.ToUpper(m[1])", "m[1]",
		guard(ptPkg, "TestIsProtected")),
	pm("a session key is not read", ptClientFile, "var:keyMethodRe",
		"^#EXT-X-(?:SESSION-)?KEY:", "^#EXT-X-KEY:",
		guard(ptPkg, "TestIsProtected", "TestStreamRefusesWhatIsProtected")),
	pm("a rung is not resolved against its master", ptClientFile, "parseMaster",
		"pending.URL = base.ResolveReference(ref).String()", "pending.URL = ref.String() + base.Host[:0]",
		guard(ptPkg, "TestParseMaster", "TestRegressionTheRealMasterReadsAsFiveRungsWithRelativeURIs")),
	pm("a blank line ends a rung's description", ptClientFile, "parseMaster",
		`case line == "" || strings.HasPrefix(line, "#"):`, `case strings.HasPrefix(line, "#"):`,
		guard(ptPkg, "TestParseMaster")),
	pm("a rung's height is not read", ptClientFile, "parseMaster",
		"fmt.Sscan(m[1], &pending.Height)", "fmt.Sscan(m[1], &pending.Bandwidth)",
		guard(ptPkg, "TestParseMaster")),
	pm("480p is a megabit", ptClientFile, "bandwidthBudget",
		"return 2_000_000", "return 1_000_000",
		guard(ptPkg, "TestBandwidthBudget")),
	pm("720p is 700p", ptClientFile, "bandwidthBudget",
		"case height < 720:", "case height < 700:",
		guard(ptPkg, "TestBandwidthBudget")),
	pm("the best rung is the smallest", ptClientFile, "orderVariants",
		"ordered[i].Bandwidth > ordered[j].Bandwidth", "ordered[i].Bandwidth < ordered[j].Bandwidth",
		guard(ptPkg, "TestOrderVariants")),
	pm("worst is not worst", ptClientFile, "orderVariants",
		`if quality == "worst" {`, `if quality == "never" {`,
		guard(ptPkg, "TestOrderVariants", "TestStreamPicksTheRungQualityAsksFor")),
	pm("a rung of the height asked for does not fit", ptClientFile, "orderVariants",
		"return v.Height <= height", "return v.Height < height",
		guard(ptPkg, "TestOrderVariants")),
	pm("the farthest rung above comes first", ptClientFile, "orderVariants",
		"above = append([]variant{v}, above...)", "above = append(above, v)",
		guard(ptPkg, "TestOrderVariants")),
	pm("the rungs above are dropped", ptClientFile, "orderVariants",
		"return append(within, above...)", "return within",
		guard(ptPkg, "TestOrderVariants")),
	pm("the title is the first clip whatever its start", ptClientFile, "contentPlaylistURL",
		`c.fromStart = c.fromStart || m[2] == "0"`, "c.fromStart = true",
		guard(ptPkg, "TestContentPlaylistURL")),
	pm("the title is the first clip whatever its end", ptClientFile, "contentPlaylistURL",
		`c.toTheEnd = c.toTheEnd || m[3] == "end"`, "c.toTheEnd = true",
		guard(ptPkg, "TestContentPlaylistURL")),
	pm("a cut title is called whole", ptClientFile, "contentPlaylistURL",
		"if title == nil || !title.fromStart || !title.toTheEnd {", "if title == nil {",
		guard(ptPkg, "TestContentPlaylistURL", "TestRegressionTheRealStitchedPlaylistsPointAtTheTitlesOwnPlaylist")),
	pm("a second big clip is no reason to give up", ptClientFile, "contentPlaylistURL",
		"c != title && c.segments > maxInterstitialSegments", "c != title && c.segments > 1000",
		guard(ptPkg, "TestContentPlaylistURL", "TestStreamFallsBackToTheStitchedPlaylistWhenTheTitleIsSeveralClips")),
	pm("the rendition's name is left out", ptClientFile, "contentPlaylistURL",
		`title.dir + "0-end/" + title.rendition + ".m3u8"`, `title.dir + "0-end/" + ".m3u8"`,
		guard(ptPkg, "TestContentPlaylistURL", "TestContentPlaylistURLNamesTheRenditionTheSegmentsAre")),
	pm("a bumper may be six hundred segments long", ptClientFile, "const:maxInterstitialSegments",
		"maxInterstitialSegments = 6", "maxInterstitialSegments = 600",
		guard(ptPkg, "TestContentPlaylistURL")),
	pm("a page is a media playlist", ptClientFile, "checkMediaPlaylist",
		`if !strings.HasPrefix(strings.TrimSpace(body), "#EXTM3U") || !strings.Contains(body, "#EXTINF") {`, `if !strings.HasPrefix(strings.TrimSpace(body), "#EXTM3U") {`,
		guard(ptPkg, "TestCheckMediaPlaylist", "TestStreamTriesTheNextRungWhenOneCannotBeUsed")),
	pm("a protected title playlist is checked by nobody", ptClientFile, "checkMediaPlaylist",
		"\tif isProtected(body) {\n\t\treturn ErrProtected\n\t}\n", "",
		guard(ptPkg, "TestCheckMediaPlaylist", "TestStreamRefusesWhatIsProtected")),

	// ---- HTTP
	pm("every host is sent the token", ptClientFile, "get",
		`if token != "" {`, "if true {",
		guard(ptPkg, "TestGetNeverCarriesATokenItWasNotGiven", "TestPlaylistsAreFetchedWithoutTheSessionToken")),
	pm("the request names no User-Agent", ptClientFile, "get",
		`req.Header.Set("User-Agent", c.userAgent)`, `req.Header.Set("User-Agent", "")`,
		guard(ptPkg, "TestSearchSendsWhatTheServiceExpects", "TestRegressionTheClientAlwaysNamesAUserAgentOnTheMediaHost")),
	pm("the request names no origin", ptClientFile, "get",
		"\treq.Header.Set(\"Origin\", siteBase)\n", "",
		guard(ptPkg, "TestSearchSendsWhatTheServiceExpects")),
	pm("an oversize body is read whole", ptClientFile, "get",
		"io.LimitReader(resp.Body, limit+1)", "io.LimitReader(resp.Body, limit)",
		guard(ptPkg, "TestStreamErrors")),
	pm("a body exactly at the limit is oversize", ptClientFile, "get",
		"int64(len(body)) > limit", "int64(len(body)) >= limit",
		guard(ptPkg, "TestStreamErrors")),
	pm("a page is read as JSON without a word", ptClientFile, "getJSON",
		`"invalid API JSON"`, `"bad json"`,
		guard(ptPkg, "TestSearchErrors", "TestBootFailures")),
	pm("the playlists are fetched with the token", ptClientFile, "getText",
		`c.get(ctx, target, "", "application/x-mpegURL, */*", layer, maxPlaylistBytes)`, `c.get(ctx, target, "tok", "application/x-mpegURL, */*", layer, maxPlaylistBytes)`,
		guard(ptPkg, "TestPlaylistsAreFetchedWithoutTheSessionToken")),

	// ---- the api layer
	pm("a movie is listed through the catalog", ptAPIFile, "GetPlutoTVEpisodes",
		"if plutotv.IsMovieURL(media.URL) {", "if false {",
		guard(ptAPIPkg, "TestPlutoTVMovieIsOneEpisodeAndAsksNothing")),
	pm("a cancelled season picker is a failure, not a way back", ptAPIFile, "GetPlutoTVEpisodes",
		"if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {", "if errors.Is(err, tui.ErrPickBack) {",
		guard(ptAPIPkg, "TestPlutoTVBackAndCancelReturnToTheSearch")),
	pm("the season rows carry no count", ptAPIFile, "GetPlutoTVEpisodes",
		"Details: episodeCountLabel(len(s.Episodes)),", `Details: "",`,
		guard(ptAPIPkg, "TestPlutoTVSeriesAsksForASeasonAndListsIt")),
	pm("the season rows carry no source", ptAPIFile, "GetPlutoTVEpisodes",
		"Tags:    []string{plutotv.SourceName},", "Tags:    nil,",
		guard(ptAPIPkg, "TestPlutoTVSeriesAsksForASeasonAndListsIt")),
	pm("the seasons are numbered by position", ptAPIFile, "GetPlutoTVSeasonNumbers",
		"nums[i] = s.Number", "nums[i] = i + 0*s.Number",
		guard(ptAPIPkg, "TestPlutoTVSeasonNumbersAndEpisodesForDownloads")),
	pm("a season that is not there is the first", ptAPIFile, "plutoSeasonEpisodes",
		"\t\tif s.Number != number {\n\t\t\tcontinue\n\t\t}\n", "",
		guard(ptAPIPkg, "TestPlutoTVSeasonNumbersAndEpisodesForDownloads")),
	pm("the chosen season is not remembered", ptAPIFile, "plutoSeasonEpisodes",
		"\t\tmedia.CurrentSeason = number\n", "",
		guard(ptAPIPkg, "TestPlutoTVSeriesAsksForASeasonAndListsIt", "TestPlutoTVSingleSeasonSkipsThePicker")),
	pm("an episode's duration is in milliseconds", ptAPIFile, "plutoSeasonEpisodes",
		"int(e.Duration / time.Second)", "int(e.Duration / time.Millisecond)",
		guard(ptAPIPkg, "TestPlutoTVEpisodesCarryWhatThePlayerNeeds")),
	pm("an episode plays by its series", ptAPIFile, "plutoSeasonEpisodes",
		"plutotv.EpisodeURL(show.ID, e.ID)", "plutotv.EpisodeURL(show.ID, show.ID)",
		guard(ptAPIPkg, "TestPlutoTVEpisodesCarryWhatThePlayerNeeds")),
	pm("an unnamed episode has no name", ptAPIFile, "plutoSeasonEpisodes",
		`title = fmt.Sprintf("Episode %d", e.Number)`, `title = ""`,
		guard(ptAPIPkg, "TestPlutoTVSeriesAsksForASeasonAndListsIt")),
	pm("a movie belongs to season one", ptAPIFile, "plutoMovieEpisode",
		`SeasonID: "0",`, `SeasonID: "1",`,
		guard(ptAPIPkg, "TestPlutoTVMovieIsOneEpisodeAndAsksNothing")),
	pm("a listing may take forever", ptAPIFile, "plutoShow",
		"context.WithTimeout(context.Background(), plutoListBudget)", "context.WithTimeout(context.Background(), time.Nanosecond)",
		guard(ptAPIPkg, "TestPlutoTVSeriesAsksForASeasonAndListsIt")),
	pm("the quality never reaches the stream", ptAPIFile, "ResolvePlutoTVStream",
		"PlutoTVClientFn().Stream(ctx, episode.URL, quality)", `PlutoTVClientFn().Stream(ctx, episode.URL, "best")`,
		guard(ptAPIPkg, "TestResolvePlutoTVStreamHonoursTheQuality")),
	pm("the media host is sent a referer", ptAPIFile, "ResolvePlutoTVStream",
		"return source.Stream{URL: stream.URL}, nil", `return source.Stream{URL: stream.URL, Referer: "https://pluto.tv/"}, nil`,
		guard(ptAPIPkg, "TestResolvePlutoTVStreamReturnsTheTitlesOwnPlaylist")),
	pm("a stream may take forever", ptAPIFile, "ResolvePlutoTVStream",
		"context.WithTimeout(context.Background(), plutoStreamBudget)", "context.WithTimeout(context.Background(), time.Nanosecond)",
		guard(ptAPIPkg, "TestResolvePlutoTVStreamReturnsTheTitlesOwnPlaylist")),

	// ---- the provider
	pm("the provider has another priority", ptProvidersFile, "plutoTVProvider.Describe",
		"Priority:    49,", "Priority:    50,",
		guard(ptProvidersPkg, "TestPlutoTVPriorityLeavesTheOtherSourcesWhereTheyWere", "TestPlutoTVDescribesItself")),
	pm("the provider has another tag", ptProvidersFile, "plutoTVProvider.Describe",
		`Tags:        []string{"[plutotv]"},`, `Tags:        []string{"[pluto]"},`,
		guard(ptProvidersPkg, "TestPlutoTVDescribesItself", "TestPlutoTVResolves")),
	pm("the provider has no seasons", ptProvidersFile, "plutoTVProvider.HasSeasons",
		"func (p *plutoTVProvider) HasSeasons() bool { return true }", "func (p *plutoTVProvider) HasSeasons() bool { return false }",
		guard(ptProvidersPkg, "TestPlutoTVDescribesItself", "TestPlutoTVIsRegisteredAndSeasoned")),
	pm("the provider builds another source's adapter", ptProvidersFile, "plutoTVProvider.scraper",
		"lazyGetAdapter(&p.once, &p.adapter, scraper.PlutoTVType)", "lazyGetAdapter(&p.once, &p.adapter, scraper.PomfyType)",
		guard(ptProvidersPkg, "TestPlutoTVBuildsItsOwnAdapterOnce")),
	pm("a search result is tagged as another source's", ptProvidersFile, "plutoTVProvider.Search",
		"tagResults(results, source.PlutoTV)", "tagResults(results, source.Pomfy)",
		guard(ptProvidersPkg, "TestPlutoTVSearchTagsWhatTheAdapterFoundAndSearchesWithTheCallersContext", "TestPlutoTVSearchWorksWithAnAdapterThatTakesNoContext")),
	pm("a search leaves the callers deadline behind", ptProvidersFile, "plutoTVProvider.Search",
		"if ca, ok := adapter.(scraper.ContextualScraper); ok {\n\t\tresults, err = ca.SearchAnimeContext(ctx, query)\n\t} else {\n\t\tresults, err = adapter.SearchAnime(query)\n\t}\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\ttagResults(results, source.PlutoTV)",
		"results, err = adapter.SearchAnime(query)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\ttagResults(results, source.PlutoTV)",
		guard(ptProvidersPkg, "TestPlutoTVSearchTagsWhatTheAdapterFoundAndSearchesWithTheCallersContext")),
	pm("a failed search is tagged anyway", ptProvidersFile, "plutoTVProvider.Search",
		"\tif err != nil {\n\t\treturn nil, err\n\t}\n\ttagResults(results, source.PlutoTV)", "\ttagResults(results, source.PlutoTV)\n\tif err != nil {\n\t\treturn results, err\n\t}",
		guard(ptProvidersPkg, "TestPlutoTVSearchPassesAFailureOnWithoutResults")),
	pm("listing reaches the service on a cancelled context", ptProvidersFile, "plutoTVProvider.FetchEpisodes",
		"func (p *plutoTVProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn nil, err\n\t}\n",
		"func (p *plutoTVProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {\n",
		guard(ptProvidersPkg, "TestPlutoTVNeverReachesTheServiceOnACancelledContext")),
	pm("the provider lists another title", ptProvidersFile, "plutoTVProvider.FetchEpisodes",
		"return plutoTVEpisodesFn(anime)", "return plutoTVEpisodesFn(&models.Anime{})",
		guard(ptProvidersPkg, "TestPlutoTVFetchEpisodesDelegates")),
	pm("the quality is lost on the way to the resolver", ptProvidersFile, "plutoTVProvider.FetchStream",
		"return plutoTVStreamFn(anime, episode, quality)", `return plutoTVStreamFn(anime, episode, "best")`,
		guard(ptProvidersPkg, "TestPlutoTVFetchStreamDelegates")),
	pm("a stream is resolved on a cancelled context", ptProvidersFile, "plutoTVProvider.FetchStream",
		"func (p *plutoTVProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n",
		"func (p *plutoTVProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n",
		guard(ptProvidersPkg, "TestPlutoTVNeverReachesTheServiceOnACancelledContext")),
	pm("the stream URL is not published", ptProvidersFile, "plutoTVProvider.FetchStreamURL",
		"func (p *plutoTVProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n\treturn fetchAndPublish(ctx, p, episode, anime, quality)",
		"func (p *plutoTVProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err",
		guard(ptProvidersPkg, "TestPlutoTVFetchStreamURLPublishesAndClearsWhatTheLastStreamLeft")),
	pm("Pluto TV titles are cleaned like Portuguese ones", ptTaggingFile, "tagResults",
		"kind == source.Anroll || kind == source.AnimesFHD\n", "kind == source.Anroll || kind == source.AnimesFHD || kind == source.PlutoTV\n",
		guard(ptProvidersPkg, "TestTagResultsForPlutoTV")),
	pm("the source's tag is not cleaned from a title", ptNamingFile, "var:tagPattern",
		"AnimesFHD|PlutoTV|StartFlix", "AnimesFHD|StartFlix",
		guard(ptProvidersPkg, "TestPlutoTVNamesAreCleanedLikeTheOthers")),

	// ---- the adapter and the scraper
	pm("the adapter reports another source", ptManagerFile, "PlutoTVAdapter.GetStreamURLContext",
		`map[string]string{"source": "plutotv"}`, `map[string]string{"source": "pluto"}`,
		guard(ptScraperPkg, "TestPlutoTVAdapterStreamHonoursTheQualityOption")),
	pm("the adapter's quality never reaches the stream", ptManagerFile, "PlutoTVAdapter.GetStreamURLContext",
		"a.client.Stream(ctx, episodeURL, qualityOption(options))", `a.client.Stream(ctx, episodeURL, "best")`,
		guard(ptScraperPkg, "TestPlutoTVAdapterStreamHonoursTheQualityOption")),
	pm("the adapter builds a client of its own", ptManagerFile, "NewAdapter",
		"return &PlutoTVAdapter{client: plutotv.Shared()}, nil", "return &PlutoTVAdapter{client: plutotv.NewClient()}, nil",
		guard(ptScraperPkg, "TestPlutoTVIsBuiltByNewAdapterAndNamedLikeTheOthers")),
	pm("the adapter is named another way", ptManagerFile, "scraperDisplayName",
		"case PlutoTVType:\n\t\treturn plutotv.SourceName", "case PlutoTVType:\n\t\treturn \"Pluto\"",
		guard(ptScraperPkg, "TestPlutoTVIsBuiltByNewAdapterAndNamedLikeTheOthers")),
	pm("the adapter's results are tagged Portuguese", ptManagerFile, "scraperLanguageTag",
		"\t\treturn \"[Multilanguage]\"\n\tcase PomfyType:", "\t\treturn \"[PT-BR]\"\n\tcase PomfyType:",
		guard(ptScraperPkg, "TestPlutoTVIsBuiltByNewAdapterAndNamedLikeTheOthers")),
	pm("Pluto TV is not health-checked", ptHealthFile, "healthTargets",
		"PomfyType, AnimesFHDType, PlutoTVType}", "PomfyType, AnimesFHDType}",
		guard(ptScraperPkg, "TestPlutoTVIsHealthChecked")),
	pm("Pluto TV is health-checked with an anime", ptHealthFile, "DefaultHealthCheckQuery",
		"case StartFlixType, TopCineType, PobreflixType, PomfyType, PlutoTVType:", "case StartFlixType, TopCineType, PobreflixType, PomfyType:",
		guard(ptScraperPkg, "TestPlutoTVIsHealthChecked")),

	// ---- where it is plugged in
	pm("the pluto selector is not accepted", ptEnhancedFile, "registryKindsForSource",
		"case \"plutotv\", \"pluto\":\n\t\treturn []apisource.SourceKind{apisource.PlutoTV}", "case \"plutotv\":\n\t\treturn []apisource.SourceKind{apisource.PlutoTV}",
		guard(ptAPIPkg, "TestRegistryKindsForPlutoTV")),
	pm("the movie selector skips Pluto TV", ptEnhancedFile, "registryKindsForSource",
		"apisource.Pobreflix, apisource.Pomfy, apisource.PlutoTV}", "apisource.Pobreflix, apisource.Pomfy}",
		guard(ptAPIPkg, "TestRegistryKindsForPlutoTV")),
	pm("a result's URL does not name Pluto TV", ptEnhancedFile, "identifyAnimeSources",
		`strings.Contains(lowerURL, "pluto.tv")`, `strings.Contains(lowerURL, "plutotv.net")`,
		guard(ptAPIPkg, "TestIdentifyAnimeSourcesNamesPlutoTV")),
	pm("the selector does not name Pluto TV", ptEnhancedFile, "identifyAnimeSources",
		"case \"plutotv\", \"pluto\":\n\t\t\t\tanime.Source = \"PlutoTV\"", "case \"plutotv\":\n\t\t\t\tanime.Source = \"PlutoTV\"",
		guard(ptAPIPkg, "TestIdentifyAnimeSourcesNamesPlutoTV")),
	pm("Pluto TV results are counted as AnimesFHD's", ptEnhancedFile, "countSourceBreakdown",
		"b.PlutoTV++", "b.AnimesFHD++",
		guard(ptAPIPkg, "TestCountSourceBreakdownCountsPlutoTV")),
	pm("a CONTINUE row of Pluto TV reopens everywhere", ptContinueFile, "continueSourceToken",
		`"animedrive", "animesfhd", "plutotv":`, `"animedrive", "animesfhd":`,
		guard(ptHandlersPkg, "TestAContinueRowOfPlutoTVReopensOnPlutoTV")),
	pm("Pluto TV opens on the first tab", ptHomeFile, "scopeIndexForSource",
		`case "movietv", "startflix", "topcine", "pobreflix", "plutotv":`, `case "movietv", "startflix", "topcine", "pobreflix":`,
		guard(ptHandlersPkg, "TestPlutoTVOpensOnTheMovieAndSeriesTab")),
	pm("a saved Pluto TV URL names no source", ptHomeFile, "sourceFromURL",
		"case strings.Contains(host, \"pluto.tv\"):\n\t\treturn \"PlutoTV\"", "case strings.Contains(host, \"pluto.tv\"):\n\t\treturn \"plutotv\"",
		guard(ptHandlersPkg, "TestASavedPlutoTVURLNamesItsSource")),
	pm("a retried Pluto TV search opens on the first tab", ptTryFile, "indexForScope",
		`case "movietv", "startflix", "topcine", "pobreflix", "plutotv":`, `case "movietv", "startflix", "topcine", "pobreflix":`,
		guard(ptTUIPkg, "TestPlutoTVScopeIsTheMovieTab")),
	pm("a Pluto TV series lists through the panel", ptWorkflowFile, "seriesListingRoute",
		"\t\treturn \"plutotv\"\n", "\t\treturn \"panel\"\n",
		guard(ptDownloadPkg, "TestPlutoTVSeriesListThroughPlutoTV")),
	pm("a Pluto TV series' seasons come from Pomfy", ptWorkflowFile, "seriesSeasonsFor",
		"case \"plutotv\":\n\t\treturn api.GetPlutoTVSeasonNumbers", "case \"plutotv\":\n\t\treturn api.GetPomfySeasonNumbers",
		guard(ptDownloadPkg, "TestPlutoTVSeriesListThroughPlutoTV")),
	pm("a Pluto TV series' episodes come from Pomfy", ptWorkflowFile, "seriesEpisodesFor",
		"case \"plutotv\":\n\t\treturn api.GetPlutoTVSeasonEpisodes", "case \"plutotv\":\n\t\treturn api.GetPomfySeasonEpisodes",
		guard(ptDownloadPkg, "TestPlutoTVSeriesListThroughPlutoTV")),
	pm("the SDK knows Pluto TV by one spelling", ptTypesFile, "var:sourceTable",
		`[]string{"plutotv", "pluto"}`, `[]string{"plutotv"}`,
		guard(ptTypesPkg, "TestPlutoTVParsesFromEverySpellingItAnswersTo")),
	pm("a Pluto TV title has no episode flow of its own", ptMediaFile, "HasInteractiveEpisodeFlow",
		`m.Source == "AnimesFHD" || m.Source == "PlutoTV" ||`, `m.Source == "AnimesFHD" ||`,
		guard(ptModelsPkg, "TestPlutoTVHasAnInteractiveEpisodeFlowBySource")),
}

func plutoMutantsAsMutants(ms []plutoMutant) []mutant {
	out := make([]mutant, len(ms))
	for i, m := range ms {
		out[i] = m.mutant
	}
	return out
}

func TestPlutoTVGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, plutoMutantsAsMutants(plutoMutants))
}

// ---- the table's own checks ------------------------------------------------

func lf(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

// declRange is where a function, a method ("Type.Method"), or a package-level
// declaration (name "var:x" or "const:x") starts and ends in src. A method
// named without its type is found only if no other declaration in src has the
// name; otherwise ambiguous is set and the mutant must name its type.
func declRange(t *testing.T, src, name string) (start, end int, found, ambiguous bool) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.SkipObjectResolution)
	require.NoError(t, err)
	kind, ident, _ := strings.Cut(name, ":")
	if ident == "" {
		kind, ident = "func", name
	}
	recv := ""
	if kind == "func" {
		if r, m, ok := strings.Cut(ident, "."); ok {
			recv, ident = r, m
		}
	}
	type span struct{ start, end int }
	var matches []span
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if kind == "func" && d.Name.Name == ident && (recv == "" || receiverName(d) == recv) {
				matches = append(matches, span{fset.Position(d.Pos()).Offset, fset.Position(d.End()).Offset})
			}
		case *ast.GenDecl:
			if (kind == "var" && d.Tok == token.VAR) || (kind == "const" && d.Tok == token.CONST) {
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, n := range vs.Names {
							if n.Name == ident {
								matches = append(matches, span{fset.Position(d.Pos()).Offset, fset.Position(d.End()).Offset})
							}
						}
					}
				}
			}
		}
	}
	switch len(matches) {
	case 0:
		return 0, 0, false, false
	case 1:
		return matches[0].start, matches[0].end, true, false
	default:
		return 0, 0, false, true
	}
}

// receiverName is the type a method is declared on, "" for a plain function.
func receiverName(d *ast.FuncDecl) string {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return ""
	}
	t := d.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// testsOf is the names of the Test functions of the package at dir.
func testsOf(t *testing.T, dir string) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	tests := map[string]bool{}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		for _, name := range declaredFuncs(t, filepath.Join(dir, e.Name())) {
			if strings.HasPrefix(name, "Test") {
				tests[name] = true
			}
		}
	}
	return tests
}

// mutantTableProblems is everything wrong with a mutant table, as sentences.
// A mutant that is wrong cannot be trusted to kill anything: one whose
// snippet moved would fail to apply, one whose snippet is in another function
// would be counted against the wrong one, one that names a guard that no
// longer exists would "be killed" by nothing.
func mutantTableProblems(t *testing.T, root string, ms []plutoMutant) []string {
	t.Helper()
	var problems []string
	add := func(m plutoMutant, format string, args ...any) {
		problems = append(problems, m.name+": "+strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf(format, args...), "\n", " ")))
	}
	names := map[string]bool{}
	testCache := map[string]map[string]bool{}
	for _, m := range ms {
		if strings.TrimSpace(m.name) == "" {
			add(m, "has no name")
		}
		if names[m.name] {
			add(m, "its name is used twice")
		}
		names[m.name] = true
		if len(m.edits) == 0 {
			add(m, "edits nothing")
		}
		if m.fn == "" {
			add(m, "does not say which function it breaks")
		}
		if len(m.kills) == 0 {
			add(m, "names no guard test")
		}
		for _, e := range m.edits {
			if e.from == e.to {
				add(m, "replaces a snippet with itself")
			}
			if e.file != m.file {
				add(m, "edits %s but is filed under %s", e.file, m.file)
			}
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(e.file))) // #nosec G304 -- a module file named in this test
			if err != nil {
				add(m, "its file %s cannot be read", e.file)
				continue
			}
			src := lf(string(raw))
			from := lf(e.from)
			switch n := strings.Count(src, from); {
			case n == 0:
				add(m, "its snippet is not in %s", e.file)
			case n > 1:
				add(m, "its snippet is in %s %d times, not once", e.file, n)
			default:
				start, end, found, ambiguous := declRange(t, src, m.fn)
				at := strings.Index(src, from)
				switch {
				case ambiguous:
					add(m, "%s names %s more than once: name its type, as Type.%s", e.file, m.fn, m.fn)
				case !found:
					add(m, "%s declares no %s", e.file, m.fn)
				case at < start || at+len(from) > end:
					add(m, "its snippet is outside %s", m.fn)
				}
			}
		}
		for pkg, tests := range m.kills {
			for _, name := range tests {
				if strings.HasPrefix(name, "TestLive") {
					add(m, "its guard %s is a live test, skipped unless GOANIME_LIVE is set", name)
				}
			}
			dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(pkg, "./")))
			if _, ok := testCache[dir]; !ok {
				if st, err := os.Stat(dir); err != nil || !st.IsDir() {
					add(m, "its guard package %s is not a directory", pkg)
					testCache[dir] = nil
					continue
				}
				testCache[dir] = testsOf(t, dir)
			}
			if len(tests) == 0 {
				add(m, "names no guard test in %s", pkg)
			}
			for _, name := range tests {
				if testCache[dir] != nil && !testCache[dir][name] {
					add(m, "its guard %s is not a test of %s", name, pkg)
				}
			}
		}
	}
	return problems
}

func TestPlutoTVMutantsAreWellFormed(t *testing.T) {
	problems := mutantTableProblems(t, moduleRoot(t), plutoMutants)
	assert.Empty(t, problems, "a mutant that is not well formed cannot be trusted to kill anything")
	assert.GreaterOrEqual(t, len(plutoMutants), 100, "the table has shrunk")
}

// The check above is only worth running if it sees each defect. These tables
// carry one defect each over a module of their own.
func TestTheMutantCheckSeesEveryDefect(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write("go.mod", "module example.test\n\ngo 1.22\n")
	write("lib/lib.go", "package lib\n\nvar table = map[string]int{\"a\": 1}\n\nconst limit = 6\n\nfunc One() int {\n\treturn 1 + 10\n}\n\nfunc Two() int {\n\treturn 2 + 2\n}\n\nfunc Three() int {\n\treturn 2 + 2\n}\n\ntype A struct{}\n\ntype B struct{}\n\nfunc (A) Same() int {\n\treturn 3 + 30\n}\n\nfunc (B) Same() int {\n\treturn 4 + 40\n}\n\nfunc Dup() int {\n\ta := 5\n\treturn a + a\n}\n")
	write("lib/lib_test.go", "package lib\n\nimport \"testing\"\n\nfunc TestOne(t *testing.T) {}\n\nfunc TestLiveOne(t *testing.T) {}\n\nfunc helper() {}\n")
	write("crlf/crlf.go", "package crlf\r\n\r\nfunc A() int {\r\n\treturn 1 +\r\n\t\t2\r\n}\r\n")
	write("crlf/crlf_test.go", "package crlf\r\n\r\nimport \"testing\"\r\n\r\nfunc TestA(t *testing.T) {}\r\n")

	const lib, crlf = "lib/lib.go", "crlf/crlf.go"
	good := pm("good", lib, "One", "return 1 + 10", "return 1 + 2", guard("./lib", "TestOne"))
	require.Empty(t, mutantTableProblems(t, root, []plutoMutant{good}), "the control: a table with no defect has no problems")

	with := func(mutate func(*plutoMutant)) plutoMutant {
		m := good
		m.name = "defective"
		mutate(&m)
		return m
	}
	cases := []struct {
		name string
		m    plutoMutant
		want string
	}{
		{"no name", with(func(m *plutoMutant) { m.name = "  " }), "has no name"},
		{"no edits", with(func(m *plutoMutant) { m.edits = nil }), "edits nothing"},
		{"no function", with(func(m *plutoMutant) { m.fn = "" }), "does not say which function"},
		{"no guards", with(func(m *plutoMutant) { m.kills = nil }), "names no guard test"},
		{"an empty guard list", with(func(m *plutoMutant) { m.kills = map[string][]string{"./lib": nil} }), "names no guard test in ./lib"},
		{"a snippet replaced by itself", with(func(m *plutoMutant) { m.edits = []codeEdit{{lib, "return 1 + 10", "return 1 + 10"}} }), "with itself"},
		{"a mutant filed under another file", with(func(m *plutoMutant) { m.file = crlf }), "but is filed under"},
		{"a snippet that moved", with(func(m *plutoMutant) { m.edits = []codeEdit{{lib, "return 7 + 7", "return 1"}} }), "snippet is not in"},
		{"a snippet that is ambiguous", with(func(m *plutoMutant) { m.edits = []codeEdit{{lib, "return 2 + 2", "return 4"}}; m.fn = "Two" }), "2 times, not once"},
		{"a snippet in the wrong function", with(func(m *plutoMutant) { m.edits = []codeEdit{{lib, "a + a", "a"}}; m.fn = "One" }), "outside One"},
		{"a function that does not exist", with(func(m *plutoMutant) { m.edits = []codeEdit{{lib, "a + a", "a"}}; m.fn = "Gone" }), "declares no Gone"},
		{"a file that does not exist", with(func(m *plutoMutant) { m.file = "lib/nope.go"; m.edits = []codeEdit{{"lib/nope.go", "x", "y"}} }), "cannot be read"},
		{"a guard that is not a test", with(func(m *plutoMutant) { m.kills = guard("./lib", "helper") }), "is not a test of"},
		{"a guard that does not exist", with(func(m *plutoMutant) { m.kills = guard("./lib", "TestNope") }), "is not a test of"},
		{"a live test as a guard", with(func(m *plutoMutant) { m.kills = guard("./lib", "TestLiveOne") }), "is a live test"},
		{"a guard in the wrong package", with(func(m *plutoMutant) { m.kills = guard("./crlf", "TestOne") }), "is not a test of ./crlf"},
		{"a guard package that does not exist", with(func(m *plutoMutant) { m.kills = guard("./nowhere", "TestOne") }), "is not a directory"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problems := mutantTableProblems(t, root, []plutoMutant{tc.m})
			require.NotEmpty(t, problems, "the defect went unseen")
			assert.Contains(t, strings.Join(problems, "\n"), tc.want)
		})
	}

	t.Run("a method is named by its type when its name is not unique", func(t *testing.T) {
		bare := pm("bare", lib, "Same", "return 3 + 30", "return 0", guard("./lib", "TestOne"))
		assert.Contains(t, strings.Join(mutantTableProblems(t, root, []plutoMutant{bare}), "\n"), "more than once")
		typed := pm("typed", lib, "A.Same", "return 3 + 30", "return 0", guard("./lib", "TestOne"))
		assert.Empty(t, mutantTableProblems(t, root, []plutoMutant{typed}))
		other := pm("other", lib, "A.Same", "return 4 + 40", "return 0", guard("./lib", "TestOne"))
		assert.Contains(t, strings.Join(mutantTableProblems(t, root, []plutoMutant{other}), "\n"), "outside A.Same", "the snippet of B.Same is not A.Same's")
	})
	t.Run("two mutants with one name", func(t *testing.T) {
		problems := mutantTableProblems(t, root, []plutoMutant{good, good})
		assert.Contains(t, strings.Join(problems, "\n"), "used twice")
	})
	t.Run("a package-level declaration is found", func(t *testing.T) {
		v := pm("v", lib, "var:table", `"a": 1`, `"a": 2`, guard("./lib", "TestOne"))
		c := pm("c", lib, "const:limit", "limit = 6", "limit = 7", guard("./lib", "TestOne"))
		assert.Empty(t, mutantTableProblems(t, root, []plutoMutant{v, c}))
		wrong := pm("w", lib, "var:limit", "limit = 6", "limit = 7", guard("./lib", "TestOne"))
		assert.Contains(t, strings.Join(mutantTableProblems(t, root, []plutoMutant{wrong}), "\n"), "declares no var:limit", "a const is not a var")
	})
	t.Run("a snippet with unix line endings is found in a windows checkout", func(t *testing.T) {
		m := pm("crlf", crlf, "A", "return 1 +\n\t\t2", "return 3", guard("./crlf", "TestA"))
		assert.Empty(t, mutantTableProblems(t, root, []plutoMutant{m}))
	})
}

// ---- every function has a mutant ------------------------------------------

// plutoTVSurface lists the functions of the feature that live in files that
// hold nothing else. Each is mutated, or has a written reason it is not.
var plutoTVSurface = []struct {
	pkg       string // module-relative directory
	file      string
	funcs     []string
	unmutated map[string]string // function → why no mutant breaks it
}{
	{"internal/scraper/providers/plutotv", "client.go", []string{
		"Shared", "NewClient", "NewClientForTest", "newClientID", "SeriesURL", "MovieURL", "EpisodeURL", "URL", "Year",
		"languageTag", "ToAnimeModel", "playID", "parsePageURL", "IsMovieURL", "trusted", "currentSession", "dropSession",
		"boot", "withSession", "Search", "Show", "Stream", "isProtected", "parseMaster", "bandwidthBudget",
		"orderVariants", "contentPlaylistURL", "checkMediaPlaylist", "get", "getJSON", "getText",
	}, map[string]string{}},
	{"internal/api", "plutotv.go", []string{
		"GetPlutoTVEpisodes", "GetPlutoTVSeasonNumbers", "GetPlutoTVSeasonEpisodes", "plutoShow", "plutoMovieEpisode",
		"plutoSeasonEpisodes", "ResolvePlutoTVStream",
	}, map[string]string{
		"GetPlutoTVSeasonEpisodes": "three guard lines (a nil title, the catalog, plutoSeasonEpisodes) with nothing of its own to break; its behaviour is plutoSeasonEpisodes's",
	}},
}

func TestPlutoTVHasNoFunctionThatIsNotListed(t *testing.T) {
	root := moduleRoot(t)
	for _, entry := range plutoTVSurface {
		declared := declaredFuncs(t, filepath.Join(root, filepath.FromSlash(entry.pkg), entry.file))
		want := append([]string(nil), entry.funcs...)
		sort.Strings(want)
		assert.Equal(t, want, declared, "%s/%s holds only this feature: a function added to it must be listed in plutoTVSurface, where its mutant and its coverage will be demanded", entry.pkg, entry.file)
	}
}

func TestPlutoTVFunctionsAreAllMutated(t *testing.T) {
	for _, entry := range plutoTVSurface {
		path := entry.pkg + "/" + entry.file
		mutated := map[string]int{}
		for _, m := range plutoMutants {
			if m.file == path {
				mutated[m.fn]++
			}
		}
		for _, fn := range entry.funcs {
			reason, exempt := entry.unmutated[fn]
			switch {
			case exempt && mutated[fn] > 0:
				t.Errorf("%s %s has %d mutants and a reason it has none (%q): drop the reason", path, fn, mutated[fn], reason)
			case exempt && strings.TrimSpace(reason) == "":
				t.Errorf("%s %s is exempt from mutation without a reason", path, fn)
			case !exempt && mutated[fn] == 0:
				t.Errorf("%s %s has no mutant: break one decision in it and name the test that must fail", path, fn)
			}
		}
		for fn := range entry.unmutated {
			assert.Contains(t, entry.funcs, fn, "%s: %s is exempt but is not a function of the file", path, fn)
		}
	}
}

// The same check, over a table of its own, so a regression in the check
// itself does not wave the real one through.
func TestTheSurfaceCheckSeesAFunctionWithoutAMutant(t *testing.T) {
	countMutants := func(ms []plutoMutant, file, fn string) int {
		n := 0
		for _, m := range ms {
			if m.file == file && m.fn == fn {
				n++
			}
		}
		return n
	}
	ms := []plutoMutant{pm("a", "f.go", "One", "x", "y", nil), pm("b", "f.go", "One", "x", "z", nil), pm("c", "g.go", "Two", "x", "y", nil)}
	assert.Equal(t, 2, countMutants(ms, "f.go", "One"))
	assert.Equal(t, 0, countMutants(ms, "f.go", "Two"), "a function filed under another file is not counted")
	assert.Equal(t, 0, countMutants(ms, "f.go", "Three"))
}

// ---- every function is covered ---------------------------------------------

func TestPlutoTVFunctionsAreFullyCovered(t *testing.T) {
	skipMutantsWhenShort(t)
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH")
	}
	root := moduleRoot(t)
	for _, entry := range plutoTVSurface {
		profile := filepath.Join(t.TempDir(), "cover.out")
		args := []string{"test", "-count=1", "-coverprofile=" + profile}
		if entry.pkg == "internal/api" {
			// Only this feature's own tests: the package's others would
			// cover its functions by accident, and not this guard itself, which would
			// run it again without end.
			args = append(args, "-run", "PlutoTV", "-skip", "KillsMutants|FullyCovered")
		}
		run := exec.Command(goBin, append(args, "./"+entry.pkg)...) // #nosec G204 -- the go toolchain, with this test's own arguments
		run.Dir = root
		out, err := run.CombinedOutput()
		require.NoError(t, err, "the package tests must pass to be measured:\n%s", out)

		report := exec.Command(goBin, "tool", "cover", "-func="+profile) // #nosec G204
		report.Dir = root
		table, err := report.CombinedOutput()
		require.NoError(t, err, string(table))
		assert.Empty(t, notFullyCovered(string(table), entry.file, entry.funcs), "%s/%s: functions of this feature the tests do not fully cover", entry.pkg, entry.file)
	}
}
