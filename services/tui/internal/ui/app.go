package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App struct holds the TUI application and its components.
type App struct {
	*tview.Application
	logo           *LogoWidget
	hosts          *HostsPanel
	containers     *ContainersPanel
	logs           *LogsPanel
	cron           *CronWidget
	servicesStatus *ServicesPanel
	credits        *CreditWidget
	root           *tview.Flex

	// Real-time data management
	containersMap   map[string][]Container // MAC address -> containers
	selectedHostMAC string
	hostsData       []Host // Store current hosts data
}

// NewApp creates and initializes the TUI application and its layout.
func NewApp() *App {
	app := &App{
		Application:   tview.NewApplication(),
		containersMap: make(map[string][]Container),
	}

	// Apply global theme before building components
	ApplyLighthouseTheme()
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// Fill background with app background color
		style := tcell.StyleDefault.Background(Theme.AppBackgroundColor).Foreground(Theme.PrimaryTextColor)
		w, h := screen.Size()
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				screen.SetContent(x, y, ' ', nil, style)
			}
		}
		return false
	})

	// --- Initialize UI components ---
	app.logo = NewLogoWidget(app)
	app.hosts = NewHostsPanel()
	app.containers = NewContainersPanel(app)
	app.logs = NewLogsPanel(app)
	app.logs.AddLog("[TUI] Initialized – waiting for snapshots (press q to quit)")
	app.cron = NewCronWidget(app)
	app.servicesStatus = NewServicesPanel(app)
	app.credits = NewCreditWidget("MadhavKrishanGoswami", "Goswamimadhav24")

	// --- Link panels together ---
	// Updated to use MAC address instead of hostname for container lookup
	app.hosts.SetHostSelectedFunc(func(hostMAC string) {
		app.selectedHostMAC = hostMAC
		// Look up containers for the selected host by MAC address
		if containers, exists := app.containersMap[hostMAC]; exists {
			app.containers.Update(containers)
		} else {
			// Clear containers panel if no containers found for this host
			app.containers.Update([]Container{})
		}
	})

	// --- Setup the main layout ---
	app.setupLayout()

	// --- Set initial focus ---
	app.SetFocus(app.hosts)

	// --- Global key handler for numeric switching ---
	// Global input capture including quit keys
	// Global input capture including quit keys
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Quit keys
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' || event.Rune() == 'Q' {
			app.Stop()
			return nil
		}
		switch event.Rune() {
		case '1':
			app.SetFocus(app.hosts)
		case '2':
			app.SetFocus(app.containers)
		case '3':
			app.SetFocus(app.logs)
		case '4':
			app.SetFocus(app.servicesStatus)
		case '5':
			app.SetFocus(app.cron)
		case '6':
			app.SetFocus(app.credits)
		}
		return event
	})
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
		AddItem(a.cron, 0, 3, false).
		AddItem(a.servicesStatus, 0, 4, false).
		AddItem(a.credits, 0, 3, false)

	// Right column
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.containers, 0, 85, false).
		AddItem(bottomRow, 0, 15, false)

	// Root flex container
	a.root = tview.NewFlex().
		AddItem(leftColumn, 0, 7, true).
		AddItem(rightColumn, 0, 13, false)

	a.SetRoot(a.root, true).EnableMouse(true)
}

// Run starts the TUI application.
func (a *App) Run() error {
	return a.Application.Run()
}

// === Real-time Data Update Methods ===

// UpdateHosts updates the hosts panel with new data from gRPC stream
func (app *App) UpdateHosts(hosts []Host) {
	app.hostsData = hosts
	app.hosts.Update(hosts)

	// If no host is currently selected and we have hosts, select the first one
	if app.selectedHostMAC == "" && len(hosts) > 0 {
		app.selectedHostMAC = hosts[0].MACAddress
		if containers, exists := app.containersMap[app.selectedHostMAC]; exists {
			app.containers.Update(containers)
		}
	}
}

// UpdateContainersMap stores container data for all hosts from gRPC stream
func (app *App) UpdateContainersMap(containersMap map[string][]Container) {
	app.containersMap = containersMap

	// If a host is currently selected, update the containers panel
	if app.selectedHostMAC != "" {
		if containers, exists := containersMap[app.selectedHostMAC]; exists {
			app.containers.Update(containers)
		} else {
			// Clear containers if selected host has no containers
			app.containers.Update([]Container{})
		}
	}
}

// UpdateServices updates the services status panel with new data
func (app *App) UpdateServices(service Service) {
	app.servicesStatus.Update(service)
}

// UpdateLogs adds new log entry to the logs panel
func (app *App) UpdateLogs(logEntry string) {
	app.logs.AddLog(logEntry)
}

// UpdateCronTime updates the cron widget with new cron time
func (app *App) UpdateCronTime(cronTime int32) {
	app.cron.UpdateTime(cronTime)
}

// GetSelectedHostMAC returns the currently selected host's MAC address
func (app *App) GetSelectedHostMAC() string {
	return app.selectedHostMAC
}

// SetWatchStatus updates the watch status for a specific container
func (app *App) SetWatchStatus(containerName string, watch bool) {
	if containers, exists := app.containersMap[app.selectedHostMAC]; exists {
		for i, container := range containers {
			if container.Name == containerName {
				containers[i].IsWatching = watch
				break
			}
		}
		app.containers.Update(containers)
	}
}

// createPlaceholderBox is a helper function to create a styled box for layout purposes.
func createPlaceholderBox(title string) *tview.Box {
	return tview.NewBox().
		SetBorder(true).
		SetTitleAlign(tview.AlignCenter).
		SetTitle(" " + title + " ")
}
