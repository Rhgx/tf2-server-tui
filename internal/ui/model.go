package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"tf2-server-tui/internal/config"
)

// NewModel builds the Bubble Tea model from the resolved config state and
// primes the terminal title before the first render happens.
func NewModel(cfg config.Config, configPath string, configErr error) tea.Model {
	initialStatus := ""
	initialError := false
	if configErr != nil {
		initialStatus = configErr.Error()
		initialError = true
	}

	m := &model{
		config:      cfg,
		configPath:  configPath,
		configErr:   configErr,
		statusText:  initialStatus,
		statusError: initialError,
	}
	m.updateTitle()
	return m
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.refreshCmd(), m.autoRefreshCmd()}
	return tea.Batch(cmds...)
}
