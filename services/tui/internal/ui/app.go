package ui

import (
	"context"
	"sync"
	"time"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
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

	// Realtime state
	client        tui.TUIServiceClient
	conn          *grpc.ClientConn
	cancel        context.CancelFunc
	dataMu        sync.RWMutex
	hostsData     []Host
	containersMap map[string][]Container // mac -> containers
	nameToMAC     map[string]string      // hostname -> mac
}

// NewApp creates and initializes the TUI application and its layout.
func NewApp(client tui.TUIServiceClient, conn *grpc.ClientConn) *App {
	app := &App{
		Application:   tview.NewApplication(),
		client:        client,
		conn:          conn,
		containersMap: make(map[string][]Container),
		nameToMAC:     make(map[string]string),
	}

	// Apply global theme before building components
	ApplyLighthouseTheme()
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
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
	app.cron = NewCronWidget(app)
	app.servicesStatus = NewServicesPanel(app)
	app.credits = NewCreditWidget("MadhavKrishanGoswami", "Goswamimadhav24")

	// Host selection -> update containers from realtime map
	app.hosts.SetHostSelectedFunc(func(hostName string) {
		app.dataMu.RLock()
		mac := app.nameToMAC[hostName]
		containers := app.containersMap[mac]
		app.dataMu.RUnlock()
		app.containers.Update(containers)
	})

	// Setup layout
	app.setupLayout()

	// Initial placeholder state
	app.hosts.Update([]Host{{Name: "(connecting...)", IP: "-", MACAddress: "-", LastHeartbeat: time.Now()}})
	app.servicesStatus.Update(Service{})
	app.SetFocus(app.hosts)

	// Start realtime streams after construction
	go app.startRealtime()

	// Global key handler
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
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.logo, 0, 1, false).
		AddItem(a.hosts, 0, 3, true).
		AddItem(a.logs, 0, 2, false)

	bottomRow := tview.NewFlex().
		AddItem(a.cron, 0, 3, false).
		AddItem(a.servicesStatus, 0, 4, false).
		AddItem(a.credits, 0, 3, false)

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.containers, 0, 85, false).
		AddItem(bottomRow, 0, 15, false)

	a.root = tview.NewFlex().
		AddItem(leftColumn, 0, 7, true).
		AddItem(rightColumn, 0, 13, false)

	a.SetRoot(a.root, true).EnableMouse(true)
}

// Run starts the TUI application.
func (a *App) Run() error { return a.Application.Run() }

// Close releases resources / cancels streams.
func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.conn != nil {
		_ = a.conn.Close()
	}
}

func createPlaceholderBox(title string) *tview.Box {
	return tview.NewBox().
		SetBorder(true).
		SetTitleAlign(tview.AlignCenter).
		SetTitle(" " + title + " ")
}
