package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tf2-server-tui/internal/config"
	"tf2-server-tui/internal/console"
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

type statusClearMsg struct{}

type autoRefreshMsg struct{}

type launchResultMsg struct {
	Text string
	Err  error
}

type model struct {
	config        config.Config
	configPath    string
	configErr     error
	servers       []query.ServerInfo
	selectedIndex int
	refreshing    bool
	lastRefresh   time.Time
	statusText    string
	statusError   bool
	width         int
	height        int
}

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

func (m *model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
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
			return m, m.openConfigCmd()
		case "enter":
			if len(m.servers) == 0 {
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
		m.refreshing = false
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
		m.updateTitle()
		if len(m.servers) == 0 {
			return m, m.setStatus(fmt.Sprintf("edit %s to add servers", m.configPath), true)
		}
		return m, m.setStatus(
			fmt.Sprintf("refreshed %d server(s) at %s", len(m.servers), msg.When.Format("15:04:05")),
			false,
		)

	case refreshErrMsg:
		m.refreshing = false
		m.updateTitle()
		return m, m.setStatus(msg.Err.Error(), true)

	case launchResultMsg:
		if msg.Err != nil {
			return m, m.setStatus(msg.Err.Error(), true)
		}
		return m, m.setStatus(msg.Text, false)

	case statusClearMsg:
		if m.configErr == nil {
			m.statusText = ""
			m.statusError = false
		}
		return m, nil

	case autoRefreshMsg:
		return m, tea.Batch(m.refreshCmd(), m.autoRefreshCmd())
	}

	return m, nil
}

func (m *model) View() string {
	parts := []string{
		m.renderHeader(),
		"",
		m.renderTable(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts,
		"",
		m.renderFooter(),
		"",
		m.autoRefreshLine(),
	)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return autoRefreshMsg{}
	})
}

func (m *model) openConfigCmd() tea.Cmd {
	return func() tea.Msg {
		err := steam.OpenPath(m.configPath)
		return launchResultMsg{
			Text: fmt.Sprintf("opened %s", m.configPath),
			Err:  err,
		}
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
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

var (
	colorCyan       = lipgloss.Color("#22D3EE")
	colorGrayBorder = lipgloss.Color("#6B7280")
	colorGrayText   = lipgloss.Color("#94A3B8")
	colorGrayMuted  = lipgloss.Color("#9CA3AF")
	colorWhite      = lipgloss.Color("#F8FAFC")
	colorBlue       = lipgloss.Color("#3B82F6")
	colorGreen      = lipgloss.Color("#22C55E")
	colorYellow     = lipgloss.Color("#EAB308")
	colorRed        = lipgloss.Color("#EF4444")
	colorAmber      = lipgloss.Color("#FACC15")

	headerBoxStyle          = newBoxStyle(lipgloss.DoubleBorder(), colorCyan)
	tableBoxStyle           = newBoxStyle(lipgloss.NormalBorder(), colorGrayBorder)
	footerBoxStyle          = newBoxStyle(lipgloss.NormalBorder(), colorGrayBorder)
	headerTitleStyle        = newTextStyle(colorCyan, true)
	headerSubtitleStyle     = newTextStyle(colorGrayText, false)
	tableHeaderStyle        = newTextStyle(colorWhite, true)
	tableSeparatorStyle     = newTextStyle(colorGrayBorder, false)
	selectedRowStyle        = newSelectedRowStyle(colorBlue)
	selectedOfflineRowStyle = newSelectedRowStyle(colorGrayBorder)
	offlineStyle            = newTextStyle(colorGrayMuted, false)
	playerActiveStyle       = newTextStyle(colorGreen, false)
	playerEmptyStyle        = newTextStyle(colorGrayMuted, false)
	pingGoodStyle           = newTextStyle(colorGreen, false)
	pingWarnStyle           = newTextStyle(colorYellow, false)
	pingBadStyle            = newTextStyle(colorRed, false)
	mapStyle                = newTextStyle(colorCyan, false)
	messageStyle            = newTextStyle(colorAmber, false)
	autoRefreshStyle        = newTextStyle(colorGrayText, false)
	helpKeyStyle            = newTextStyle(colorAmber, false)
	helpTextStyle           = newTextStyle(colorGrayText, false)
	mutedStyle              = newTextStyle(colorGrayText, false)
)

func newBoxStyle(border lipgloss.Border, borderColor lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(borderColor).
		Padding(0, 1)
}

func newTextStyle(foreground lipgloss.Color, bold bool) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(bold).
		Foreground(foreground)
}

func newSelectedRowStyle(background lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorWhite).
		Background(background)
}

func (m *model) renderHeader() string {
	status := "Loading..."
	if m.refreshing {
		status = "Refreshing..."
	} else if !m.lastRefresh.IsZero() {
		status = "Last updated: " + m.lastRefresh.Format("15:04:05")
	}

	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(headerBoxStyle, outerWidth)
	lineWidth := m.lineWidth(headerBoxStyle, outerWidth)
	lines := []string{
		centerLine(headerTitleStyle.Render("TF2 SERVER BROWSER"), lineWidth),
		centerLine(headerSubtitleStyle.Render(status), lineWidth),
	}

	return headerBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderTable() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)
	layout := m.tableLayout(lineWidth)

	lines := []string{
		m.renderTableHeader(layout),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
	}

	if len(m.servers) == 0 {
		lines = append(lines, centerLine(mutedStyle.Render("No servers configured. Press E to edit servers.json."), lineWidth))
	} else {
		for index, server := range m.servers {
			lines = append(lines, m.renderTableRow(server, index, layout))
		}
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderFooter() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(footerBoxStyle, outerWidth)
	lineWidth := m.lineWidth(footerBoxStyle, outerWidth)

	primary := []string{
		renderHelpItem("[Up/Down]", "Navigate"),
		renderHelpItem("[Enter]", "Connect"),
		renderHelpItem("[R]", "Refresh"),
	}
	secondary := []string{
		renderHelpItem("[E]", "Edit Config"),
		renderHelpItem("[Q]", "Quit"),
	}

	lines := []string{
		centerLine(strings.Join(primary, "  "), lineWidth),
	}
	if lineWidth >= 42 {
		lines = append(lines, centerLine(strings.Join(secondary, "  "), lineWidth))
	} else {
		lines[0] = centerLine(strings.Join(append(primary, secondary...), "  "), lineWidth)
	}

	return footerBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) visibleName(server query.ServerInfo) string {
	if server.Label != "" {
		return server.Label
	}
	return server.Name
}

func (m *model) updateTitle() {
	_ = console.SetTitle(m.windowTitle())
}

// windowTitle computes the live terminal title text. It shows either refresh
// activity or the currently selected server and its latest known state so the
// title stays useful even when the window is unfocused or partially covered.
func (m *model) windowTitle() string {
	base := "TF2 Server Browser"

	if m.refreshing {
		return base + " - Refreshing..."
	}

	if len(m.servers) == 0 {
		return base
	}

	index := m.selectedIndex
	if index < 0 || index >= len(m.servers) {
		return base
	}

	server := m.servers[index]
	name := m.visibleName(server)
	if name == "" {
		return base
	}

	status := "Offline"
	if server.Online {
		status = fmt.Sprintf("%d/%d - %dms", server.Players, server.MaxPlayers, server.Ping)
	}

	return fmt.Sprintf("%s - %s - %s", base, name, status)
}

func padRight(value string, width int) string {
	runes := []rune(value)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return value + strings.Repeat(" ", width-len(runes))
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// layoutWidth computes the outer width budget for the rendered UI. It leaves a
// little breathing room against the terminal edges and falls back to a
// sensible default before Bubble Tea reports the real window size.
func (m *model) layoutWidth() int {
	if m.width <= 0 {
		return 78
	}
	return maxInt(24, m.width-2)
}

func (m *model) contentWidth(style lipgloss.Style, outerWidth int) int {
	return maxInt(1, outerWidth-style.GetHorizontalFrameSize())
}

// lineWidth computes the printable width inside a styled box by subtracting
// both borders and padding so row and separator calculations land on the true
// inner width.
func (m *model) lineWidth(style lipgloss.Style, outerWidth int) int {
	return maxInt(1, m.contentWidth(style, outerWidth)-style.GetHorizontalPadding())
}

func (m *model) centeredMessage() string {
	if m.statusText == "" {
		return ""
	}
	return centerLine(messageStyle.Render(truncate(m.statusText, m.layoutWidth())), m.layoutWidth())
}

func (m *model) autoRefreshLine() string {
	return centerLine(
		autoRefreshStyle.Render(fmt.Sprintf("Auto-refresh every %d seconds", m.config.RefreshInterval)),
		m.layoutWidth(),
	)
}

type tableLayout struct {
	nameWidth    int
	playersWidth int
	pingWidth    int
	mapWidth     int
	showPing     bool
	showMap      bool
}

// tableLayout chooses responsive column widths for the server table. On
// narrower terminals it progressively drops columns instead of letting the
// table wrap or overflow awkwardly.
func (m *model) tableLayout(contentWidth int) tableLayout {
	layout := tableLayout{
		playersWidth: 7,
		pingWidth:    6,
		mapWidth:     20,
		showPing:     true,
		showMap:      true,
	}

	switch {
	case contentWidth >= 44:
		layout.mapWidth = minInt(20, maxInt(12, contentWidth/4))
		layout.nameWidth = contentWidth - 2 - layout.playersWidth - layout.pingWidth - layout.mapWidth - 6
		if layout.nameWidth < 16 {
			layout.mapWidth = maxInt(10, layout.mapWidth-(16-layout.nameWidth))
			layout.nameWidth = contentWidth - 2 - layout.playersWidth - layout.pingWidth - layout.mapWidth - 6
		}
	case contentWidth >= 30:
		layout.showMap = false
		layout.nameWidth = contentWidth - 2 - layout.playersWidth - layout.pingWidth - 4
	default:
		layout.showMap = false
		layout.showPing = false
		layout.nameWidth = contentWidth - 2 - layout.playersWidth - 2
	}

	layout.nameWidth = maxInt(8, layout.nameWidth)
	return layout
}

func (m *model) renderTableHeader(layout tableLayout) string {
	parts := []string{
		"  ",
		padRight("Server Name", layout.nameWidth),
		"  ",
		padRight("Players", layout.playersWidth),
	}
	if layout.showPing {
		parts = append(parts, "  ", padRight("Ping", layout.pingWidth))
	}
	if layout.showMap {
		parts = append(parts, "  ", padRight("Map", layout.mapWidth))
	}

	return tableHeaderStyle.Render(strings.Join(parts, ""))
}

// renderTableRow renders one server row for the current table layout. Selected
// rows are styled as a single block so the highlight spans the full row width,
// while unselected rows color individual columns.
func (m *model) renderTableRow(server query.ServerInfo, index int, layout tableLayout) string {
	cursor := "  "
	if index == m.selectedIndex {
		cursor = "> "
	}

	nameField := padRight(truncate(m.visibleName(server), layout.nameWidth), layout.nameWidth)
	if !server.Online {
		plainRow := cursor + nameField
		remaining := layout.playersWidth
		if layout.showPing {
			remaining += 2 + layout.pingWidth
		}
		if layout.showMap {
			remaining += 2 + layout.mapWidth
		}
		plainRow += "  " + padRight("OFFLINE", remaining)

		if index == m.selectedIndex {
			return selectedOfflineRowStyle.Render(plainRow)
		}

		prefix := offlineStyle.Render(cursor + nameField)
		suffix := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("  " + padRight("OFFLINE", remaining))
		return prefix + suffix
	}

	playersText := fmt.Sprintf("%02d/%02d", server.Players, server.MaxPlayers)
	playersStyle := playerEmptyStyle
	if server.Players > 0 {
		playersStyle = playerActiveStyle
	}

	pingText := fmt.Sprintf("%dms", server.Ping)
	pingStyle := pingBadStyle
	switch {
	case server.Ping < 50:
		pingStyle = pingGoodStyle
	case server.Ping < 100:
		pingStyle = pingWarnStyle
	}

	plainParts := []string{
		cursor,
		nameField,
		"  ",
		padRight(playersText, layout.playersWidth),
	}

	if layout.showPing {
		plainParts = append(plainParts, "  ", padRight(pingText, layout.pingWidth))
	}
	if layout.showMap {
		plainParts = append(plainParts, "  ", padRight(truncate(server.Map, layout.mapWidth), layout.mapWidth))
	}

	if index == m.selectedIndex {
		return selectedRowStyle.Render(strings.Join(plainParts, ""))
	}

	parts := []string{
		cursor,
		nameField,
		"  ",
		playersStyle.Render(padRight(playersText, layout.playersWidth)),
	}
	if layout.showPing {
		parts = append(parts, "  ", pingStyle.Render(padRight(pingText, layout.pingWidth)))
	}
	if layout.showMap {
		parts = append(parts, "  ", mapStyle.Render(padRight(truncate(server.Map, layout.mapWidth), layout.mapWidth)))
	}

	return strings.Join(parts, "")
}

func renderHelpItem(key string, label string) string {
	return helpKeyStyle.Render(key) + helpTextStyle.Render(" "+label)
}

func centerLine(value string, width int) string {
	return lipgloss.NewStyle().
		Width(maxInt(1, width)).
		Align(lipgloss.Center).
		Render(value)
}
