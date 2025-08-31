package ui

import (
	"time"

	"github.com/rivo/tview"
)

// App struct holds the TUI application and its components.
type App struct {
	*tview.Application
	logo                *LogoWidget
	hosts               *HostsPanel
	containers          *ContainersPanel // Changed from *tview.Box
	logs                *tview.Box
	cron                *CronWidget
	servicesStatus      *tview.Box
	totalServicesStatus *tview.Box
	root                *tview.Flex
}

// NewApp creates and initializes the TUI application and its layout.
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
	}

	// --- Initialize UI components ---
	app.logo = NewLogoWidget(app)
	app.hosts = NewHostsPanel()
	app.containers = NewContainersPanel(app) // Initialize the real ContainersPanel
	app.logs = createPlaceholderBox("Logs")
	app.cron = NewCronWidget(app)
	app.servicesStatus = createPlaceholderBox("Services Status")
	app.totalServicesStatus = createPlaceholderBox("Credits")

	// --- Link panels together ---
	// This is the updated link: when a host is selected, this function is called.
	app.hosts.SetHostSelectedFunc(func(hostName string) {
		// Fetch the mock container data for the selected host.
		containersForHost := fetchMockContainers(hostName)
		// Update the containers panel with the new data.
		app.containers.Update(containersForHost)
	})

	// --- Setup the main layout ---
	app.setupLayout()

	// --- Load initial data and set initial focus ---
	// This function runs safely after the first screen draw to prevent deadlocks.

	initialHosts := fetchMockHosts()
	app.hosts.Update(initialHosts)
	app.SetFocus(app.hosts)
	// Start the cron widget countdown.

	return app
}

// setupLayout defines the grid structure of the dashboard.
func (a *App) setupLayout() {
	// Left column
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.logo, 0, 1, false).
		AddItem(a.hosts, 0, 3, true).
		AddItem(a.logs, 0, 2, false)

	// Bottom row for the right column, laid out horizontally
	bottomRow := tview.NewFlex().
		AddItem(a.cron, 0, 1, false).
		AddItem(a.servicesStatus, 0, 2, false).
		AddItem(a.totalServicesStatus, 0, 1, false)

	// Right column
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.containers, 0, 7, false).
		AddItem(bottomRow, 0, 3, false)

	// Root flex container
	a.root = tview.NewFlex().
		AddItem(leftColumn, 0, 9, true).
		AddItem(rightColumn, 0, 11, false)

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
		SetTitle(" " + title + " ")
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

// fetchMockContainers simulates fetching container data for a given host.
func fetchMockContainers(hostName string) []Container {
	containers := []Container{
		{Name: hostName + "-container1", Image: "nginx:latest", Status: "Running", IsWatching: true, IsUpdating: false},
		{Name: hostName + "-container2", Image: "redis:alpine", Status: "Exited", IsWatching: false, IsUpdating: true},
		{Name: hostName + "-container3", Image: "postgres:13", Status: "Running", IsWatching: true, IsUpdating: true},
	}
	return containers
}
