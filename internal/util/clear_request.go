package util

// ClearRequest is what --clear-continue, --clear-recent and --clear-history
// ask for: which of the SEARCH home screen's two lists to empty, and whether
// the question may be skipped (--yes).
type ClearRequest struct {
	Continue bool // empty the CONTINUE list (watch history)
	Recent   bool // empty the RECENT searches
	Yes      bool // do not ask first
}

// GlobalClearRequest is the request FlagParser stored for ErrClearRequested.
var GlobalClearRequest *ClearRequest

// clearRequestFrom turns the flags into a request: --clear-history is both
// lists. It is nil when none of the three was given, and --yes alone asks for
// nothing.
func clearRequestFrom(continueList, recent, history, yes bool) *ClearRequest {
	if !continueList && !recent && !history {
		return nil
	}
	return &ClearRequest{
		Continue: continueList || history,
		Recent:   recent || history,
		Yes:      yes,
	}
}
