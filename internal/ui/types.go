package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"tf2-server-tui/internal/config"
	"tf2-server-tui/internal/query"
	"tf2-server-tui/internal/steam"
)

type refreshMsg struct {
	Servers []query.ServerInfo
	When    time.Time
	Config  config.Config
	Err     error
}

type refreshErrMsg struct {
	Err error
}

type statusClearMsg struct {
	Version int
}

type autoRefreshMsg struct {
	Version int
}

type launchResultMsg struct {
	Text string
	Err  error
}

type favoritesLoadedMsg struct {
	Version int
	Items   []favoriteImportItem
	Err     error
}

type importFavoritesResultMsg struct {
	Count int
	Err   error
}

type screenMode int

const (
	modeBrowse screenMode = iota
	modeSettingsMenu
	modeSavedServers
	modeEditServer
	modeAddServer
	modeImportFavorites
	modeEditRefreshInterval
)

type favoriteImportItem struct {
	Server       steam.FavoriteServer
	Info         query.ServerInfo
	Selected     bool
	AlreadySaved bool
}

type settingsOption struct {
	Title       string
	Description string
}

type serverFormState struct {
	Title           string
	ServerIndex     int
	OriginalAddress string
	FieldIndex      int
	Values          [3]string
}

var settingsOptions = []settingsOption{
	{Title: "Saved Servers", Description: "Edit or delete servers already in servers.json"},
	{Title: "Add Server", Description: "Manually add a new server entry"},
	{Title: "Import Favorites", Description: "Choose servers from Steam favorites to import"},
	{Title: "Refresh Interval", Description: "Change how often the list refreshes"},
	{Title: "Back", Description: "Return to the main server list"},
}

type model struct {
	config          config.Config
	configPath      string
	configErr       error
	servers         []query.ServerInfo
	selectedIndex   int
	refreshing      bool
	lastRefresh     time.Time
	statusText      string
	statusError     bool
	width           int
	height          int
	mode            screenMode
	importItems     []favoriteImportItem
	importIndex     int
	importLoading   bool
	settingsIndex   int
	savedIndex      int
	serverForm      serverFormState
	refreshValue    string
	statusVersion   int
	autoRefreshID   int
	favoritesLoadID int
	quitConfirm     bool
	deleteConfirm   bool
}

var _ tea.Model = (*model)(nil)
