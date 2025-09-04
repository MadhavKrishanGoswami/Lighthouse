package ui

import (
	"fmt"
	"time"

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
}

// NewHostsPanel creates a new panel for displaying hosts.
func NewHostsPanel() *HostsPanel {
	table := tview.NewTable().
		SetSelectable(true, false)

	hp := &HostsPanel{
		Table: table,
	}

	// Re-add the border and title to the panel itself.
	hp.SetTitle(" Hosts ").SetBorder(true).SetTitleAlign(tview.AlignLeft).
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)

	// Set the headers for the table.
	headers := []string{"Name", "IP", "MAC Address", "Last Heartbeat"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(Theme.TitleColor).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetBackgroundColor(Theme.PanelBackgroundColor).
			SetExpansion(1) // Set expansion factor to 1 for equal distribution.
		hp.SetCell(0, i, cell)
	}

	hp.SetSelectionChangedFunc(hp.handleSelectionChange)

	return hp
}

// Update an entire set of hosts in the table.
func (hp *HostsPanel) Update(hosts []Host) {
	hp.hosts = hosts
	hp.Clear()

	// Re-add headers.
	headers := []string{"Name", "IP", "MAC Address", "Last Heartbeat"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(Theme.TitleColor).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetBackgroundColor(Theme.PanelBackgroundColor).
			SetExpansion(1) // Ensure headers also expand.
		hp.SetCell(0, i, cell)
	}

	// Add data rows.
	for i, host := range hosts {
		row := i + 1
		// Set Expansion(1) for all data cells to make them share space equally.
		hp.SetCell(row, 0, tview.NewTableCell(host.Name).SetTextColor(Theme.PrimaryTextColor).SetBackgroundColor(Theme.PanelBackgroundColor).SetExpansion(3))
		hp.SetCell(row, 1, tview.NewTableCell(host.IP).SetTextColor(Theme.PrimaryTextColor).SetBackgroundColor(Theme.PanelBackgroundColor).SetExpansion(2))
		hp.SetCell(row, 2, tview.NewTableCell(host.MACAddress).SetTextColor(Theme.PrimaryTextColor).SetBackgroundColor(Theme.PanelBackgroundColor).SetExpansion(2))
		hp.SetCell(row, 3, tview.NewTableCell(formatHeartbeat(host.LastHeartbeat)).SetTextColor(Theme.AccentGoodColor).SetBackgroundColor(Theme.PanelBackgroundColor).SetExpansion(1))
	}

	// Select the first host by default if the list is not empty.
	if len(hp.hosts) > 0 {
		hp.Select(1, 0)
	}
}

// SetHostSelectedFunc sets the callback function that is triggered when a new host is selected.
func (hp *HostsPanel) SetHostSelectedFunc(handler func(hostName string)) {
	hp.onHostSelected = handler
}

// handleSelectionChange is called by tview whenever the user navigates the table.
func (hp *HostsPanel) handleSelectionChange(row, column int) {
	if row > 0 && row-1 < len(hp.hosts) && hp.onHostSelected != nil {
		selectedHost := hp.hosts[row-1]
		if selectedHost.Name != hp.selectedHostName {
			hp.selectedHostName = selectedHost.Name
			hp.onHostSelected(selectedHost.Name)
		}
	}
}

// formatHeartbeat calculates the time since the last heartbeat and formats it.
func formatHeartbeat(t time.Time) string {
	duration := time.Since(t)
	if duration.Seconds() < 60 {
		return fmt.Sprintf("%.0fs ago", duration.Seconds())
	} else if duration.Minutes() < 60 {
		return fmt.Sprintf("%.0fm ago", duration.Minutes())
	}
	return fmt.Sprintf("%.0fh ago", duration.Hours())
}
