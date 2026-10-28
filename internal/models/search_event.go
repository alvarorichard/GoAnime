package models

import "time"

// Live search events stream the source fan-out to the SEARCHING screen
// (tela 2): first the plan (which sources run, which sit out, the give-up
// budget), then one outcome per source as it answers. They live in models,
// the leaf package, so providers, api and tui can all speak them.

// SearchSourceState is one searched source's outcome.
type SearchSourceState int

const (
	// SearchAsking means the source has not answered yet.
	SearchAsking SearchSourceState = iota
	// SearchDone means the source answered (possibly with nothing).
	SearchDone
	// SearchFailed means the source errored or timed out.
	SearchFailed
)

// SearchOffEntry is a wanted source that sits out of the fan-out.
type SearchOffEntry struct {
	Name   string // display name ("Goyabu")
	Reason string // "off by default", "disabled", "cooling down"
}

// SearchEvent is one fan-out message. The first is always the plan
// (Plan=true with Searched/Off/GiveUp); the rest are source outcomes.
type SearchEvent struct {
	Plan     bool
	Searched []string
	Off      []SearchOffEntry
	GiveUp   time.Duration

	Name    string
	State   SearchSourceState
	Results []*Anime
	Count   int
	Elapsed time.Duration
	Reason  string
}
