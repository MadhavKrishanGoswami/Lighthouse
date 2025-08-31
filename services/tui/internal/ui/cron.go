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

	// Main layout is a simple Flex container to help with centering.
	flex := tview.NewFlex().
		AddItem(intervalView, 0, 1, true)

	// Changed the title to reflect the new purpose.
	flex.SetBorder(true).SetTitle(" Update Interval ")

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
	// Enforce a maximum of 24 hours.
	if cw.duration > 24*time.Hour {
		cw.duration = 24 * time.Hour
	}
	// No need to reset a timer, just update the display text.
	cw.updateIntervalText()
}

// updateIntervalText refreshes the text in the right-hand box.
func (cw *CronWidget) updateIntervalText() {
	// Added more newlines for better vertical centering.
	text := fmt.Sprintf("\n\n\n[yellow]Interval:\n\n[white]%d Hour(s)\n\n\n\n[grey](Press +/- to change)", int(cw.duration.Hours()))
	cw.intervalView.SetText(text)
}
