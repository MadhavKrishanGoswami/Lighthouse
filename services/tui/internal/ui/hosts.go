package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Host represents the data for a single host machine.
type Host struct {
	Name          string
	IP            string
	MACAddress    string
	LastHeartbeat time.Time
}

// HostsPanel represents the TUI component that displays a list of hosts.
type HostsPanel struct {
	*tview.Table
	hosts            []Host
	onHostSelected   func(hostName string)
	selectedHostName string
	showMAC          bool // toggle for MAC visibility
}

// NewHostsPanel creates a new panel for displaying hosts.
func NewHostsPanel() *HostsPanel {
	table := tview.NewTable().
		SetSelectable(true, false)

	hp := &HostsPanel{
		Table:   table,
		showMAC: false,
	}

	// Border and title
	hp.SetTitle("[1] Hosts ").SetBorder(true).SetTitleAlign(tview.AlignLeft).
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)

	// Headers
	hp.setHeaders()

	// Handle selection highlighting (font color only)
	hp.SetSelectionChangedFunc(func(row, column int) {
		for r := 1; r <= len(hp.hosts); r++ { // skip header
			for c := 0; c < 4; c++ {
				cell := hp.GetCell(r, c)
				if r == row {
					cell.SetTextColor(tcell.ColorLightBlue) // selected font
				} else {
					cell.SetTextColor(hp.getDefaultColor(c))
				}
			}
		}
		hp.handleSelectionChange(row, column)
	})

	// Toggle MAC visibility with 'm'
	hp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'm', 'M':
			hp.showMAC = !hp.showMAC
			hp.Update(hp.hosts)
			return nil
		}
		return event
	})

	return hp
}

// Update an entire set of hosts in the table.
func (hp *HostsPanel) Update(hosts []Host) {
	hp.hosts = hosts
	hp.Clear()

	hp.setHeaders()

	expansions := []int{3, 2, 2, 3}
	for i, host := range hosts {
		row := i + 1

		hp.SetCell(row, 0, tview.NewTableCell(host.Name).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(expansions[0]))

		hp.SetCell(row, 1, tview.NewTableCell(host.IP).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignCenter).
			SetExpansion(expansions[1]))

		macText := host.MACAddress
		if !hp.showMAC {
			if len(macText) > 5 {
				macText = macText[:5] + "..."
			} else {
				// leave as-is if shorter than 5 chars (avoid slice panic)
			}
		}
		hp.SetCell(row, 2, tview.NewTableCell(macText).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(expansions[2]))

		hp.SetCell(row, 3, tview.NewTableCell(formatHeartbeat(host.LastHeartbeat)).
			SetTextColor(tcell.ColorOrange).
			SetAlign(tview.AlignRight).
			SetExpansion(expansions[3]))
	}

	if len(hp.hosts) > 0 {
		hp.Select(1, 0)
	}
}

// SetHostSelectedFunc sets the callback function when a host is selected.
func (hp *HostsPanel) SetHostSelectedFunc(handler func(hostName string)) {
	hp.onHostSelected = handler
}

// handleSelectionChange triggers when the user navigates the table.
func (hp *HostsPanel) handleSelectionChange(row, column int) {
	if row > 0 && row-1 < len(hp.hosts) && hp.onHostSelected != nil {
		selectedHost := hp.hosts[row-1]
		if selectedHost.Name != hp.selectedHostName {
			hp.selectedHostName = selectedHost.Name
			hp.onHostSelected(selectedHost.Name)
		}
	}
}

// Helper: returns default color for a column
func (hp *HostsPanel) getDefaultColor(column int) tcell.Color {
	switch column {
	case 0:
		return tcell.ColorWhite
	case 1:
		return tcell.ColorGreen
	case 2:
		return tcell.ColorYellow
	case 3:
		return tcell.ColorOrange
	default:
		return tcell.ColorWhite
	}
}

// Helper: sets table headers
func (hp *HostsPanel) setHeaders() {
	headers := []string{"Name", "IP", "MAC Address", "Heartbeat"}
	expansions := []int{3, 2, 2, 3}
	for i, header := range headers {
		hp.SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tcell.ColorLightGray).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetExpansion(expansions[i]))
	}
}

// formatHeartbeat calculates time since last heartbeat.
func formatHeartbeat(t time.Time) string {
	duration := time.Since(t)
	if duration.Seconds() < 60 {
		return fmt.Sprintf("%.0fs ago", duration.Seconds())
	} else if duration.Minutes() < 60 {
		return fmt.Sprintf("%.0fm ago", duration.Minutes())
	}
	return fmt.Sprintf("%.0fh ago", duration.Hours())
}
