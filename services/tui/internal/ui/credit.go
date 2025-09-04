package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreditWidget displays a small corner credit in a box with GitHub and Twitter info.
type CreditWidget struct {
	*tview.TextView
	username   string
	twitter    string
	githubURL  string
	twitterURL string
}

// NewCreditWidget creates a polished corner box credit widget.
func NewCreditWidget(username, twitter string) *CreditWidget {
	widget := &CreditWidget{
		TextView: tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignCenter). // center text
			SetWrap(false).
			SetRegions(true), // for interactivity
		username:   username,
		twitter:    twitter,
		githubURL:  fmt.Sprintf("https://github.com/%s", username),
		twitterURL: fmt.Sprintf("https://twitter.com/%s", twitter),
	}

	// Corner box content with GitHub cat + username + Twitter
	text := fmt.Sprintf("\n[white]🐱 %s\n[white]🐦 @%s", username, twitter)
	widget.SetText(text)

	// Border and title
	widget.SetBorder(true)
	widget.SetTitle("[6] Made with 💖 by")
	widget.SetTitleAlign(tview.AlignCenter)
	widget.SetBorderPadding(0, 0, 1, 1)
	widget.SetBorderColor(Theme.BorderColor)
	widget.SetTitleColor(Theme.TitleColor)
	widget.SetBackgroundColor(Theme.PanelBackgroundColor)

	// Interactive: open GitHub or Twitter on Enter / Ctrl+Enter
	widget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEnter:
			widget.openBrowser(widget.githubURL)
			return nil
		case event.Key() == tcell.KeyCtrlE: // example: Ctrl+Enter to open Twitter
			widget.openBrowser(widget.twitterURL)
			return nil
		}
		return event
	})

	return widget
}

// openBrowser opens the given URL in the default browser.
func (c *CreditWidget) openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start()
}
