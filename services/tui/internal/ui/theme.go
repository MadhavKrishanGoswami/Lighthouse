package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LighthouseTheme struct {
	// AppBackgroundColor is the darkest color, filling the entire background.
	AppBackgroundColor tcell.Color

	// PanelBackgroundColor is for primitives like Flex, Form, and other containers.
	PanelBackgroundColor tcell.Color

	// BorderColor is for unfocused panel borders.
	BorderColor tcell.Color

	// FocusedBorderColor highlights the currently active panel.
	FocusedBorderColor tcell.Color

	// PrimaryTextColor is the main color for text content.
	PrimaryTextColor tcell.Color

	// SecondaryTextColor is for less important text, like placeholders or subtitles.
	SecondaryTextColor tcell.Color

	// TitleColor is used for the titles of windows and panels.
	TitleColor tcell.Color

	// HighlightBackgroundColor is the background for selected items in lists or tables.
	HighlightBackgroundColor tcell.Color

	// AccentGoodColor is for "online", "success", or positive status indicators.
	AccentGoodColor tcell.Color

	// AccentWarningColor is for "pending", "warning", or neutral status indicators.
	AccentWarningColor tcell.Color

	// AccentErrorColor is for "offline", "error", or critical status indicators.
	AccentErrorColor tcell.Color
}

// The global instance of the Lighthouse theme.
var Theme = LighthouseTheme{
	// Backgrounds
	AppBackgroundColor:   tcell.NewHexColor(0x0D1B2A), // Deep Sea Blue
	PanelBackgroundColor: tcell.NewHexColor(0x1B263B), // Stormy Grey

	// Borders
	BorderColor:        tcell.NewHexColor(0x415A77), // Dim Grey
	FocusedBorderColor: tcell.NewHexColor(0xEAE7DC), // Lighthouse Fog (Aged Paper)

	// Text
	PrimaryTextColor:   tcell.NewHexColor(0xE0E1DD), // Off-White
	SecondaryTextColor: tcell.NewHexColor(0x778DA9), // Muted Blue-Grey
	TitleColor:         tcell.NewHexColor(0xEAE7DC), // Lighthouse Fog

	// Highlights & Accents
	HighlightBackgroundColor: tcell.NewHexColor(0x2A4D69), // Wet Rock
	AccentGoodColor:          tcell.NewHexColor(0x6A994E), // Sea Foam Green
	AccentWarningColor:       tcell.NewHexColor(0xF4A261), // Old Brass
	AccentErrorColor:         tcell.NewHexColor(0xE76F51), // Rust Red
}

// ---------------------------------------------------------------- //
// GLOBAL THEME APPLICATION (BONUS)
// ---------------------------------------------------------------- //

// ApplyLighthouseTheme sets the defined theme colors for all default tview components.
// Call this function once in your main application setup.
func ApplyLighthouseTheme() {
	tview.Styles.PrimitiveBackgroundColor = Theme.PanelBackgroundColor
	tview.Styles.ContrastBackgroundColor = Theme.HighlightBackgroundColor
	tview.Styles.MoreContrastBackgroundColor = Theme.AppBackgroundColor

	tview.Styles.BorderColor = Theme.BorderColor
	tview.Styles.TitleColor = Theme.TitleColor
	tview.Styles.GraphicsColor = Theme.BorderColor // For box drawings

	// Text colors
	tview.Styles.PrimaryTextColor = Theme.PrimaryTextColor
	tview.Styles.SecondaryTextColor = Theme.SecondaryTextColor
	tview.Styles.TertiaryTextColor = Theme.SecondaryTextColor
	tview.Styles.InverseTextColor = Theme.PrimaryTextColor // Text on highlighted backgrounds
}

// ---------------------------------------------------------------- //
// THEMED COMPONENT USAGE EXAMPLES
// ---------------------------------------------------------------- //

// NewThemedTextView returns a tview.TextView with basic theme styling.
func NewThemedTextView(text string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetText(text)
	tv.SetTextColor(Theme.PrimaryTextColor)
	tv.SetBackgroundColor(Theme.PanelBackgroundColor)
	tv.SetBorder(true)
	tv.SetBorderColor(Theme.BorderColor)
	tv.SetTitle(" Themed TextView ")
	tv.SetTitleColor(Theme.TitleColor)
	return tv
}

// NewThemedFlex returns a tview.Flex container with themed background and borders.
func NewThemedFlex() *tview.Flex {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(Theme.PanelBackgroundColor)
	flex.SetBorder(true)
	flex.SetBorderColor(Theme.BorderColor)
	flex.SetTitle(" Themed Flex Container ")
	flex.SetTitleColor(Theme.TitleColor)
	return flex
}

// NewThemedTable returns a styled tview.Table ready for data.
func NewThemedTable() *tview.Table {
	table := tview.NewTable()
	table.SetBorders(true)
	table.SetBordersColor(Theme.BorderColor)
	table.SetTitle(" Themed Table ")
	table.SetTitleColor(Theme.TitleColor)

	// Style for selected cells
	table.SetSelectable(true, true)
	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(Theme.PrimaryTextColor).
		Background(Theme.HighlightBackgroundColor))

	// Example Header
	headerStyle := tcell.StyleDefault.Foreground(Theme.TitleColor).Background(Theme.PanelBackgroundColor).Bold(true)
	table.SetCell(0, 0, tview.NewTableCell("Header 1").SetStyle(headerStyle).SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Header 2").SetStyle(headerStyle).SetAlign(tview.AlignCenter))

	// Example Data
	cellStyle := tcell.StyleDefault.Foreground(Theme.PrimaryTextColor).Background(Theme.PanelBackgroundColor)
	table.SetCell(1, 0, tview.NewTableCell("Data A1").SetStyle(cellStyle))
	table.SetCell(1, 1, tview.NewTableCell("Data A2").SetStyle(cellStyle))

	return table
}
