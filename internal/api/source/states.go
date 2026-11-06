package source

import "github.com/alvarorichard/Goanime/v2/internal/util"

// State is one source as the SOURCES screen shows it.
type State struct {
	Kind SourceKind
	// On is whether the source joins the search right now.
	On bool
	// Locked means GOANIME_DISABLED_SOURCES switches it off: the screen cannot
	// change that, only the environment can.
	Locked bool
	// Note says why the source is in its state, or a caveat about it.
	Note string
}

// SourceStates lists every registered source, in resolution order, with its
// search state and the reason for it. A source the environment switched off
// stays in the list, marked Locked, so the user can see why it is missing.
func SourceStates() []State {
	srcs := allByPriority()
	out := make([]State, 0, len(srcs))
	for _, s := range srcs {
		out = append(out, stateOf(s.Describe()))
	}
	return out
}

// stateOf derives the State of one descriptor from the same inputs
// IsSearchEnabled reads, so the screen can never disagree with the search.
func stateOf(d Descriptor) State {
	st := State{Kind: d.Kind, On: IsSearchEnabled(d)}
	name := string(d.Kind)
	pref, hasPref := util.SourcePref(name)
	switch {
	case !IsEnabled(d):
		st.Locked = true
		st.Note = "locked by GOANIME_DISABLED_SOURCES"
	case st.On && util.SourceForceEnabled(name) && hasPref && !pref:
		st.Note = "kept on by GOANIME_ENABLED_SOURCES"
	case st.On:
		st.Note = d.Note
	case hasPref:
		st.Note = joinNote("turned off by you", d.Note)
	default:
		st.Note = joinNote("off by default", d.Note)
	}
	return st
}

func joinNote(reason, note string) string {
	if note == "" {
		return reason
	}
	return reason + "; " + note
}
