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
// Like the list screens it is always the dark theme. huh asks the terminal
// for its background, but RunClean suppresses that query, so huh would guess
// light and draw crimson text on what is almost always a dark terminal.
func FormTheme() huh.Theme {
	return huh.ThemeFunc(func(bool) *huh.Styles { return formStyles() })
}

// SpinnerTheme is the matching theme for huh spinners.
func SpinnerTheme() spinner.Theme {
	return spinner.ThemeFunc(func(bool) *spinner.Styles {
		t := NewTheme(true)
		return &spinner.Styles{Spinner: t.Primary.UnsetBold(), Title: t.Text}
	})
}

func formStyles() *huh.Styles {
	t := NewTheme(true)
	s := huh.ThemeBase(true)
	accent := t.Primary.UnsetBold()

	s.Focused.Base = s.Focused.Base.BorderForeground(t.Shadow.GetForeground())
	s.Focused.Card = s.Focused.Base
	s.Focused.Title = t.Primary
	s.Focused.NoteTitle = t.Primary.MarginBottom(1)
	s.Focused.Description = t.Muted
	s.Focused.ErrorIndicator = accent.SetString(" *")
	s.Focused.ErrorMessage = accent.SetString(" *")
	s.Focused.SelectSelector = accent.SetString("▶ ")
	s.Focused.NextIndicator = accent.MarginLeft(1).SetString("→")
	s.Focused.PrevIndicator = accent.MarginRight(1).SetString("←")
	s.Focused.Option = t.Text
	s.Focused.MultiSelectSelector = accent.SetString("▶ ")
	s.Focused.SelectedOption = accent
	s.Focused.SelectedPrefix = accent.SetString("✓ ")
	s.Focused.UnselectedPrefix = t.Muted.SetString("• ")
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
