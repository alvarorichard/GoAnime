package tui

import (
	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"charm.land/lipgloss/v2"
)

// FormTheme dresses huh forms (the search prompt, confirmations, download
// ranges) in the same theme as the list screens: a crimson bar beside the
// focused field, the title in the accent color, buttons drawn like the
// footer's keys.
//
// Like the list screens it follows the detected theme. huh asks the terminal
// for its background, but RunClean suppresses that query, so without
// detection huh would guess light on what is almost always a dark terminal.
func FormTheme() huh.Theme {
	return huh.ThemeFunc(func(bool) *huh.Styles { return formStyles() })
}

// SpinnerTheme is the matching theme for huh spinners.
func SpinnerTheme() spinner.Theme {
	return spinner.ThemeFunc(func(bool) *spinner.Styles {
		t := CurrentTheme()
		return &spinner.Styles{Spinner: t.Primary.UnsetBold(), Title: t.Text}
	})
}

func formStyles() *huh.Styles {
	t := CurrentTheme()
	s := huh.ThemeBase(IsDark())
	accent := t.Primary.UnsetBold()

	selector, checked, unchecked := "▶ ", "✓ ", "• "
	next, prev := "→", "←"
	if Plain() {
		selector, checked, unchecked = "> ", "x ", "- "
		next, prev = ">", "<"
	}

	s.Focused.Base = s.Focused.Base.BorderForeground(t.Shadow.GetForeground())
	s.Focused.Card = s.Focused.Base
	s.Focused.Title = t.Primary
	s.Focused.NoteTitle = t.Primary.MarginBottom(1)
	s.Focused.Description = t.Muted
	s.Focused.ErrorIndicator = accent.SetString(" *")
	s.Focused.ErrorMessage = accent.SetString(" *")
	s.Focused.SelectSelector = accent.SetString(selector)
	s.Focused.NextIndicator = accent.MarginLeft(1).SetString(next)
	s.Focused.PrevIndicator = accent.MarginRight(1).SetString(prev)
	s.Focused.Option = t.Text
	s.Focused.MultiSelectSelector = accent.SetString(selector)
	s.Focused.SelectedOption = accent
	s.Focused.SelectedPrefix = accent.SetString(checked)
	s.Focused.UnselectedPrefix = t.Muted.SetString(unchecked)
	s.Focused.UnselectedOption = t.Text
	s.Focused.FocusedButton = t.Selected.Padding(0, 2).MarginRight(1)
	s.Focused.Next = s.Focused.FocusedButton
	s.Focused.BlurredButton = t.Keycap.Padding(0, 2).MarginRight(1)
	s.Focused.TextInput.Cursor = accent
	s.Focused.TextInput.Placeholder = t.Muted
	s.Focused.TextInput.Prompt = accent
	s.Focused.TextInput.Text = t.Text

	s.Blurred = s.Focused
	s.Blurred.Base = s.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	s.Blurred.Card = s.Blurred.Base
	s.Blurred.NextIndicator = lipgloss.NewStyle()
	s.Blurred.PrevIndicator = lipgloss.NewStyle()

	s.Group.Title = s.Focused.Title
	s.Group.Description = s.Focused.Description

	s.Help.ShortKey = t.Soft
	s.Help.ShortDesc = t.Muted
	s.Help.ShortSeparator = t.Faint
	s.Help.FullKey = t.Soft
	s.Help.FullDesc = t.Muted
	s.Help.FullSeparator = t.Faint
	s.Help.Ellipsis = t.Faint
	return s
}
