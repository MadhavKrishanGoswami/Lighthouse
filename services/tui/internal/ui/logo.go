package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogoWidget is a tview primitive that displays the animated ASCII logo.
type LogoWidget struct {
	*tview.Box
	animationFrames  []string
	currentFrame     int
	animationRunning bool
	app              *App
	animationSpeed   time.Duration
}

// NewLogoWidget creates a new widget for the animated logo.
func NewLogoWidget(app *App) *LogoWidget {
	box := tview.NewBox().SetBorder(false)
	logo := &LogoWidget{
		Box:              box,
		animationFrames:  generateLogoFrames(),
		currentFrame:     0,
		animationRunning: false,
		app:              app,
		animationSpeed:   80 * time.Millisecond, // Default speed
	}
	return logo
}

// Draw renders the widget.
func (l *LogoWidget) Draw(screen tcell.Screen) {
	l.Box.Draw(screen) // Draw the container box.

	x, y, width, height := l.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	lines := strings.Split(l.animationFrames[l.currentFrame], "\n")
	startY := y + (height-len(lines))/2 // Center vertically

	for i, line := range lines {
		// Center each line horizontally
		startX := x + (width-getVisibleLength(line))/2
		tview.Print(screen, line, startX, startY+i, width, tview.AlignLeft, tview.Styles.PrimaryTextColor)
	}
}

// StartAnimation begins the animation loop.
func (l *LogoWidget) StartAnimation() {
	if l.animationRunning {
		return
	}
	l.animationRunning = true
	go func() {
		for l.animationRunning {
			l.currentFrame++
			// After the intro animation, loop through the flicker frames.
			// The first 12 frames are the intro scan-line.
			if l.currentFrame >= len(l.animationFrames) {
				// Jump back to the start of the flicker loop (frame 12)
				l.currentFrame = 12
			}

			// Adjust speed for different parts of the animation
			if l.currentFrame < 12 {
				l.animationSpeed = 80 * time.Millisecond // Faster for intro
			} else {
				l.animationSpeed = 150 * time.Millisecond // Slower for flicker
			}

			l.app.QueueUpdateDraw(func() {})
			time.Sleep(l.animationSpeed)
		}
	}()
}

// StopAnimation halts the animation loop.
func (l *LogoWidget) StopAnimation() {
	l.animationRunning = false
}

// getVisibleLength calculates the length of a string, ignoring ANSI escape codes.
func getVisibleLength(s string) int {
	inEscape := false
	length := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			inEscape = false
		} else if !inEscape {
			length++
		}
	}
	return length
}

// --- ASCII Art & Animation Frames ---
const (
	brightCyan = "\033[1;96m"
	dimCyan    = "\033[2;36m"
	white      = "\033[1;97m"
	reset      = "\033[0m"
)

// The base ASCII art provided by the user.
// BUG FIX: Trimmed trailing spaces from each line to ensure proper centering.
var baseArt = []string{
	`██▓   ██▓  ▄████  ██░ ██ ▄▄▄█████▓ ██░ ██  ▒█████   █    ██    ██████ ▓█████`,
	`▓██▒   ▓██▒ ██▒ ▀█▒▓██░ ██▒▓  ██▒ ▓▒▓██░ ██▒▒██▒  ██▒ ██  ▓██▒▒██    ▒ ▓█   ▀`,
	`▒██░   ▒██▒▒██░▄▄▄░▒██▀▀██░▒ ▓██░ ▒░▒██▀▀██░▒██░  ██▒▓██  ▒██░░ ▓██▄   ▒███`,
	`▒██░   ░██░░▓█  ██▓░▓█ ░██ ░ ▓██▓ ░ ░▓█ ░██ ▒██   ██░▓▓█  ░██░  ▒   ██▒▒▓█  ▄`,
	`░██████▒░██░░▒▓███▀▒░▓█▒░██▓  ▒██▒ ░ ░▓█▒░██▓░ ████▓▒░▒▒█████▓ ▒██████▒▒░▒████▒`,
	`░ ▒░▓  ░░▓   ░▒   ▒  ▒ ░░▒░▒  ▒ ░░    ▒ ░░▒░▒░ ▒░▒░▒░ ░▒▓▒ ▒ ▒ ▒ ▒▓▒ ▒ ░░░ ▒░ ░`,
	`░ ░ ▒  ░ ▒ ░  ░   ░  ▒ ░▒░ ░    ░     ▒ ░▒░ ░  ░ ▒ ▒░ ░░▒░ ░ ░ ░ ░▒  ░ ░ ░ ░  ░`,
	`  ░ ░    ▒ ░░ ░   ░  ░  ░░ ░    ░      ░  ░░ ░░ ░ ░ ▒   ░░░ ░ ░  ░  ░   ░  ░`,
	`    ░  ░ ░      ░  ░  ░  ░      ░      ░  ░  ░    ░ ░     ░         ░     ░  ░`,
}

func generateLogoFrames() []string {
	var frames []string
	numLines := len(baseArt)
	// rand.Seed is deprecated and no longer needed in modern Go.

	// Phase 1: Scan-line reveal animation
	for i := 0; i <= numLines+1; i++ {
		var frameBuilder strings.Builder
		for j, line := range baseArt {
			if j < i-1 { // Lines above the scan-line
				frameBuilder.WriteString(dimCyan + line + reset + "\n")
			} else if j == i-1 { // The scan-line itself
				frameBuilder.WriteString(white + line + reset + "\n")
			} else {
				frameBuilder.WriteString("\n") // Empty line
			}
		}
		frames = append(frames, frameBuilder.String())
	}

	// Phase 2: Generate flicker/glitch frames for the loop
	glitchChars := []rune{'░', '▒', '▓', '█'}
	for i := 0; i < 20; i++ { // Generate 20 random flicker frames
		var frameBuilder strings.Builder
		for _, line := range baseArt {
			runes := []rune(line)
			for k := 0; k < len(runes); k++ {
				// With a small probability, glitch a character
				if runes[k] != ' ' && rand.Intn(100) < 3 {
					// Change character
					if rand.Intn(100) < 50 {
						runes[k] = glitchChars[rand.Intn(len(glitchChars))]
					}
					// Change color
					if rand.Intn(100) < 50 {
						frameBuilder.WriteString(white + string(runes[k]) + reset)
					} else {
						frameBuilder.WriteString(brightCyan + string(runes[k]) + reset)
					}
				} else {
					frameBuilder.WriteString(dimCyan + string(runes[k]) + reset)
				}
			}
			frameBuilder.WriteString("\n")
		}
		frames = append(frames, frameBuilder.String())
	}

	return frames
}
