package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tf2-server-tui/internal/console"
	"tf2-server-tui/internal/config"
	"tf2-server-tui/internal/lockfile"
	"tf2-server-tui/internal/ui"
)

func main() {
	if err := console.SetTitle("TF2 Server Browser"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to set console title: %v\n", err)
	}

	configPath := flag.String("config", "", "path to servers.json")
	flag.Parse()

	appLock, err := lockfile.Acquire(".tf2-server-tui.lock")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := appLock.Release(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove lock file: %v\n", err)
		}
	}()

	loadedConfig, resolvedPath, configErr := config.Load(*configPath)
	model := ui.NewModel(loadedConfig, resolvedPath, configErr)

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run TUI: %v\n", err)
		os.Exit(1)
	}
}
