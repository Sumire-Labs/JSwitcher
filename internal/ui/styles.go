/*
 * JSwitcher
 *
 * Copyright 2025 s12kuma01
 *
 * This software is licensed under the Open Software License version
 * 3.0. The full text of this license can be found in https://opensource.org/licenses/OSL-3.0
 * or in the LICENSES directory which is distributed along with the software.
 */

package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	// タイトル・ヘッダー
	Title         lipgloss.Style
	TitleGradient lipgloss.Style
	Subtitle      lipgloss.Style
	Header        lipgloss.Style

	// リストアイテム
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Hover    lipgloss.Style

	// ステータス
	Current     lipgloss.Style
	ActiveBadge lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style

	// コンテナ
	Box        lipgloss.Style
	Card       lipgloss.Style
	CardHeader lipgloss.Style
	Panel      lipgloss.Style

	// バッジ・アクセント
	Badge    lipgloss.Style
	BadgeAlt lipgloss.Style
	Accent   lipgloss.Style
	Muted    lipgloss.Style

	// ボーダー・装飾
	Border    lipgloss.Style
	Divider   lipgloss.Style
	FooterBar lipgloss.Style

	// プログレスバー
	ProgressBar  lipgloss.Style
	ProgressFill lipgloss.Style
}

func NewStyles() Styles {
	// 🎨 モダンカラーパレット
	primary := lipgloss.Color("#7C3AED")   // 紫（プライマリ）
	secondary := lipgloss.Color("#06B6D4") // シアン（セカンダリ）
	success := lipgloss.Color("#10B981")   // エメラルドグリーン
	warning := lipgloss.Color("#F59E0B")   // アンバー
	danger := lipgloss.Color("#EF4444")    // レッド

	// グレースケール
	white := lipgloss.Color("#FFFFFF")
	lightGray := lipgloss.Color("#E5E7EB")
	gray := lipgloss.Color("#9CA3AF")
	darkGray := lipgloss.Color("#374151")
	darkerGray := lipgloss.Color("#1F2937")

	// アクセントカラー
	pink := lipgloss.Color("#EC4899")
	yellow := lipgloss.Color("#FBBF24")

	return Styles{
		// タイトル系
		Title: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1),

		TitleGradient: lipgloss.NewStyle().
			Foreground(primary).
			Background(darkerGray).
			Bold(true).
			Padding(1, 3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary),

		Subtitle: lipgloss.NewStyle().
			Foreground(secondary).
			Italic(true).
			Padding(0, 2),

		Header: lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true).
			Padding(0, 1).
			MarginTop(1),

		// リストアイテム
		Selected: lipgloss.NewStyle().
			Foreground(white).
			Background(primary).
			Bold(true).
			Padding(0, 2).
			MarginLeft(1).
			MarginRight(1),

		Normal: lipgloss.NewStyle().
			Foreground(lightGray).
			Padding(0, 2).
			MarginLeft(1),

		Hover: lipgloss.NewStyle().
			Foreground(white).
			Background(darkGray).
			Padding(0, 2).
			MarginLeft(1),

		// ステータス
		Current: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),

		ActiveBadge: lipgloss.NewStyle().
			Foreground(darkerGray).
			Background(success).
			Bold(true).
			Padding(0, 1).
			MarginLeft(1),

		Error: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(warning).
			Bold(true),

		// コンテナ
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(0, 1).
			MarginTop(0).
			MarginBottom(0),

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1, 2).
			MarginTop(1).
			Background(darkerGray),

		CardHeader: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.Border{Bottom: "─"}).
			BorderForeground(primary).
			BorderBottom(true).
			MarginBottom(1),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(darkGray).
			Padding(1, 2),

		// バッジ
		Badge: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),

		BadgeAlt: lipgloss.NewStyle().
			Foreground(yellow).
			Background(darkerGray).
			Bold(true).
			Padding(0, 1),

		Accent: lipgloss.NewStyle().
			Foreground(pink).
			Bold(true),

		Muted: lipgloss.NewStyle().
			Foreground(gray).
			Italic(true),

		// ボーダー・装飾
		Border: lipgloss.NewStyle().
			Foreground(gray),

		Divider: lipgloss.NewStyle().
			Foreground(darkGray).
			Border(lipgloss.Border{Top: "─"}).
			BorderTop(true).
			MarginTop(1).
			MarginBottom(1),

		FooterBar: lipgloss.NewStyle().
			Foreground(lightGray).
			Background(darkerGray).
			Padding(1, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(primary).
			BorderTop(true),

		// プログレスバー
		ProgressBar: lipgloss.NewStyle().
			Foreground(gray).
			Background(darkGray),

		ProgressFill: lipgloss.NewStyle().
			Foreground(white).
			Background(primary),
	}
}
