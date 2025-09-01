package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

// LogsPanel displays log messages in a scrollable view.
type LogsPanel struct {
	*tview.Flex
	app      *App
	textView *tview.TextView
}

// NewLogsPanel creates a new panel for displaying logs.
func NewLogsPanel(app *App) *LogsPanel {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw() // redraw when text changes
		})

	flex := tview.NewFlex().
		AddItem(textView, 0, 1, false)

	flex.SetBorder(true).SetTitle(" Logs ")

	return &LogsPanel{
		Flex:     flex,
		app:      app,
		textView: textView,
	}
}

// AddLog appends a new message safely from any goroutine.
func (lp *LogsPanel) AddLog(message string) {
	lp.app.QueueUpdateDraw(func() {
		fmt.Fprintln(lp.textView, message)
	})
}

// ClearLogs removes all messages safely.
func (lp *LogsPanel) ClearLogs() {
	lp.app.QueueUpdateDraw(func() {
		lp.textView.Clear()
	})
}

// SafeLogSimulator runs logs asynchronously to avoid blocking.
func (lp *LogsPanel) SafeLogSimulator(logs []string, delaySec int) {
	go func() {
		for _, log := range logs {
			lp.AddLog(log) // safe to call from goroutine
			time.Sleep(time.Duration(delaySec) * time.Second)
		}
	}()
}
