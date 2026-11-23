package surf

import (
	"context"
	"slices"

	"github.com/enetx/http"
)

// Chromium decrements its separate ALPS reconstruction budget from twenty and
// fails at zero. Unsent preparations therefore permit nineteen meaningful adds.
// Critical-CH keeps its independent once-per-origin navigation budget.
// https://github.com/chromium/chromium/blob/158.0.8127.18/content/browser/loader/navigation_url_loader_impl.cc#L1895
const clientHintALPSPreparationLimit = 19

type clientHintALPSBudgetError struct{}

func (clientHintALPSBudgetError) Error() string {
	return "client hints: too many ALPS request preparations"
}

func (n *clientHintNavigationState) takeALPS() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.alpsPreparations == clientHintALPSPreparationLimit {
		return false
	}

	n.alpsPreparations++

	return true
}

// applyALPS only changes a private unsent navigation. Connection instructions
// add missing configured values, never response-cache grants or body replays.
// https://github.com/chromium/chromium/blob/158.0.8127.18/services/network/accept_ch_frame_interceptor.cc#L25
func (t *clientHintsTransport) applyALPS(
	req *http.Request,
	value string,
	nav *clientHintNavigationState,
) error {
	if err := req.Context().Err(); err != nil {
		return err
	}

	if !clientHintTrustworthy(req.URL) || !isClientHintNavigation(req) {
		return nil
	}

	requested, valid := parseClientHints(value)
	if !valid {
		return nil
	}

	var added []string
	for _, name := range knownClientHintNames {
		if name == "width" || name == "sec-ch-width" || !containsClientHint(requested, name) {
			continue
		}

		if _, available := t.values[name]; !available {
			continue
		}

		if _, present := profileHeaderValues(req.Header, name); !present {
			added = append(added, name)
		}
	}

	if len(added) == 0 {
		return nil
	}

	if !nav.takeALPS() {
		return clientHintALPSBudgetError{}
	}

	state := &clientHintAttempt{navigation: nav, generated: make(http.Header)}
	if previous := previousClientHintAttempt(req); previous != nil {
		*state = *previous
		state.generated = previous.generated.Clone()
		state.alpsAdded = slices.Clone(previous.alpsAdded)
	}

	headers := req.Header.Clone()
	if headers == nil {
		headers = make(http.Header)
	}

	for _, name := range added {
		value := t.values[name]
		headers.Set(name, value)
		state.generated[name] = []string{value}
	}

	state.alpsAdded = append(state.alpsAdded, added...)
	req.Header = headers
	*req = *req.WithContext(context.WithValue(req.Context(), clientHintAttemptKey{}, state))

	return nil
}
