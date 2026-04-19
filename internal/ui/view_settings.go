package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) settingsView() string {
	parts := []string{
		m.renderSettingsPanel(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts, "", m.renderSettingsFooter(), "", m.lastUpdatedLine())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) savedServersView() string {
	parts := []string{
		m.renderSavedServersPanel(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts, "", m.renderSavedServersFooter(), "", m.lastUpdatedLine())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) serverFormView() string {
	parts := []string{
		m.renderServerFormPanel(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts, "", m.renderServerFormFooter(), "", m.lastUpdatedLine())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) refreshIntervalView() string {
	parts := []string{
		m.renderRefreshIntervalPanel(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts, "", m.renderRefreshIntervalFooter(), "", m.lastUpdatedLine())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) importView() string {
	parts := []string{
		m.renderImportPanel(),
	}
	if message := m.centeredMessage(); message != "" {
		parts = append(parts, "", message)
	}
	parts = append(parts, "", m.renderImportFooter(), "", m.lastUpdatedLine())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *model) renderSettingsPanel() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)

	lines := []string{
		centerLine(tableHeaderStyle.Render("SETTINGS"), lineWidth),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
	}

	for index, option := range settingsOptions {
		cursor := "  "
		if index == m.settingsIndex {
			cursor = "> "
		}
		row := cursor + option.Title
		desc := mutedStyle.Render(option.Description)
		line := row + "\n" + "   " + desc

		if index == m.settingsIndex {
			line = selectedRowStyle.Render(row) + "\n" + "   " + desc
		}
		lines = append(lines, line)
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n\n"))
}

func (m *model) renderSettingsFooter() string {
	return m.renderTwoLineFooter(
		[]string{
			renderHelpItem("[Up/Down]", "Move"),
			renderHelpItem("[Enter]", "Open"),
		},
		[]string{
			renderHelpItem("[Esc]", "Back"),
			renderHelpItem("[Q]", "Quit"),
		},
	)
}

func (m *model) renderSavedServersPanel() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)
	layout := m.savedServerLayout(lineWidth)

	lines := []string{
		centerLine(tableHeaderStyle.Render("SAVED SERVERS"), lineWidth),
		m.renderSavedServerHeader(layout),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
	}

	if len(m.config.Servers) == 0 {
		lines = append(lines,
			centerLine(mutedStyle.Render("No saved servers yet. Press A to add one."), lineWidth),
		)
	} else {
		for index, server := range m.config.Servers {
			label := server.Label
			if strings.TrimSpace(label) == "" {
				label = fmt.Sprintf("%s:%d", server.IP, server.Port)
			}
			address := fmt.Sprintf("%s:%d", server.IP, server.Port)
			lines = append(lines, m.renderSavedServerRow(index, label, address, layout))
		}
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderSavedServersFooter() string {
	return m.renderTwoLineFooter(
		[]string{
			renderHelpItem("[Up/Down]", "Move"),
			renderHelpItem("[Left/Right]", "Reorder"),
			renderHelpItem("[Enter]", "Edit"),
			renderHelpItem("[A]", "Add"),
		},
		[]string{
			renderHelpItem("[D]", "Delete"),
			renderHelpItem("[Esc]", "Back"),
			renderHelpItem("[Q]", "Quit"),
		},
	)
}

func (m *model) renderServerFormPanel() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)

	fieldNames := []string{"IP / Host", "Port", "Label"}
	lines := []string{
		centerLine(tableHeaderStyle.Render(strings.ToUpper(m.serverForm.Title)), lineWidth),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
		mutedStyle.Render("Type directly into the selected field. Label is optional."),
		"",
	}

	for index, name := range fieldNames {
		value := m.serverForm.Values[index]
		if value == "" {
			value = mutedStyle.Render("<empty>")
		}
		label := fmt.Sprintf("%s: %v", name, value)
		if index == m.serverForm.FieldIndex {
			label = selectedRowStyle.Render("> " + truncate(fmt.Sprintf("%s: %s", name, m.serverForm.Values[index]), lineWidth-2))
		}
		lines = append(lines, label)
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderServerFormFooter() string {
	return m.renderTwoLineFooter(
		[]string{
			renderHelpItem("[Up/Down]", "Field"),
			renderHelpItem("[Type]", "Edit"),
			renderHelpItem("[Backspace]", "Delete"),
		},
		[]string{
			renderHelpItem("[Enter]", "Save"),
			renderHelpItem("[Esc]", "Cancel"),
			renderHelpItem("[Q]", "Quit"),
		},
	)
}

func (m *model) renderRefreshIntervalPanel() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)

	lines := []string{
		centerLine(tableHeaderStyle.Render("REFRESH INTERVAL"), lineWidth),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
		mutedStyle.Render("Enter the number of seconds between automatic refreshes."),
		"",
		selectedRowStyle.Render("> Seconds: " + m.refreshValue),
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderRefreshIntervalFooter() string {
	return m.renderTwoLineFooter(
		[]string{
			renderHelpItem("[Type]", "Edit"),
			renderHelpItem("[Backspace]", "Delete"),
		},
		[]string{
			renderHelpItem("[Enter]", "Save"),
			renderHelpItem("[Esc]", "Back"),
			renderHelpItem("[Q]", "Quit"),
		},
	)
}

func (m *model) renderImportPanel() string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(tableBoxStyle, outerWidth)
	lineWidth := m.lineWidth(tableBoxStyle, outerWidth)
	layout := m.importTableLayout(lineWidth)

	pageSize := m.importPageSize()
	page := m.importPage()
	pageCount := m.importPageCount(pageSize)
	start := page * pageSize
	end := minInt(len(m.importItems), start+pageSize)

	lines := []string{
		centerLine(tableHeaderStyle.Render("STEAM FAVORITES IMPORT"), lineWidth),
		centerLine(mutedStyle.Render(fmt.Sprintf("Page %d/%d", page+1, maxInt(1, pageCount))), lineWidth),
		centerLine(mutedStyle.Render("[x] selected   [*] already saved"), lineWidth),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
		m.renderImportHeader(layout),
		tableSeparatorStyle.Render(strings.Repeat("-", lineWidth)),
	}

	switch {
	case m.importLoading:
		lines = append(lines, centerLine(mutedStyle.Render("Loading Steam favorites..."), lineWidth))
	case len(m.importItems) == 0:
		lines = append(lines, centerLine(mutedStyle.Render("No Steam favorites found."), lineWidth))
	default:
		for index := start; index < end; index++ {
			item := m.importItems[index]
			lines = append(lines, m.renderImportRow(item, index, layout))
		}
	}

	return tableBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m *model) renderImportRow(item favoriteImportItem, index int, layout importTableLayout) string {
	cursor := "  "
	if index == m.importIndex {
		cursor = "> "
	}

	marker := "[ ]"
	if item.Selected {
		marker = "[x]"
	}
	if item.AlreadySaved {
		marker = "[*]"
	}

	name := strings.TrimSpace(item.Info.Name)
	if name == "" || name == item.Server.Address {
		name = "-"
	}

	playersText := "--/--"
	pingText := "--"
	playersStyle := mutedStyle
	pingStyle := mutedStyle

	if item.Info.Online {
		playersText = fmt.Sprintf("%02d/%02d", item.Info.Players, item.Info.MaxPlayers)
		playersStyle = playerCountStyle(item.Info.Players, item.Info.MaxPlayers)

		pingText = fmt.Sprintf("%dms", item.Info.Ping)
		pingStyle = pingValueStyle(item.Info.Ping)
	}

	plainParts := []string{
		cursor,
		padRight(marker, layout.pickWidth),
		"  ",
		padRight(truncate(name, layout.nameWidth), layout.nameWidth),
		"  ",
		padRight(playersText, layout.playersWidth),
	}
	if layout.showPing {
		plainParts = append(plainParts, "  ", padRight(pingText, layout.pingWidth))
	}
	if layout.showAddress {
		plainParts = append(plainParts, "  ", padRight(truncate(item.Server.Address, layout.addressWidth), layout.addressWidth))
	}

	row := strings.Join(plainParts, "")

	switch {
	case index == m.importIndex:
		if item.AlreadySaved {
			return selectedOfflineRowStyle.Render(row)
		}
		return selectedRowStyle.Render(row)
	}

	parts := []string{
		cursor,
		padRight(marker, layout.pickWidth),
		"  ",
	}
	nameField := padRight(truncate(name, layout.nameWidth), layout.nameWidth)
	if item.AlreadySaved {
		parts = append(parts, offlineStyle.Render(nameField))
	} else {
		parts = append(parts, nameField)
	}
	parts = append(parts, "  ", playersStyle.Render(padRight(playersText, layout.playersWidth)))
	if layout.showPing {
		parts = append(parts, "  ", pingStyle.Render(padRight(pingText, layout.pingWidth)))
	}
	if layout.showAddress {
		addressField := padRight(truncate(item.Server.Address, layout.addressWidth), layout.addressWidth)
		if item.AlreadySaved {
			parts = append(parts, "  ", mutedStyle.Render(addressField))
		} else {
			parts = append(parts, "  ", offlineStyle.Render(addressField))
		}
	}

	return strings.Join(parts, "")
}

func (m *model) renderImportFooter() string {
	return m.renderTwoLineFooter(
		[]string{
			renderHelpItem("[Up/Down]", "Move"),
			renderHelpItem("[Space]", "Toggle"),
			renderHelpItem("[Enter]", "Import"),
		},
		[]string{
			renderHelpItem("[PgUp/PgDn]", "Page"),
			renderHelpItem("[R]", "Reload"),
			renderHelpItem("[Esc]", "Back"),
		},
	)
}

type savedServerLayout struct {
	indexWidth   int
	nameWidth    int
	addressWidth int
	showAddress  bool
}

func (m *model) savedServerLayout(contentWidth int) savedServerLayout {
	layout := savedServerLayout{
		indexWidth:   4,
		addressWidth: minInt(26, maxInt(18, contentWidth/3)),
		showAddress:  contentWidth >= 46,
	}

	if layout.showAddress {
		layout.nameWidth = contentWidth - 2 - layout.indexWidth - 1 - 2 - layout.addressWidth
		if layout.nameWidth < 16 {
			layout.showAddress = false
		}
	}

	if !layout.showAddress {
		layout.nameWidth = contentWidth - 2 - layout.indexWidth - 1
	}

	layout.nameWidth = maxInt(10, layout.nameWidth)
	return layout
}

func (m *model) renderSavedServerHeader(layout savedServerLayout) string {
	parts := []string{
		"  ",
		padRight("#", layout.indexWidth),
		" ",
		padRight("Server Name", layout.nameWidth),
	}
	if layout.showAddress {
		parts = append(parts, "  ", padRight("Address", layout.addressWidth))
	}
	return tableHeaderStyle.Render(strings.Join(parts, ""))
}

func (m *model) renderSavedServerRow(index int, label string, address string, layout savedServerLayout) string {
	cursor := "  "
	if index == m.savedIndex {
		cursor = "> "
	}

	plainParts := []string{
		cursor,
		padRight(fmt.Sprintf("%02d.", index+1), layout.indexWidth),
		" ",
		padRight(truncate(label, layout.nameWidth), layout.nameWidth),
	}
	if layout.showAddress {
		plainParts = append(plainParts, "  ", padRight(truncate(address, layout.addressWidth), layout.addressWidth))
	}

	if index == m.savedIndex {
		return selectedRowStyle.Render(strings.Join(plainParts, ""))
	}

	parts := []string{
		cursor,
		padRight(fmt.Sprintf("%02d.", index+1), layout.indexWidth),
		" ",
		padRight(truncate(label, layout.nameWidth), layout.nameWidth),
	}
	if layout.showAddress {
		parts = append(parts, "  ", mutedStyle.Render(padRight(truncate(address, layout.addressWidth), layout.addressWidth)))
	}

	return strings.Join(parts, "")
}

type importTableLayout struct {
	pickWidth    int
	nameWidth    int
	playersWidth int
	pingWidth    int
	addressWidth int
	showPing     bool
	showAddress  bool
}

func (m *model) importTableLayout(contentWidth int) importTableLayout {
	layout := importTableLayout{
		pickWidth:    4,
		playersWidth: 12,
		pingWidth:    6,
		addressWidth: minInt(24, maxInt(17, contentWidth/3)),
		showPing:     true,
		showAddress:  true,
	}

	layout.nameWidth = m.importNameWidth(contentWidth, layout)
	if layout.nameWidth < 16 {
		layout.showAddress = false
		layout.nameWidth = m.importNameWidth(contentWidth, layout)
	}
	if layout.nameWidth < 12 {
		layout.showPing = false
		layout.nameWidth = m.importNameWidth(contentWidth, layout)
	}

	layout.nameWidth = maxInt(10, layout.nameWidth)
	return layout
}

func (m *model) importNameWidth(contentWidth int, layout importTableLayout) int {
	width := contentWidth - 2 - layout.pickWidth - 2 - layout.playersWidth
	if layout.showPing {
		width -= 2 + layout.pingWidth
	}
	if layout.showAddress {
		width -= 2 + layout.addressWidth
	}
	return width
}

func (m *model) renderImportHeader(layout importTableLayout) string {
	parts := []string{
		"  ",
		padRight("Pick", layout.pickWidth),
		"  ",
		padRight("Name", layout.nameWidth),
		"  ",
		padRight("Player Count", layout.playersWidth),
	}
	if layout.showPing {
		parts = append(parts, "  ", padRight("Ping", layout.pingWidth))
	}
	if layout.showAddress {
		parts = append(parts, "  ", padRight("Server IP:Port", layout.addressWidth))
	}
	return tableHeaderStyle.Render(strings.Join(parts, ""))
}
