package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"tf2-server-tui/internal/config"
	"tf2-server-tui/internal/console"
)

func (m *model) updateTitle() {
	_ = console.SetTitle(m.windowTitle())
}

// windowTitle returns the static terminal title for the application.
func (m *model) windowTitle() string {
	return "TF2 Server Browser"
}

func padRight(value string, width int) string {
	if width <= 0 {
		return ""
	}

	clipped := trimToWidth(value, width)
	padding := width - runewidth.StringWidth(clipped)
	if padding <= 0 {
		return clipped
	}
	return clipped + strings.Repeat(" ", padding)
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return trimToWidth(value, width)
	}
	return trimToWidth(value, width-3) + "..."
}

func trimToWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}

	var builder strings.Builder
	currentWidth := 0
	for _, r := range value {
		runeWidth := runewidth.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}

	return builder.String()
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

func renderHelpItem(key string, label string) string {
	return helpKeyStyle.Render(key) + helpTextStyle.Render(" "+label)
}

func centerLine(value string, width int) string {
	return lipgloss.NewStyle().
		Width(maxInt(1, width)).
		Align(lipgloss.Center).
		Render(value)
}

func (m *model) renderTwoLineFooter(primary []string, secondary []string) string {
	outerWidth := m.layoutWidth()
	contentWidth := m.contentWidth(footerBoxStyle, outerWidth)
	lineWidth := m.lineWidth(footerBoxStyle, outerWidth)

	lines := []string{
		centerLine(strings.Join(primary, "  "), lineWidth),
		centerLine(strings.Join(secondary, "  "), lineWidth),
	}

	return footerBoxStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func isPrintableInput(text string) bool {
	if text == "" {
		return false
	}
	runes := []rune(text)
	return len(runes) == 1 && runes[0] >= 32 && runes[0] != 127
}

func (m *model) importPageSize() int {
	if m.height <= 0 {
		return 8
	}
	return maxInt(4, m.height-16)
}

func (m *model) importPage() int {
	pageSize := m.importPageSize()
	if pageSize <= 0 {
		return 0
	}
	return m.importIndex / pageSize
}

func (m *model) importPageCount(pageSize int) int {
	if pageSize <= 0 || len(m.importItems) == 0 {
		return 1
	}
	return (len(m.importItems) + pageSize - 1) / pageSize
}

func (m *model) existingServerAddresses() map[string]bool {
	addresses := make(map[string]bool, len(m.config.Servers))
	for _, server := range m.config.Servers {
		addresses[serverConfigAddress(server)] = true
	}
	return addresses
}

func (m *model) selectedSavedServerSummary() (string, string) {
	if len(m.config.Servers) == 0 {
		return "No saved server selected", ""
	}

	index := minInt(maxInt(0, m.savedIndex), len(m.config.Servers)-1)
	server := m.config.Servers[index]
	name := strings.TrimSpace(server.Label)
	if name == "" {
		name = fmt.Sprintf("%s:%d", server.IP, server.Port)
	}

	return name, fmt.Sprintf("%s:%d", server.IP, server.Port)
}

func serverConfigAddress(server config.ServerConfig) string {
	return fmt.Sprintf("%s:%d", server.IP, server.Port)
}
