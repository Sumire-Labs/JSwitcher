package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title      lipgloss.Style
	Header     lipgloss.Style
	Selected   lipgloss.Style
	Normal     lipgloss.Style
	Current    lipgloss.Style
	Error      lipgloss.Style

	// 新しいスタイル
	Box        lipgloss.Style
	Card       lipgloss.Style
	Badge      lipgloss.Style
	BadgeAlt   lipgloss.Style  // 代替バッジスタイル
	Accent     lipgloss.Style
	Muted      lipgloss.Style
	Border     lipgloss.Style
}

func NewStyles() Styles {
	// カラーパレット
	primaryBlue := lipgloss.Color("#00D7FF")
	successGreen := lipgloss.Color("#00D700")
	warningOrange := lipgloss.Color("#FF8C00")
	errorRed := lipgloss.Color("#FF5555")
	neutralGray := lipgloss.Color("#8A8A8A")
	darkGray := lipgloss.Color("#404040")
	lightGray := lipgloss.Color("#D3D3D3")

	// 新しいバッジ用カラー
	softGreen := lipgloss.Color("#98FB98")   // 薄い緑色（目に優しい）
	darkGreen := lipgloss.Color("#228B22")  // 濃い緑色（テキスト用）

	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(primaryBlue).
			Bold(true).
			Padding(1, 2),

		Header: lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true).
			Padding(0, 2),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryBlue).
			Bold(true).
			Padding(0, 2).
			Margin(0, 1),

		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2),

		Current: lipgloss.NewStyle().
			Foreground(successGreen).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(errorRed).
			Bold(true),

		// 新しいスタイル
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(neutralGray).
			Padding(1, 2).
			Margin(1, 0),

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryBlue).
			Padding(1, 2).
			Margin(0, 1).
			Background(darkGray),

		Badge: lipgloss.NewStyle().
			Foreground(darkGreen).
			Background(softGreen).
			Padding(0, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(darkGreen).
			Bold(true).
			Italic(false),

		BadgeAlt: lipgloss.NewStyle().
			Foreground(successGreen).
			Padding(0, 1).
			Bold(true).
			Italic(true),

		Accent: lipgloss.NewStyle().
			Foreground(warningOrange).
			Bold(true),

		Muted: lipgloss.NewStyle().
			Foreground(neutralGray),

		Border: lipgloss.NewStyle().
			Foreground(neutralGray),
	}
}