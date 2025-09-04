package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogoWidget is a custom tview primitive that displays and animates an ASCII logo.
type LogoWidget struct {
	*tview.Box
	app              *App
	animationRunning bool
	animationSpeed   time.Duration
	artLines         []string
	width, height    int
}

// NewLogoWidget creates a new LogoWidget from ASCII art.
func NewLogoWidget(app *App) *LogoWidget {
	// Clean and split art into lines
	cleanedArt := strings.ReplaceAll(logoArt, " ", " ")
	lines := strings.Split(strings.TrimRight(cleanedArt, "\n"), "\n")

	// Find width & height
	h := len(lines)
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}

	return &LogoWidget{
		Box:              tview.NewBox(),
		app:              app,
		animationRunning: false,
		animationSpeed:   100 * time.Millisecond,
		artLines:         lines,
		width:            w,
		height:           h,
	}
}

// Draw renders the ASCII art inside the widget's box.

func (l *LogoWidget) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, w, h := l.GetInnerRect() // full inner rectangle

	// Calculate top-left corner to center the logo
	startX := x + (w-l.width)/2
	startY := y + (h-l.height)/2

	staticStep := time.Now().UnixNano() / int64(5e7) // ~20 FPS
	phase := int(staticStep % int64(l.width))

	for row, line := range l.artLines {
		for col, ch := range line {
			if ch == ' ' {
				continue
			}

			color := dimColor
			if abs(col-phase) < 2 {
				color = brightColor
			}

			tview.Print(screen, string(ch), startX+col, startY+row, 1, tview.AlignLeft, color)
		}
	}
}

// StartAnimation starts the shimmer animation.
func (l *LogoWidget) StartAnimation() {
	if l.animationRunning {
		return
	}
	l.animationRunning = true

	go func() {
		for l.animationRunning {
			l.app.QueueUpdateDraw(func() {})
			time.Sleep(l.animationSpeed)
		}
	}()
}

// StopAnimation halts the animation loop.
func (l *LogoWidget) StopAnimation() {
	l.animationRunning = false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// --- ASCII Art & Colors ---
const (
	brightColor = tcell.ColorDarkGray
	dimColor    = tcell.ColorLightGray
)

// ASCII Logo (Slim Modern style) for "Lighthouse"
const logoArt = `   __ _       _     _   _                          
  / /(_) __ _| |__ | |_| |__   ___  _   _ ___  ___ 
 / / | |/ _` + "`" + ` | '_ \| __| '_ \ / _ \| | | / __|/ _ \
/ /__| | (_| | | | | |_| | | | (_) | |_| \__ \  __/
\____/_|\__, |_| |_|\__|_| |_|\___/ \__,_|___/\___|
        |___/                                       `
