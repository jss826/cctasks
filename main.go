package main

import (
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/jss826/cctasks/internal/model"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	// Disable East Asian Width to fix box drawing character width
	runewidth.DefaultCondition.EastAsianWidth = false

	// Use build info version if not set via ldflags (e.g., go install)
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
			Version = info.Main.Version
		}
	}

	// Handle --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("cctasks %s\n", Version)
		return
	}

	model.AppVersion = Version

	app := model.NewApp()

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
