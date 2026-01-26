package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/jss826/cctasks/internal/model"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	// Disable East Asian Width to fix box drawing character width
	runewidth.DefaultCondition.EastAsianWidth = false

	model.AppVersion = Version

	app := model.NewApp()

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
