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

// NewContainersPanel creates a new container panel.
func NewContainersPanel(app *App) *ContainersPanel {
	table := tview.NewTable().
		SetSelectable(true, false)

	cp := &ContainersPanel{
		Table: table,
		app:   app,
	}

	cp.SetTitle("[2] Containers ").SetBorder(true).SetTitleAlign(tview.AlignLeft).
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)

	// Input capture
	cp.SetInputCapture(cp.handleInput)

	// Handle selection font-color only
	cp.SetSelectionChangedFunc(func(row, column int) {
		for r := 1; r <= len(cp.containers); r++ {
			for c := 0; c < 4; c++ {
				cell := cp.GetCell(r, c)
				if r == row {
					cell.SetTextColor(tcell.ColorLightBlue) // selected font
				} else {
					cp.restoreCellColor(r, c)
				}
			}
		}
	})

	return cp
}

// Update the table with container data.
func (cp *ContainersPanel) Update(containers []Container) {
	cp.containers = containers
	cp.Clear()
	cp.SetFixed(1, 0) // fix header

	// Set headers
	headers := []string{"Name", "Image", "Status", "Watching"}
	for i, h := range headers {
		cp.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(Theme.TitleColor).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	for i, c := range containers {
		cp.drawRow(i+1, c)
	}
}

// Draws a single row
func (cp *ContainersPanel) drawRow(row int, c Container) {
	textColor := Theme.PrimaryTextColor
	if c.IsUpdating {
		textColor = Theme.AccentWarningColor
	}

	// Name
	cp.SetCell(row, 0, tview.NewTableCell(c.Name).SetTextColor(textColor).SetExpansion(1))

	// Image (truncate if too long)
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

// Restore the correct color for a given cell
func (cp *ContainersPanel) restoreCellColor(row, col int) {
	if row <= 0 || row-1 >= len(cp.containers) {
		return
	}
	c := cp.containers[row-1]
	switch col {
	case 0:
		color := Theme.PrimaryTextColor
		if c.IsUpdating {
			color = Theme.AccentWarningColor
		}
		cp.GetCell(row, col).SetTextColor(color)
	case 1:
		color := Theme.PrimaryTextColor
		if c.IsUpdating {
			color = Theme.AccentWarningColor
		}
		cp.GetCell(row, col).SetTextColor(color)
	case 2:
		color := Theme.AccentErrorColor
		if strings.Contains(strings.ToLower(c.Status), "running") {
			color = Theme.AccentGoodColor
		}
		if c.IsUpdating {
			color = Theme.AccentWarningColor
		}
		cp.GetCell(row, col).SetTextColor(color)
	case 3:
		color := Theme.PrimaryTextColor
		if c.IsWatching || c.IsUpdating {
			color = Theme.AccentWarningColor
		}
		cp.GetCell(row, col).SetTextColor(color)
	}
}

// Input handling for toggling watch and updating
func (cp *ContainersPanel) handleInput(event *tcell.EventKey) *tcell.EventKey {
	row, _ := cp.GetSelection()
	if row <= 0 || row-1 >= len(cp.containers) {
		return event
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
			return nil
		}
		c.IsUpdating = true
		cp.drawRow(row, *c)

		go func(container *Container, row int) {
			if cp.app != nil {
				cp.app.OnUpdateContainer(*container)
			}
			time.Sleep(2 * time.Second)
			cp.app.QueueUpdateDraw(func() {
				container.IsUpdating = false
				cp.drawRow(row, *container)
			})
		}(c, row)
		return nil
	}
	return event
}

// simulate container update
func (a *App) OnUpdateContainer(c Container) {
	time.Sleep(2 * time.Second)
	c.IsUpdating = false
}

func (a *App) OnWatchToggle(c Container) {
	c.IsWatching = !c.IsWatching
}
