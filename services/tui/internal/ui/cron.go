package ui

import (
	"context"
	"fmt"
	"time"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CronWidget displays and updates cron interval.
type CronWidget struct {
	*tview.Flex
	app          *App
	intervalView *tview.TextView
	duration     time.Duration
}

func NewCronWidget(app *App) *CronWidget {
	intervalView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	intervalView.SetBackgroundColor(Theme.PanelBackgroundColor)
	flex := tview.NewFlex().AddItem(intervalView, 0, 1, true)
	flex.SetBorder(true).SetTitle("[5] Cron Interval ").
		SetBorderColor(Theme.BorderColor).
		SetTitleColor(Theme.TitleColor).
		SetBackgroundColor(Theme.PanelBackgroundColor)
	w := &CronWidget{Flex: flex, app: app, intervalView: intervalView, duration: 1 * time.Hour}
	w.updateIntervalText()
	w.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '+':
			w.adjustInterval(1 * time.Hour)
		case '-':
			w.adjustInterval(-1 * time.Hour)
		}
		return event
	})
	return w
}

func (cw *CronWidget) adjustInterval(delta time.Duration) {
	cw.duration += delta
	if cw.duration < 1*time.Hour {
		cw.duration = 1 * time.Hour
	}
	cw.updateIntervalText()
	// Propagate to server
	go func(hours int32) {
		if cw.app != nil && cw.app.client != nil {
			_, err := cw.app.client.SetCronTime(context.Background(), &tui.SetCronTimeRequest{CronTime: hours})
			if err != nil {
				cw.app.logs.AddLog("[red]SetCronTime failed: " + err.Error())
			} else {
				cw.app.logs.AddLog("[green]Cron time set")
			}
		}
	}(int32(cw.duration.Hours()))
}

func (cw *CronWidget) updateIntervalText() {
	hours := int(cw.duration.Hours())
	text := fmt.Sprintf("\n[green]%2d Hour(s)\n\n [yellow](Use + / - to adjust)\n", hours)
	cw.intervalView.SetText(text)
}

func (cw *CronWidget) UpdateTime(hours int32) {
	cw.duration = time.Duration(hours) * time.Hour
	cw.updateIntervalText()
}
