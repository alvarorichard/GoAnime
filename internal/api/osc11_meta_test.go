package api

import "testing"

// Tests of the tests for the terminal background probe (internal/tui/osc11.go).
//
// The probe once left its read of stdin running when the terminal did not
// answer, and that read took the user's next keys: in tmux the first episode
// picker looked frozen (2026-11-21). Each mutant below puts one part of that
// back; the guard tests it names must fail on it.

const (
	osc11TUIPkg = "./internal/tui"
	osc11File   = "internal/tui/osc11.go"
)

var osc11Mutants = []mutant{
	{
		name:  "the pre-fix probe: a silent terminal's read is left running",
		edits: []codeEdit{{osc11File, "\t\tin.Cancel()\n\t\tselect {\n\t\tcase <-ch: // the read returned: nothing is left reading the terminal\n\t\tcase <-time.After(osc11Drain):\n\t\t}\n", ""}},
		kills: map[string][]string{osc11TUIPkg: {"TestOSC11SilentTerminalLeavesNothingReading", "TestOSC11LateReplyAfterTheTimeoutIsNotReadByTheProbe"}},
	},
	{
		name:  "the read is cancelled but not waited for",
		edits: []codeEdit{{osc11File, "\t\tcase <-ch: // the read returned: nothing is left reading the terminal\n\t\tcase <-time.After(osc11Drain):\n", "\t\tcase <-time.After(0):\n"}},
		kills: map[string][]string{osc11TUIPkg: {"TestOSC11SilentTerminalLeavesNothingReading"}},
		race:  false,
	},
	{
		name:  "the probe waits out its whole drain on a silent terminal",
		edits: []codeEdit{{osc11File, "\t\tcase <-ch: // the read returned: nothing is left reading the terminal\n", "\t\tcase <-make(chan result): // the read returned: nothing is left reading the terminal\n"}},
		kills: map[string][]string{osc11TUIPkg: {"TestOSC11SilentTerminalLeavesNothingReading"}},
	},
	{
		name:  "the probe reads on past the reply",
		edits: []codeEdit{{osc11File, "\t\t\tif c == '\\a' { // BEL ends the reply\n\t\t\t\tch <- result{sb.String(), true}\n\t\t\t\treturn\n", "\t\t\tif c == '\\a' { // BEL ends the reply\n\t\t\t\tch <- result{sb.String(), true}\n"}},
		kills: map[string][]string{osc11TUIPkg: {"TestOSC11ProbeDoesNotReadPastTheReply"}},
	},
	{
		name:  "the ST that ends a reply is not seen",
		edits: []codeEdit{{osc11File, "if prevESC && c == '\\\\' { // ST ends the reply", "if prevESC && c == 0 { // ST ends the reply"}},
		kills: map[string][]string{osc11TUIPkg: {"TestOSC11ReplyIsRead"}},
	},
}

func TestOSC11GuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, osc11Mutants)
}
