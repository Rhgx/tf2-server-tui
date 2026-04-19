package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"tf2-server-tui/internal/query"
)

func (m *model) View() string {
	var view string

	switch m.mode {
	case modeImportFavorites:
		view = m.importView()
	case modeSettingsMenu:
		view = m.settingsView()
	case modeSavedServers:
		view = m.savedServersView()
	case modeEditServer, modeAddServer:
		view = m.serverFormView()
	case modeEditRefreshInterval:
		view = m.refreshIntervalView()
	default:
		view = m.browseView()
	}

	if m.deleteConfirm {
		return m.renderDeleteConfirm()
	}

	if m.quitConfirm {
		return m.renderQuitConfirm()
	}

	return view
}

func (m *model) browseView() string {
	parts := []string{
		m.renderTable(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts,
		"",
		m.renderFooter(),
		"",
		m.lastUpdatedLine(),
	)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) renderQuitConfirm() string {
	width := m.layoutWidth()
	height := maxInt(9, m.height)
	outerWidth := minInt(46, maxInt(28, width-6))
	outerWidth = minInt(outerWidth, width)
	contentWidth := maxInt(1, outerWidth-confirmBoxStyle.GetHorizontalFrameSize())
	lineWidth := maxInt(1, contentWidth-confirmBoxStyle.GetHorizontalPadding())

	lines := []string{
		centerLine(confirmTitleStyle.Render("Quit TF2 Server Browser?"), lineWidth),
		"",
		centerLine(mutedStyle.Render("Are you sure you want to quit?"), lineWidth),
		"",
		centerLine(renderHelpItem("[Enter/Y]", "Quit"), lineWidth),
		centerLine(renderHelpItem("[Esc/N]", "Cancel"), lineWidth),
	}

	modal := confirmBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

func (m *model) renderDeleteConfirm() string {
	width := m.layoutWidth()
	height := maxInt(9, m.height)
	outerWidth := minInt(52, maxInt(32, width-6))
	outerWidth = minInt(outerWidth, width)
	contentWidth := maxInt(1, outerWidth-confirmBoxStyle.GetHorizontalFrameSize())
	lineWidth := maxInt(1, contentWidth-confirmBoxStyle.GetHorizontalPadding())
	serverName, serverAddress := m.selectedSavedServerSummary()

	lines := []string{
		centerLine(confirmTitleStyle.Render("Delete Saved Server?"), lineWidth),
		"",
		centerLine(mutedStyle.Render("Are you sure you want to delete this server?"), lineWidth),
		"",
		centerLine(tableHeaderStyle.Render(truncate(serverName, lineWidth)), lineWidth),
		centerLine(mutedStyle.Render(truncate(serverAddress, lineWidth)), lineWidth),
		"",
		centerLine(renderHelpItem("[Enter/Y]", "Delete"), lineWidth),
		centerLine(renderHelpItem("[Esc/N]", "Cancel"), lineWidth),
	}

	modal := confirmBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
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
		message := "No servers configured. Press E to open settings."
		if len(m.config.Servers) > 0 {
			if m.refreshing {
				message = "Loading configured servers..."
			} else {
				message = "Configured servers are still loading..."
			}
		}
		lines = append(lines, centerLine(mutedStyle.Render(message), lineWidth))
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
		renderHelpItem("[E]", "Settings"),
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

func (m *model) lastUpdatedLine() string {
	lastUpdated := "--:--:--"
	if !m.lastRefresh.IsZero() {
		lastUpdated = m.lastRefresh.Format("15:04:05")
	}
	return centerLine(
		autoRefreshStyle.Render("Last updated: "+lastUpdated),
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
		suffix := lipgloss.NewStyle().Foreground(colorRed).Render("  " + padRight("OFFLINE", remaining))
		return prefix + suffix
	}

	playersText := fmt.Sprintf("%02d/%02d", server.Players, server.MaxPlayers)
	playersStyle := playerCountStyle(server.Players, server.MaxPlayers)

	pingText := fmt.Sprintf("%dms", server.Ping)
	pingStyle := pingValueStyle(server.Ping)

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
