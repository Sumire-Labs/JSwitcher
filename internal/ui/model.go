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
	showPreview bool
	filterText  string
	filterMode  bool
	history     []string
	compactMode bool
	showCredits bool

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
		showPreview:      true,
		filterMode:       false,
		filterText:       "",
		history:          []string{},
		compactMode:      true,  // デフォルトでコンパクトモード
		showCredits:      false, // デフォルトでクレジット非表示
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

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredInstalls) {
		m.cursor = 0
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
			if len(activeList) > 0 {
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
		case "p", "P":
			// プレビュー表示切り替え
			m.showPreview = !m.showPreview
		case "c", "C":
			// コンパクトモード切り替え
			m.compactMode = !m.compactMode
		case "i", "I":
			// クレジット情報表示切り替え
			m.showCredits = !m.showCredits
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
	if m.compactMode {
		return m.renderCompactView()
	}
	return m.renderDetailedView()
}

// コンパクトモード表示
func (m Model) renderCompactView() string {
	var s string

	// ヘッダー
	header := fmt.Sprintf("🔄 JavaSwitcher v0.2.4")
	if m.loading {
		header += " " + m.styles.Muted.Render("(検出中...)")
	}
	s += m.styles.Title.Render(header) + "\n"

	// エラー表示
	if m.err != nil {
		s += m.styles.Error.Render(fmt.Sprintf("❌ エラー: %v", m.err)) + "\n"
		return s
	}

	// 現在のJava表示
	currentJava := m.getCurrentJava()
	if currentJava != nil {
		currentCard := m.styles.Box.Render(
			m.styles.Current.Render("📍 現在アクティブ") + "\n" +
				fmt.Sprintf("🏷️  %s", currentJava.Version) + "\n" +
				m.styles.Muted.Render(fmt.Sprintf("📁 %s", currentJava.Home)),
		)
		s += currentCard + "\n"
	}

	// フィルター状況
	if m.filterMode {
		filterBox := m.styles.Card.Render(
			m.styles.Accent.Render("🔍 フィルター: ") +
				m.styles.Selected.Render(m.filterText+"█") +
				m.styles.Muted.Render(" (Escでキャンセル)"),
		)
		s += filterBox + "\n"
	} else if m.filterText != "" {
		filterBox := m.styles.Box.Render(
			m.styles.Accent.Render("🔍 フィルター: ") +
				m.filterText +
				m.styles.Muted.Render(fmt.Sprintf(" (%d件)", len(m.filteredInstalls))),
		)
		s += filterBox + "\n"
	}

	// Java一覧
	if m.loading {
		s += m.spinner.View() + " " + m.styles.Muted.Render("Javaインストールを検出中...") + "\n"
		return s
	}

	if len(m.javaInstalls) == 0 {
		s += m.styles.Error.Render("❌ Javaインストールが見つかりません") + "\n"
		return s
	}

	// 選択可能なJava一覧
	activeList := m.javaInstalls
	if m.filterText != "" {
		activeList = m.filteredInstalls
	}

	if len(activeList) == 0 && m.filterText != "" {
		s += m.styles.Error.Render("❌ フィルターに一致するJavaが見つかりません") + "\n"
	} else {
		listHeader := "🔄 利用可能なJava:"
		if len(activeList) != len(m.javaInstalls) {
			listHeader += fmt.Sprintf(" (%d/%d)", len(activeList), len(m.javaInstalls))
		}
		s += m.styles.Header.Render(listHeader) + "\n"

		for i, install := range activeList {
			cursor := " "
			if m.cursor == i {
				cursor = "▶"
			}

			// バージョンにバッジを追加
			version := install.Version
			if install.Current {
				version += " " + m.styles.Badge.Render("[ACTIVE]")
			}

			line := fmt.Sprintf("%s %s", cursor, version)
			shortPath := m.shortenPath(install.Home)
			line += m.styles.Muted.Render(fmt.Sprintf(" 📁 %s", shortPath))

			if m.cursor == i {
				s += m.styles.Selected.Render(line) + "\n"
			} else {
				s += m.styles.Normal.Render(line) + "\n"
			}
		}
	}

	// プレビュー（コンパクト版）
	if m.showPreview && len(activeList) > 0 && m.cursor < len(activeList) {
		selected := activeList[m.cursor]
		preview := fmt.Sprintf("🔍 %s | 📁 %s", selected.Version, m.shortenPath(selected.Home))
		s += "\n" + m.styles.Box.Render(preview) + "\n"
	}

	// コントロール
	s += "\n"
	if m.filterMode {
		s += m.styles.Border.Render("💡 文字入力でフィルター | Enter:確定 | Esc:キャンセル")
	} else {
		s += m.styles.Border.Render("💡 ↑↓:移動 | Enter:選択 | /:検索 | c:詳細 | i:情報 | q:終了")
	}

	return s
}

// 詳細モード表示（従来版ベース）
func (m Model) renderDetailedView() string {
	var s string

	// タイトル
	s += m.styles.Title.Render("🔄 JavaSwitcher v0.2.1") + "\n"
	s += m.styles.Header.Render("高速Java環境切り替えツール") + "\n"

	// クレジット情報（iキーで切り替え）
	if m.showCredits {
		creditBox := m.styles.Box.Render(
			"📦 バージョン: v0.2.1\n" +
				"🎯 作者: s12kuma01\n" +
				"📚 ライセンス: OSL-3.0\n" +
				"🌐 GitHub: https://github.com/Sumire-Labs/JSwitcher",
		)
		s += creditBox + "\n"
	}

	// エラー・ローディング
	if m.err != nil {
		s += m.styles.Error.Render(fmt.Sprintf("❌ エラー: %v", m.err)) + "\n"
		return s
	}

	if m.loading {
		s += m.spinner.View() + " " + m.styles.Muted.Render("Javaインストールを検出中...") + "\n"
		return s
	}

	// 残りの詳細表示ロジック...
	// （省略して後で実装）

	return s + m.styles.Border.Render("💡 c:コンパクト | その他の操作...")
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
