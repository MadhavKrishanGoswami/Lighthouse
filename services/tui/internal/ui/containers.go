package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Container struct {
	Name       string
	Image      string
	Status     string
	IsWatching bool
	IsUpdating bool
}

type ContainersPanel struct {
	*tview.Table
	containers []Container
	app        *App
}

func NewContainersPanel(app *App) *ContainersPanel {
	table := tview.NewTable().
		SetSelectable(true, false)

	cp := &ContainersPanel{
		Table: table,
		app:   app,
	}

	cp.SetTitle(" Containers ").SetBorder(true).SetTitleAlign(tview.AlignLeft).
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)
	cp.SetInputCapture(cp.handleInput)

	return cp
}

func (cp *ContainersPanel) Update(containers []Container) {
	cp.containers = containers
	cp.Clear() // clear all previous cells

	cp.SetSelectable(true, false) // ensure table is selectable

	// Set headers
	headers := []string{"Name", "Image", "Status", "Watching"}
	for i, h := range headers {
		cp.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(Theme.TitleColor).
				SetBackgroundColor(Theme.PanelBackgroundColor).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	// Draw each container in its own row
	for i, c := range containers {
		row := i + 1
		cp.drawRow(row, c)
	}

	// Make sure table height is enough to show all rows
	cp.SetFixed(1, 0) // fix header row
}

func (cp *ContainersPanel) drawRow(row int, c Container) {
	textColor := Theme.PrimaryTextColor
	if c.IsUpdating {
		textColor = Theme.AccentWarningColor
	}

	// Name
	cp.SetCell(row,
		0, tview.NewTableCell(c.Name).SetTextColor(textColor).SetExpansion(1))

	// Image
	imageName := c.Image
	if len(imageName) > 30 {
		imageName = imageName[:27] + "..."
	}
	cp.SetCell(row, 1, tview.NewTableCell(imageName).SetTextColor(textColor).SetExpansion(1))

	// Status
	statusColor := Theme.AccentErrorColor
	if strings.Contains(strings.ToLower(c.Status), "running") {
		statusColor = Theme.AccentGoodColor
	}
	if c.IsUpdating {
		statusColor = Theme.AccentWarningColor
	}
	cp.SetCell(row, 2, tview.NewTableCell(c.Status).SetTextColor(statusColor).SetAlign(tview.AlignCenter).SetExpansion(1))

	// Watching
	watchText := "No"
	watchColor := textColor
	if c.IsWatching {
		watchText = "Yes"
		watchColor = Theme.AccentWarningColor
		if c.IsUpdating {
			watchColor = Theme.AccentWarningColor
		}
	}
	cp.SetCell(row, 3, tview.NewTableCell(watchText).SetTextColor(watchColor).SetAlign(tview.AlignCenter))
}

func (cp *ContainersPanel) handleInput(event *tcell.EventKey) *tcell.EventKey {
	row, _ := cp.GetSelection()
	if row <= 0 || row-1 >= len(cp.containers) {
		return event // nothing to do
	}

	containerIndex := row - 1
	c := &cp.containers[containerIndex]

	switch event.Rune() {
	case 'w', 'W':
		c.IsWatching = !c.IsWatching
		cp.drawRow(row, *c)
		if cp.app != nil {
			cp.app.OnWatchToggle(*c)
		}
		return nil

	case 'u', 'U':
		if c.IsUpdating {
			return nil // already updating
		}

		c.IsUpdating = true
		cp.drawRow(row, *c)

		go func(container *Container, row int) {
			if cp.app != nil {
				cp.app.OnUpdateContainer(*container)
			}
			time.Sleep(2 * time.Second) // simulate work

			cp.app.QueueUpdateDraw(func() {
				container.IsUpdating = false
				cp.drawRow(row, *container)
			})
		}(c, row)

		return nil
	}

	return event
}

// simple function to simulate container update
func (a *App) OnUpdateContainer(c Container) {
	// set update to true
	c.IsUpdating = true
	// simulate update process
	time.Sleep(2 * time.Second) // simulate work
	// set update to false
	c.IsUpdating = false
}

func (a *App) OnWatchToggle(c Container) {
	// handle watch toggle logic here
	c.IsWatching = !c.IsWatching
}
