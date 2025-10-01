package ui

import (
	"fmt"
	"strings"
	"time"

	"javaswitcher/internal/java"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	javaInstalls     []java.Installation
	filteredInstalls []java.Installation
	cursor           int
	selected         map[int]struct{}
	loading          bool
	loadingDone      bool
	err              error
	detector         *java.Detector
	switcher         *java.Switcher
	styles           Styles

	// UX improvements
	filterText  string
	filterMode  bool
	history     []string

	// Loading animation
	spinner spinner.Model
}

type JavaDetectedMsg struct {
	Installations []java.Installation
}

type JavaSetMsg struct {
	JavaHome string
	Err      error
}

type JavaRestoreMsg struct {
	Err error
}

func NewModel() Model {
	// スピナーを設定
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))

	return Model{
		javaInstalls:     []java.Installation{},
		filteredInstalls: []java.Installation{},
		selected:         make(map[int]struct{}),
		loading:          true,
		loadingDone:      false,
		detector:         java.NewDetector(),
		switcher:         java.NewSwitcher(),
		styles:           NewStyles(),
		filterMode:       false,
		filterText:       "",
		history:          []string{},
		spinner:          s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.detectJavaInstallations,
		tea.EnterAltScreen,
		m.spinner.Tick,
		// Mouse support enabled by default in modern bubbletea
	)
}

func (m Model) detectJavaInstallations() tea.Msg {
	// Show loading screen with credits for 2 seconds
	time.Sleep(2 * time.Second)

	installations, err := m.detector.DetectInstallations()
	if err != nil {
		return JavaDetectedMsg{Installations: []java.Installation{}}
	}
	return JavaDetectedMsg{Installations: installations}
}

func (m Model) setJavaHome(javaHome string) tea.Cmd {
	return func() tea.Msg {
		err := m.switcher.SetJavaHome(javaHome)
		return JavaSetMsg{JavaHome: javaHome, Err: err}
	}
}

func (m Model) restorePreviousJavaHome() tea.Cmd {
	return func() tea.Msg {
		err := m.switcher.RestorePrevious()
		return JavaRestoreMsg{Err: err}
	}
}

func (m *Model) applyFilter() {
	if m.filterText == "" {
		m.filteredInstalls = m.javaInstalls
	} else {
		m.filteredInstalls = []java.Installation{}
		for _, install := range m.javaInstalls {
			if strings.Contains(strings.ToLower(install.Version), strings.ToLower(m.filterText)) ||
				strings.Contains(strings.ToLower(install.Home), strings.ToLower(m.filterText)) {
				m.filteredInstalls = append(m.filteredInstalls, install)
			}
		}
	}

	// Reset cursor if out of bounds or list is empty
	if len(m.filteredInstalls) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filteredInstalls) {
		m.cursor = len(m.filteredInstalls) - 1 // 最後の要素に設定
	}
}

func (m *Model) addToHistory(javaHome string) {
	// Add to beginning of history, remove duplicates
	newHistory := []string{javaHome}
	for _, item := range m.history {
		if item != javaHome && len(newHistory) < 5 { // Keep last 5
			newHistory = append(newHistory, item)
		}
	}
	m.history = newHistory
}

func (m *Model) getCurrentJava() *java.Installation {
	for _, install := range m.javaInstalls {
		if install.Current {
			return &install
		}
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// スピナーを更新
	m.spinner, cmd = m.spinner.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// フィルターモード時のキー処理
		if m.filterMode {
			switch msg.String() {
			case "esc":
				m.filterMode = false
				m.filterText = ""
				m.applyFilter()
			case "enter":
				m.filterMode = false
				m.applyFilter()
			case "backspace":
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.filterText += msg.String()
					m.applyFilter()
				}
			}
			return m, nil
		}

		// 通常モードのキー処理
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			activeList := m.javaInstalls
			if m.filterText != "" {
				activeList = m.filteredInstalls
			}
			if m.cursor < len(activeList)-1 {
				m.cursor++
			}
		case "enter", " ":
			activeList := m.javaInstalls
			if m.filterText != "" {
				activeList = m.filteredInstalls
			}
			if len(activeList) > 0 && m.cursor < len(activeList) {
				selected := activeList[m.cursor]
				m.addToHistory(selected.Home)
				return m, m.setJavaHome(selected.Home)
			}
		case "r", "R":
			// 前のJAVA_HOMEに復元
			return m, m.restorePreviousJavaHome()
		case "/":
			// フィルターモードに切り替え
			m.filterMode = true
			m.filterText = ""
		}

	// Mouse events disabled to prevent interference with CMD scrolling

	case JavaDetectedMsg:
		m.javaInstalls = msg.Installations
		m.loading = false
		m.loadingDone = true
		m.applyFilter() // Initialize filtered list

	case JavaSetMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			// Update current status
			for i := range m.javaInstalls {
				m.javaInstalls[i].Current = m.javaInstalls[i].Home == msg.JavaHome
			}
		}

	case JavaRestoreMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			// Refresh installations to update current status
			return m, m.detectJavaInstallations
		}
	}

	return m, cmd
}

func (m Model) View() string {
	return m.renderView()
}

// メイン表示
func (m Model) renderView() string {
	var s string

	// ✨ ヘッダー（角丸枠）
	headerText := "  🚀 JavaSwitcher v0.2.6  "
	if m.loading {
		headerText += m.styles.Muted.Render("(検出中...)")
	}
	s += m.styles.TitleGradient.Render(headerText) + "\n"

	// エラー表示
	if m.err != nil {
		errorBox := m.styles.Card.Render(
			m.styles.Error.Render("❌ エラー") + "\n" +
				m.styles.Muted.Render(fmt.Sprintf("詳細: %v", m.err)),
		)
		s += errorBox + "\n"
		return s
	}

	// 🔍 フィルター状況
	if m.filterMode {
		filterBox := m.styles.Card.Render(
			m.styles.Accent.Render("🔍 検索モード") + "\n" +
				m.styles.Selected.Render("┃ "+m.filterText+"█") + "\n" +
				m.styles.Muted.Render("┃ Escでキャンセル"),
		)
		s += filterBox + "\n"
	} else if m.filterText != "" {
		filterBox := m.styles.Box.Render(
			m.styles.Accent.Render("🔍 フィルター適用中: ") +
				m.styles.Badge.Render(m.filterText) +
				m.styles.Muted.Render(fmt.Sprintf(" (%d件表示)", len(m.filteredInstalls))),
		)
		s += filterBox + "\n"
	}

	// Java一覧
	if m.loading {
		loadingBox := m.styles.Card.Render(
			m.spinner.View() + " " + m.styles.Muted.Render("Javaインストールを検出中..."),
		)
		s += loadingBox + "\n"
		return s
	}

	if len(m.javaInstalls) == 0 {
		emptyBox := m.styles.Card.Render(
			m.styles.Error.Render("❌ Javaインストールが見つかりません") + "\n" +
				m.styles.Muted.Render("システム内にJavaがインストールされていない可能性があります"),
		)
		s += emptyBox + "\n"
		return s
	}

	// 📋 選択可能なJava一覧
	activeList := m.javaInstalls
	if m.filterText != "" {
		activeList = m.filteredInstalls
	}

	if len(activeList) == 0 && m.filterText != "" {
		noResultBox := m.styles.Card.Render(
			m.styles.Warning.Render("⚠️  フィルターに一致するJavaが見つかりません") + "\n" +
				m.styles.Muted.Render("別のキーワードで検索してください"),
		)
		s += noResultBox + "\n"
	} else {
		listHeader := "📋 利用可能なJava"
		if len(activeList) != len(m.javaInstalls) {
			listHeader += fmt.Sprintf(" (%d/%d件)", len(activeList), len(m.javaInstalls))
		}
		s += m.styles.CardHeader.Render(listHeader) + "\n"

		for i, install := range activeList {
			// カーソルアイコン
			cursor := "  "
			if m.cursor == i {
				cursor = "▶ "
			}

			// バージョン表示
			version := install.Version
			badgeStr := ""
			if install.Current {
				badgeStr = m.styles.ActiveBadge.Render("ACTIVE")
			}

			line := cursor + version
			if badgeStr != "" {
				line += " " + badgeStr
			}

			// パス表示
			shortPath := m.shortenPath(install.Home)
			pathLine := "\n   " + m.styles.Muted.Render("📁 "+shortPath)

			fullLine := line + pathLine

			if m.cursor == i {
				s += m.styles.Selected.Render(fullLine) + "\n"
			} else {
				s += m.styles.Normal.Render(fullLine) + "\n"
			}
		}
	}

	// 💡 操作ガイド（シンプル表示）
	s += "\n"
	if m.filterMode {
		s += m.styles.Muted.Render("  文字入力でフィルター │ ") +
			m.styles.Accent.Render("Enter") + m.styles.Muted.Render(":確定 │ ") +
			m.styles.Accent.Render("Esc") + m.styles.Muted.Render(":キャンセル")
	} else {
		s += m.styles.Muted.Render("  ") +
			m.styles.Accent.Render("↑↓") + m.styles.Muted.Render(":移動 │ ") +
			m.styles.Accent.Render("Enter") + m.styles.Muted.Render(":選択 │ ") +
			m.styles.Accent.Render("/") + m.styles.Muted.Render(":検索 │ ") +
			m.styles.Accent.Render("r") + m.styles.Muted.Render(":復元 │ ") +
			m.styles.Accent.Render("q") + m.styles.Muted.Render(":終了")
	}

	return s
}

// パスを短縮する
func (m Model) shortenPath(path string) string {
	if len(path) <= 40 {
		return path
	}
	// 最初と最後を残して中間を省略
	if len(path) > 40 {
		return path[:15] + "..." + path[len(path)-20:]
	}
	return path
}
