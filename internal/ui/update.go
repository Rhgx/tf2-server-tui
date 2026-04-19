package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"tf2-server-tui/internal/config"
	"tf2-server-tui/internal/query"
	"tf2-server-tui/internal/steam"
)

func (m *model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.deleteConfirm {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter", "y", "Y":
				return m.confirmDeleteSavedServer()
			case "esc", "n", "N", "q", "d", "delete", "backspace":
				m.deleteConfirm = false
				return m, nil
			}
			return m, nil
		}

		if m.quitConfirm {
			switch msg.String() {
			case "ctrl+c", "enter", "y", "Y":
				return m, tea.Quit
			case "esc", "n", "N", "q":
				m.quitConfirm = false
				return m, nil
			}
			return m, nil
		}

		if msg.String() == "q" {
			m.quitConfirm = true
			return m, nil
		}

		switch m.mode {
		case modeImportFavorites:
			return m.handleImportKey(msg)
		case modeSettingsMenu:
			return m.handleSettingsKey(msg)
		case modeSavedServers:
			return m.handleSavedServersKey(msg)
		case modeEditServer, modeAddServer:
			return m.handleServerFormKey(msg)
		case modeEditRefreshInterval:
			return m.handleRefreshIntervalKey(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if len(m.servers) > 0 {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				} else {
					m.selectedIndex = len(m.servers) - 1
				}
				m.updateTitle()
			}
			return m, nil
		case "down", "j":
			if len(m.servers) > 0 {
				if m.selectedIndex < len(m.servers)-1 {
					m.selectedIndex++
				} else {
					m.selectedIndex = 0
				}
				m.updateTitle()
			}
			return m, nil
		case "r":
			return m, m.refreshCmd()
		case "e":
			m.mode = modeSettingsMenu
			m.settingsIndex = 0
			m.updateTitle()
			return m, nil
		case "enter":
			if len(m.servers) == 0 {
				if len(m.config.Servers) > 0 {
					return m, m.setStatus("servers are still loading", true)
				}
				return m, m.setStatus("no configured servers yet", true)
			}
			server := m.servers[m.selectedIndex]
			if !server.Online {
				return m, m.setStatus("cannot connect to an offline server", true)
			}
			return m, m.launchServerCmd(server)
		}
		return m, nil

	case refreshMsg:
		oldInterval := m.config.RefreshInterval
		m.refreshing = false
		if m.mode != modeBrowse {
			if msg.Err == nil {
				m.servers = msg.Servers
				m.lastRefresh = msg.When
				if m.selectedIndex >= len(m.servers) && len(m.servers) > 0 {
					m.selectedIndex = len(m.servers) - 1
				}
			}
			m.updateTitle()
			return m, nil
		}

		m.config = msg.Config
		m.configErr = msg.Err
		if msg.Err != nil {
			m.updateTitle()
			return m, m.setStatus(msg.Err.Error(), true)
		}
		m.servers = msg.Servers
		m.lastRefresh = msg.When
		if m.selectedIndex >= len(m.servers) && len(m.servers) > 0 {
			m.selectedIndex = len(m.servers) - 1
		}
		if m.savedIndex >= len(m.config.Servers) && len(m.config.Servers) > 0 {
			m.savedIndex = len(m.config.Servers) - 1
		} else if len(m.config.Servers) == 0 {
			m.savedIndex = 0
		}
		m.updateTitle()

		cmds := make([]tea.Cmd, 0, 2)
		if oldInterval != m.config.RefreshInterval {
			cmds = append(cmds, m.restartAutoRefreshCmd())
		}

		if len(m.servers) == 0 {
			cmds = append(cmds, m.setStatus(fmt.Sprintf("edit %s to add servers", m.configPath), true))
			return m, tea.Batch(cmds...)
		}

		cmds = append(cmds, m.setStatus(
			fmt.Sprintf("refreshed %d server(s) at %s", len(m.servers), msg.When.Format("15:04:05")),
			false,
		))
		return m, tea.Batch(cmds...)

	case refreshErrMsg:
		m.refreshing = false
		m.updateTitle()
		return m, m.setStatus(msg.Err.Error(), true)

	case launchResultMsg:
		if msg.Err != nil {
			return m, m.setStatus(msg.Err.Error(), true)
		}
		return m, m.setStatus(msg.Text, false)

	case favoritesLoadedMsg:
		if msg.Version != m.favoritesLoadID || m.mode != modeImportFavorites {
			return m, nil
		}
		m.importLoading = false
		m.importItems = msg.Items
		if m.importIndex >= len(m.importItems) {
			m.importIndex = maxInt(0, len(m.importItems)-1)
		}
		m.updateTitle()
		if msg.Err != nil {
			return m, m.setStatus(msg.Err.Error(), true)
		}
		if len(m.importItems) == 0 {
			return m, m.setStatus("no Steam favorites available to import", true)
		}
		return m, m.setStatus(
			fmt.Sprintf("loaded %d favorite server(s) from Steam", len(m.importItems)),
			false,
		)

	case importFavoritesResultMsg:
		if msg.Err != nil {
			return m, m.setStatus(msg.Err.Error(), true)
		}
		m.mode = modeSettingsMenu
		m.importItems = nil
		m.importIndex = 0
		m.importLoading = false
		m.updateTitle()
		return m, tea.Batch(
			m.refreshCmd(),
			m.setStatus(fmt.Sprintf("imported %d server(s) from Steam favorites", msg.Count), false),
		)

	case statusClearMsg:
		if msg.Version == m.statusVersion && m.configErr == nil {
			m.statusText = ""
			m.statusError = false
		}
		return m, nil

	case autoRefreshMsg:
		if msg.Version != m.autoRefreshID {
			return m, nil
		}
		if m.mode != modeBrowse {
			return m, m.autoRefreshCmd()
		}
		return m, tea.Batch(m.refreshCmd(), m.autoRefreshCmd())
	}

	return m, nil
}

func (m *model) refreshCmd() tea.Cmd {
	if m.refreshing {
		return nil
	}

	m.refreshing = true
	m.updateTitle()
	return func() tea.Msg {
		// Re-read the config on every refresh so editing servers.json while the
		// app is open takes effect without restarting the program.
		cfg, err := config.Reload(m.configPath)
		if err != nil {
			return refreshMsg{
				Config: cfg,
				Err:    err,
				When:   time.Now(),
			}
		}

		return refreshMsg{
			Servers: query.QueryAll(cfg.Servers),
			When:    time.Now(),
			Config:  cfg,
		}
	}
}

func (m *model) autoRefreshCmd() tea.Cmd {
	interval := time.Duration(m.config.RefreshInterval) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	version := m.autoRefreshID
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return autoRefreshMsg{Version: version}
	})
}

func (m *model) loadFavoritesCmd() tea.Cmd {
	m.mode = modeImportFavorites
	m.importLoading = true
	m.importItems = nil
	m.importIndex = 0
	m.favoritesLoadID++
	version := m.favoritesLoadID
	m.updateTitle()

	return func() tea.Msg {
		favorites, err := steam.FindFavoriteServers()
		if err != nil {
			return favoritesLoadedMsg{Version: version, Err: err}
		}

		queryTargets := make([]config.ServerConfig, 0, len(favorites))
		for _, favorite := range favorites {
			queryTargets = append(queryTargets, config.ServerConfig{
				IP:   favorite.IP,
				Port: favorite.Port,
			})
		}
		results := query.QueryAll(queryTargets)

		existing := m.existingServerAddresses()
		items := make([]favoriteImportItem, 0, len(favorites))
		for index, favorite := range favorites {
			info := query.ServerInfo{
				IP:   favorite.IP,
				Port: favorite.Port,
				Name: favorite.Address,
			}
			if index < len(results) {
				info = results[index]
			}
			items = append(items, favoriteImportItem{
				Server:       favorite,
				Info:         info,
				AlreadySaved: existing[favorite.Address],
			})
		}

		return favoritesLoadedMsg{Version: version, Items: items}
	}
}

func (m *model) launchServerCmd(server query.ServerInfo) tea.Cmd {
	return func() tea.Msg {
		err := steam.Launch(server.IP, server.Port)
		text := fmt.Sprintf("connecting to %s", m.visibleName(server))
		return launchResultMsg{Text: text, Err: err}
	}
}

func (m *model) setStatus(text string, isError bool) tea.Cmd {
	m.statusText = text
	m.statusError = isError
	m.statusVersion++
	version := m.statusVersion
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{Version: version}
	})
}

func (m *model) restartAutoRefreshCmd() tea.Cmd {
	m.autoRefreshID++
	return m.autoRefreshCmd()
}

func (m *model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeBrowse
		m.updateTitle()
		return m, nil
	case "up", "k":
		if m.settingsIndex > 0 {
			m.settingsIndex--
		} else {
			m.settingsIndex = len(settingsOptions) - 1
		}
		m.updateTitle()
		return m, nil
	case "down", "j":
		if m.settingsIndex < len(settingsOptions)-1 {
			m.settingsIndex++
		} else {
			m.settingsIndex = 0
		}
		m.updateTitle()
		return m, nil
	case "enter":
		switch m.settingsIndex {
		case 0:
			m.mode = modeSavedServers
			m.savedIndex = minInt(m.savedIndex, maxInt(0, len(m.config.Servers)-1))
			m.updateTitle()
			return m, nil
		case 1:
			m.startAddServer()
			return m, nil
		case 2:
			return m, m.loadFavoritesCmd()
		case 3:
			m.startRefreshIntervalEdit()
			return m, nil
		default:
			m.mode = modeBrowse
			m.updateTitle()
			return m, nil
		}
	}

	return m, nil
}

func (m *model) handleSavedServersKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.config.Servers) > 0 {
		m.savedIndex = minInt(maxInt(0, m.savedIndex), len(m.config.Servers)-1)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeSettingsMenu
		m.updateTitle()
		return m, nil
	case "up", "k":
		if len(m.config.Servers) > 0 {
			if m.savedIndex > 0 {
				m.savedIndex--
			} else {
				m.savedIndex = len(m.config.Servers) - 1
			}
			m.updateTitle()
		}
		return m, nil
	case "down", "j":
		if len(m.config.Servers) > 0 {
			if m.savedIndex < len(m.config.Servers)-1 {
				m.savedIndex++
			} else {
				m.savedIndex = 0
			}
			m.updateTitle()
		}
		return m, nil
	case "a":
		m.startAddServer()
		return m, nil
	case "left", "h", "shift+up", "[":
		if len(m.config.Servers) <= 1 {
			return m, m.setStatus("need at least two saved servers to reorder", true)
		}
		if m.savedIndex <= 0 {
			return m, m.setStatus("selected server is already at the top", true)
		}
		m.config.Servers[m.savedIndex], m.config.Servers[m.savedIndex-1] =
			m.config.Servers[m.savedIndex-1], m.config.Servers[m.savedIndex]
		m.savedIndex--
		m.updateTitle()

		moved := m.config.Servers[m.savedIndex]
		name := strings.TrimSpace(moved.Label)
		if name == "" {
			name = fmt.Sprintf("%s:%d", moved.IP, moved.Port)
		}

		return m, m.saveConfigAndRefresh(
			fmt.Sprintf("moved %s to position %d", name, m.savedIndex+1),
			false,
		)
	case "right", "l", "shift+down", "]":
		if len(m.config.Servers) <= 1 {
			return m, m.setStatus("need at least two saved servers to reorder", true)
		}
		if m.savedIndex >= len(m.config.Servers)-1 {
			return m, m.setStatus("selected server is already at the bottom", true)
		}
		m.config.Servers[m.savedIndex], m.config.Servers[m.savedIndex+1] =
			m.config.Servers[m.savedIndex+1], m.config.Servers[m.savedIndex]
		m.savedIndex++
		m.updateTitle()

		moved := m.config.Servers[m.savedIndex]
		name := strings.TrimSpace(moved.Label)
		if name == "" {
			name = fmt.Sprintf("%s:%d", moved.IP, moved.Port)
		}

		return m, m.saveConfigAndRefresh(
			fmt.Sprintf("moved %s to position %d", name, m.savedIndex+1),
			false,
		)
	case "d", "delete", "backspace":
		if len(m.config.Servers) == 0 {
			return m, m.setStatus("there are no saved servers to delete", true)
		}
		m.deleteConfirm = true
		return m, nil
	case "enter":
		if len(m.config.Servers) == 0 {
			return m, m.setStatus("there are no saved servers to edit", true)
		}
		m.startEditServer(m.savedIndex)
		return m, nil
	}

	return m, nil
}

func (m *model) handleServerFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeSavedServers
		m.updateTitle()
		return m, nil
	case "up", "k":
		if m.serverForm.FieldIndex > 0 {
			m.serverForm.FieldIndex--
		} else {
			m.serverForm.FieldIndex = 2
		}
		return m, nil
	case "down", "j", "tab":
		if m.serverForm.FieldIndex < 2 {
			m.serverForm.FieldIndex++
		} else {
			m.serverForm.FieldIndex = 0
		}
		return m, nil
	case "enter":
		return m, m.saveServerForm()
	case "backspace":
		value := []rune(m.serverForm.Values[m.serverForm.FieldIndex])
		if len(value) > 0 {
			m.serverForm.Values[m.serverForm.FieldIndex] = string(value[:len(value)-1])
		}
		return m, nil
	}

	if text := msg.String(); isPrintableInput(text) {
		m.serverForm.Values[m.serverForm.FieldIndex] += text
	}
	return m, nil
}

func (m *model) handleRefreshIntervalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeSettingsMenu
		m.updateTitle()
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.refreshValue)
		interval, err := strconv.Atoi(value)
		if err != nil || interval <= 0 {
			return m, m.setStatus("refresh interval must be a positive number", true)
		}
		m.config.RefreshInterval = interval
		m.mode = modeSettingsMenu
		m.updateTitle()
		return m, tea.Batch(
			m.restartAutoRefreshCmd(),
			m.saveConfigAndRefresh(
				fmt.Sprintf("set refresh interval to %d seconds", interval),
				false,
			),
		)
	case "backspace":
		value := []rune(m.refreshValue)
		if len(value) > 0 {
			m.refreshValue = string(value[:len(value)-1])
		}
		return m, nil
	}

	if text := msg.String(); isPrintableInput(text) {
		if text[0] >= '0' && text[0] <= '9' {
			m.refreshValue += text
		}
	}

	return m, nil
}

func (m *model) handleImportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.importItems) > 0 {
		m.importIndex = minInt(maxInt(0, m.importIndex), len(m.importItems)-1)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeSettingsMenu
		m.importLoading = false
		m.importItems = nil
		m.importIndex = 0
		m.updateTitle()
		return m, nil
	case "up", "k":
		if len(m.importItems) > 0 {
			if m.importIndex > 0 {
				m.importIndex--
			} else {
				m.importIndex = len(m.importItems) - 1
			}
			m.updateTitle()
		}
		return m, nil
	case "down", "j":
		if len(m.importItems) > 0 {
			if m.importIndex < len(m.importItems)-1 {
				m.importIndex++
			} else {
				m.importIndex = 0
			}
			m.updateTitle()
		}
		return m, nil
	case "pgup", "left", "h":
		if len(m.importItems) > 0 {
			m.importIndex = maxInt(0, m.importIndex-m.importPageSize())
			m.updateTitle()
		}
		return m, nil
	case "pgdown", "right", "l":
		if len(m.importItems) > 0 {
			m.importIndex = minInt(len(m.importItems)-1, m.importIndex+m.importPageSize())
			m.updateTitle()
		}
		return m, nil
	case "r":
		return m, m.loadFavoritesCmd()
	case " ":
		if len(m.importItems) == 0 {
			return m, nil
		}
		item := &m.importItems[m.importIndex]
		if item.AlreadySaved {
			return m, m.setStatus(item.Server.Address+" is already saved", true)
		}
		item.Selected = !item.Selected
		return m, nil
	case "enter", "a":
		return m, m.importSelectedFavoritesCmd()
	}

	return m, nil
}

func (m *model) importSelectedFavoritesCmd() tea.Cmd {
	selected := make([]steam.FavoriteServer, 0)
	selectedInfo := make([]query.ServerInfo, 0)
	for _, item := range m.importItems {
		if item.Selected && !item.AlreadySaved {
			selected = append(selected, item.Server)
			selectedInfo = append(selectedInfo, item.Info)
		}
	}

	if len(selected) == 0 {
		return func() tea.Msg {
			return importFavoritesResultMsg{
				Err: fmt.Errorf("select one or more Steam favorite servers to import"),
			}
		}
	}

	cfg := m.config
	for index, favorite := range selected {
		entry := config.ServerConfig{
			IP:   favorite.IP,
			Port: favorite.Port,
		}
		if index < len(selectedInfo) {
			name := strings.TrimSpace(selectedInfo[index].Name)
			if name != "" && name != favorite.Address {
				entry.Label = name
			}
		}
		cfg.Servers = append(cfg.Servers, entry)
	}

	return func() tea.Msg {
		if err := config.Save(m.configPath, cfg); err != nil {
			return importFavoritesResultMsg{Err: err}
		}
		return importFavoritesResultMsg{Count: len(selected)}
	}
}

func (m *model) startAddServer() {
	m.mode = modeAddServer
	m.serverForm = serverFormState{
		Title:       "Add Server",
		ServerIndex: -1,
	}
	m.updateTitle()
}

func (m *model) startEditServer(index int) {
	if index < 0 || index >= len(m.config.Servers) {
		return
	}

	server := m.config.Servers[index]
	m.mode = modeEditServer
	m.serverForm = serverFormState{
		Title:           "Edit Server",
		ServerIndex:     index,
		OriginalAddress: serverConfigAddress(server),
		Values: [3]string{
			server.IP,
			strconv.Itoa(server.Port),
			server.Label,
		},
	}
	m.updateTitle()
}

func (m *model) startRefreshIntervalEdit() {
	m.mode = modeEditRefreshInterval
	m.refreshValue = strconv.Itoa(m.config.RefreshInterval)
	m.updateTitle()
}

func (m *model) confirmDeleteSavedServer() (tea.Model, tea.Cmd) {
	m.deleteConfirm = false
	if len(m.config.Servers) == 0 {
		return m, m.setStatus("there are no saved servers to delete", true)
	}

	m.savedIndex = minInt(maxInt(0, m.savedIndex), len(m.config.Servers)-1)
	server := m.config.Servers[m.savedIndex]
	m.config.Servers = append(m.config.Servers[:m.savedIndex], m.config.Servers[m.savedIndex+1:]...)
	if m.savedIndex >= len(m.config.Servers) && len(m.config.Servers) > 0 {
		m.savedIndex = len(m.config.Servers) - 1
	}
	m.updateTitle()

	return m, m.saveConfigAndRefresh(
		fmt.Sprintf("deleted %s:%d", server.IP, server.Port),
		false,
	)
}

func (m *model) saveServerForm() tea.Cmd {
	ip := strings.TrimSpace(m.serverForm.Values[0])
	portText := strings.TrimSpace(m.serverForm.Values[1])
	label := strings.TrimSpace(m.serverForm.Values[2])

	if ip == "" {
		return m.setStatus("server IP or hostname is required", true)
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port <= 0 || port > 65535 {
		return m.setStatus("port must be between 1 and 65535", true)
	}

	entry := config.ServerConfig{
		IP:    ip,
		Port:  port,
		Label: label,
	}

	status := fmt.Sprintf("added %s:%d", entry.IP, entry.Port)
	if m.mode == modeEditServer {
		serverIndex := -1
		if m.serverForm.ServerIndex >= 0 && m.serverForm.ServerIndex < len(m.config.Servers) &&
			serverConfigAddress(m.config.Servers[m.serverForm.ServerIndex]) == m.serverForm.OriginalAddress {
			serverIndex = m.serverForm.ServerIndex
		} else if m.serverForm.OriginalAddress != "" {
			for index, server := range m.config.Servers {
				if serverConfigAddress(server) == m.serverForm.OriginalAddress {
					serverIndex = index
					break
				}
			}
		}

		if serverIndex < 0 {
			return m.setStatus("the server list changed while editing; reopen the server and try again", true)
		}

		m.config.Servers[serverIndex] = entry
		m.savedIndex = serverIndex
		status = fmt.Sprintf("updated %s:%d", entry.IP, entry.Port)
	} else {
		m.config.Servers = append(m.config.Servers, entry)
		m.savedIndex = len(m.config.Servers) - 1
	}

	m.mode = modeSavedServers
	m.updateTitle()
	return m.saveConfigAndRefresh(status, false)
}

func (m *model) saveConfigAndRefresh(status string, isError bool) tea.Cmd {
	if err := config.Save(m.configPath, m.config); err != nil {
		return m.setStatus(err.Error(), true)
	}

	if isError {
		return tea.Batch(m.refreshCmd(), m.setStatus(status, true))
	}
	return tea.Batch(m.refreshCmd(), m.setStatus(status, false))
}
