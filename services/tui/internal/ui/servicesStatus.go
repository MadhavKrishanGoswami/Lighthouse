package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"} // Braille spinner

// Service holds the status of various components.
type Service struct {
	OrchestratorStatus    bool
	RegistryMonitorStatus bool
	TotalHosts            int
	DatabaseStatus        bool
}

// ServicesPanel displays the status of services in a table.
type ServicesPanel struct {
	*tview.Flex
	app        *App
	table      *tview.Table // Changed from tview.TextView to tview.Table
	spinnerIdx int
	ticker     *time.Ticker
	quit       chan struct{}
	lastData   Service
}

// NewServicesPanel creates a new panel for displaying service statuses.
func NewServicesPanel(app *App) *ServicesPanel {
	// Use a table for a responsive layout that scales properly.
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)

	flex := tview.NewFlex().
		AddItem(table, 0, 1, false)

	flex.SetBorder(true).SetTitle(" Services Status ").
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)

	widget := &ServicesPanel{
		Flex:       flex,
		app:        app,
		table:      table, // Use the table instead of textView
		spinnerIdx: 0,
		quit:       make(chan struct{}),
	}

	// The animation ticker remains the same.
	widget.ticker = time.NewTicker(200 * time.Millisecond)
	go widget.animate()

	return widget
}

// animate cycles through spinner frames to indicate activity.
func (sp *ServicesPanel) animate() {
	for {
		select {
		case <-sp.ticker.C:
			sp.spinnerIdx = (sp.spinnerIdx + 1) % len(spinnerFrames)
			sp.app.QueueUpdateDraw(func() {
				sp.Update(sp.lastData)
			})
		case <-sp.quit:
			return
		}
	}
}

// Stop halts the panel's animation goroutine.
func (sp *ServicesPanel) Stop() {
	sp.ticker.Stop()
	close(sp.quit)
}

// Update refreshes the table with the latest service data.
func (sp *ServicesPanel) Update(services Service) {
	sp.lastData = services
	sp.table.Clear() // Clear the table before drawing new data

	spin := spinnerFrames[sp.spinnerIdx]

	// Helper function to create and style a row for a service.
	setServiceStatus := func(row int, name string, status bool) {
		// Column 1: Service Name
		sp.table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf(" %s", name)).
			SetTextColor(Theme.PrimaryTextColor).
			SetBackgroundColor(Theme.PanelBackgroundColor).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)) // Allow this column to expand

		// Column 2: Status ("CONNECTED" / "DISCONNECTED")
		statusCell := tview.NewTableCell("")
		if status {
			statusCell.SetText("CONNECTED").SetTextColor(Theme.AccentGoodColor)
		} else {
			statusCell.SetText("DISCONNECTED").SetTextColor(Theme.AccentErrorColor)
		}
		statusCell.SetAlign(tview.AlignLeft)
		sp.table.SetCell(row, 1, statusCell)

		// Column 3: Spinner
		sp.table.SetCell(row, 2, tview.NewTableCell(spin+" ").
			SetTextColor(Theme.SecondaryTextColor).
			SetAlign(tview.AlignRight))
	}

	// Populate the table with the current status of each service.
	setServiceStatus(0, "Orchestrator", services.OrchestratorStatus)
	setServiceStatus(1, "Registry Monitor", services.RegistryMonitorStatus)
	setServiceStatus(2, "Database", services.DatabaseStatus)

	// Add a separate, styled row for "Total Hosts".
	sp.table.SetCell(3, 0, tview.NewTableCell(" Total Hosts").
		SetTextColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor).
		SetAlign(tview.AlignLeft))

	sp.table.SetCell(3, 1, tview.NewTableCell(fmt.Sprintf("%d", services.TotalHosts)).
		SetTextColor(Theme.PrimaryTextColor).
		SetBackgroundColor(Theme.PanelBackgroundColor).
		SetAlign(tview.AlignLeft))

	sp.table.SetCell(3, 2, tview.NewTableCell(spin+" ").
		SetTextColor(Theme.SecondaryTextColor).
		SetAlign(tview.AlignRight))
}
