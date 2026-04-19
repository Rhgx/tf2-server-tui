package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorCyan       = lipgloss.Color("#22D3EE")
	colorGrayBorder = lipgloss.Color("#6B7280")
	colorGrayText   = lipgloss.Color("#94A3B8")
	colorGrayMuted  = lipgloss.Color("#9CA3AF")
	colorWhite      = lipgloss.Color("#F8FAFC")
	colorBlue       = lipgloss.Color("#3B82F6")
	colorGreen      = lipgloss.Color("#22C55E")
	colorGreenSoft  = lipgloss.Color("#86EFAC")
	colorGreenMid   = lipgloss.Color("#4ADE80")
	colorGreenDeep  = lipgloss.Color("#16A34A")
	colorGreenDark  = lipgloss.Color("#15803D")
	colorYellow     = lipgloss.Color("#EAB308")
	colorOrange     = lipgloss.Color("#F97316")
	colorRedSoft    = lipgloss.Color("#F87171")
	colorRed        = lipgloss.Color("#EF4444")
	colorRedDeep    = lipgloss.Color("#DC2626")
	colorRedDark    = lipgloss.Color("#B91C1C")
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
	playerEmptyStyle        = newTextStyle(colorGrayMuted, false)
	mapStyle                = newTextStyle(colorCyan, false)
	messageStyle            = newTextStyle(colorAmber, false)
	autoRefreshStyle        = newTextStyle(colorGrayText, false)
	helpKeyStyle            = newTextStyle(colorAmber, false)
	helpTextStyle           = newTextStyle(colorGrayText, false)
	mutedStyle              = newTextStyle(colorGrayText, false)
	confirmBoxStyle         = newBoxStyle(lipgloss.RoundedBorder(), colorAmber).Padding(1, 2)
	confirmTitleStyle       = newTextStyle(colorWhite, true)
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

func playerCountStyle(players int, maxPlayers int) lipgloss.Style {
	if maxPlayers <= 0 || players <= 0 {
		return playerEmptyStyle
	}
	if players >= maxPlayers {
		return newTextStyle(colorRedDark, false)
	}

	fillRatio := float64(players) / float64(maxPlayers)
	switch {
	case fillRatio >= 0.85:
		return newTextStyle(colorGreenDark, false)
	case fillRatio >= 0.65:
		return newTextStyle(colorGreenDeep, false)
	case fillRatio >= 0.45:
		return newTextStyle(colorGreen, false)
	case fillRatio >= 0.2:
		return newTextStyle(colorGreenMid, false)
	default:
		return newTextStyle(colorGreenSoft, false)
	}
}

func pingValueStyle(ping int) lipgloss.Style {
	switch {
	case ping <= 40:
		return newTextStyle(colorGreenSoft, false)
	case ping <= 70:
		return newTextStyle(colorGreen, false)
	case ping <= 95:
		return newTextStyle(colorYellow, false)
	case ping <= 125:
		return newTextStyle(colorOrange, false)
	case ping <= 170:
		return newTextStyle(colorRedSoft, false)
	case ping <= 240:
		return newTextStyle(colorRedDeep, false)
	default:
		return newTextStyle(colorRedDark, false)
	}
}
