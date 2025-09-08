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
	containers     *ContainersPanel // Changed from *tview.Box
	logs           *LogsPanel
	cron           *CronWidget
	servicesStatus *ServicesPanel
	credits        *CreditWidget
	root           *tview.Flex
}

// NewApp creates and initializes the TUI application and its layout.
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
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
	app.containers = NewContainersPanel(app) // Initialize the real ContainersPanel
	app.logs = NewLogsPanel(app)
	app.cron = NewCronWidget(app)
	app.servicesStatus = NewServicesPanel(app)
	app.credits = NewCreditWidget("MadhavKrishanGoswami", "Goswamimadhav24") // Replace with your GitHub username

	// --- Link panels together ---
	// This is the updated link: when a host is selected, this function is called.
	app.hosts.SetHostSelectedFunc(func(hostName string) {
		// Fetch the mock container data for the selected host.
		containersForHost := fetchMockContainers(hostName)
		// Update the containers panel with the new data.
		app.containers.Update(containersForHost)
	})

	// --- Setup the main layout ---

	// Focus handling skipped (tview does not expose direct global focus hook)
	app.setupLayout()

	// --- Load initial data and set initial focus ---
	// This function runs safely after the first screen draw to prevent deadlocks.

	initialHosts := fetchMockHosts()
	app.hosts.Update(initialHosts)
	app.SetFocus(app.hosts)
	app.servicesStatus.Update(fetchMockServices())
	app.logs.SafeLogSimulator(fetchMockLogs(), 5)

	// --- Global key handler for numeric switching ---
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
	// Background handled in SetBeforeDrawFunc

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
