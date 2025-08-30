package ui

import (
	"time"

	"github.com/rivo/tview"
)

// App struct holds the TUI application and its components.
type App struct {
	*tview.Application
	logo       *tview.Box
	hosts      *HostsPanel
	containers *tview.Box
	logs       *tview.Box
	cron       *tview.Box
	root       *tview.Flex
}

// NewApp creates and initializes the TUI application and its layout.
func NewApp() *App {
	// Initialize the main application.
	app := &App{
		Application: tview.NewApplication(),
	}

	// --- Initialize UI components (as placeholders) ---
	// In a real application, these would be initialized from their respective files
	// e.g., logo := NewLogoWidget()
	app.logo = createPlaceholderBox("Animated ASCII Logo")
	app.hosts = NewHostsPanel()
	app.containers = createPlaceholderBox("Containers")
	app.logs = createPlaceholderBox("Logs")
	app.cron = createPlaceholderBox("Cron Timer")
	// --- Link panels together ---
	app.hosts.SetHostSelectedFunc(func(hostName string) {
		app.containers.SetTitle("Containers on " + hostName)
	})

	// --- Load initial data ---
	// Fetch the mock host data and update the panel to display it.
	initialHosts := fetchMockHosts()
	app.hosts.Update(initialHosts)

	// --- Setup the main layout ---
	app.setupLayout()

	return app
}

// setupLayout defines the grid structure of the dashboard.
func (a *App) setupLayout() {
	// Left column containing Logo, Hosts, and Logs
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.logo, 0, 1, false).  // Proportion 1
		AddItem(a.hosts, 0, 2, false). // Proportion 2
		AddItem(a.cron, 0, 2, false)   // Proportion 2

	// Right column containing Containers and Cron timer
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.containers, 0, 15, false). // Proportion 5 (takes most space)
		AddItem(a.logs, 0, 5, false)         // Fixed size of 3 rows

	// Root flex container that holds the two columns
	a.root = tview.NewFlex().
		AddItem(leftColumn, 0, 8, false).  // Left column takes 1/3 of the width
		AddItem(rightColumn, 0, 12, false) // Right column takes 2/3 of the width

	// Set the root of the application
	a.SetRoot(a.root, true).EnableMouse(true)
}

// Run starts the TUI application.
func (a *App) Run() error {
	return a.Application.Run()
}

// createPlaceholderBox is a helper function to create a styled box for layout purposes.
func createPlaceholderBox(title string) *tview.Box {
	return tview.NewBox().
		SetBorder(true).
		SetTitleAlign(tview.AlignCenter).
		SetTitle(title)
}

// fetchMockHosts simulates fetching host data from an external source.
func fetchMockHosts() []Host {
	hosts := []Host{
		{Name: "host1", IP: "192:168.1.10", LastHeartbeat: time.Now(), MACAddress: "00:1A:2B:3C:4D:5E"},
		{Name: "host2", IP: "192:168.1.11", LastHeartbeat: time.Now().Add(-5 * time.Minute), MACAddress: "00:1A:2B:3C:4D:5F"},
		{Name: "host3", IP: "192:168.1.12", LastHeartbeat: time.Now().Add(-10 * time.Minute), MACAddress: "00:1A:2B:3C:4D:60"},
	}
	return hosts
}
