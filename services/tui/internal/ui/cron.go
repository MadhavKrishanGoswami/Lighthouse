package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CronWidget now displays a static decorative clock and the update interval.
type CronWidget struct {
	*tview.Flex
	app          *App
	intervalView *tview.TextView
	duration     time.Duration
}

// NewCronWidget creates the redesigned cron widget.
func NewCronWidget(app *App) *CronWidget {
	// The interval configuration text is now the only element.
	intervalView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	intervalView.SetBackgroundColor(Theme.PanelBackgroundColor)

	// Main layout is a simple Flex container to help with centering.
	flex := tview.NewFlex().
		AddItem(intervalView, 0, 1, true)

	// Changed the title to reflect the new purpose.
	flex.SetBorder(true).SetTitle("[5] Cron Interval ").
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)

	widget := &CronWidget{
		Flex:         flex,
		app:          app,
		intervalView: intervalView,
		duration:     1 * time.Hour, // Default interval
	}

	widget.updateIntervalText()

	// Input capture for '+' and '-' keys to adjust the interval.
	widget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '+':
			widget.adjustInterval(1 * time.Hour)
		case '-':
			widget.adjustInterval(-1 * time.Hour)
		}
		return event
	})

	return widget
}

// adjustInterval handles changing the cron duration.
func (cw *CronWidget) adjustInterval(delta time.Duration) {
	cw.duration += delta
	// Enforce a minimum of 1 hour.
	if cw.duration < 1*time.Hour {
		cw.duration = 1 * time.Hour
	}

	// No need to reset a timer, just update the display text.
	cw.updateIntervalText()
}

// updateIntervalText refreshes the text in the right-hand box.

func (cw *CronWidget) updateIntervalText() {
	hours := int(cw.duration.Hours())

	text := fmt.Sprintf(
		"\n"+ // Subtle label
			"[green]%2d Hour(s)\n\n "+ // Highlighted main value
			"[yellow](Use + / - to adjust)\n", // Subtle instruction
		hours,
	)

	cw.intervalView.SetText(text)
}

func (cw *CronWidget) UpdateTime(hours int32) {
	// Convert seconds to duration
	cw.duration = time.Duration(hours) * time.Hour
	cw.updateIntervalText()
}
