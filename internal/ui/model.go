package ui

import (
	"fmt"
	"strings"
	"time"

	"javaswitcher/internal/java"

	tea "github.com/charmbracelet/bubbletea"
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
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.detectJavaInstallations,
		tea.EnterAltScreen,
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	return m, nil
}

func (m Model) View() string {
	s := m.styles.Title.Render("🔄 JavaSwitcher")
	s += "\n"
	s += m.styles.Header.Render("Windows向け高速Java環境切り替えツール")

	// Always show the loading/credit section
	s += "\n\n"

	loadingText := "🔍 Javaインストールを検出中..."
	if m.loadingDone {
		loadingText = "✅ 検出完了"
	}
	s += m.styles.Normal.Render(loadingText)
	s += "\n\n"

	// Credit information (always shown)
	s += m.styles.Header.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	s += "\n\n"
	s += m.styles.Normal.Render("📦 バージョン: v0.2.0")
	s += "\n"
	s += m.styles.Normal.Render("🎯 作者: s12kuma01")
	s += "\n"
	s += m.styles.Normal.Render("📚 ライセンス: OSL-3.0")
	s += "\n"
	s += m.styles.Normal.Render("🌐 GitHub: https://github.com/Sumire-Labs/JSwitcher")
	s += "\n\n"
	s += m.styles.Header.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// If still loading, return here
	if m.loading {
		return s
	}

	// Show error if any
	if m.err != nil {
		s += "\n\n"
		s += m.styles.Error.Render(fmt.Sprintf("❌ エラー: %v", m.err))
		return s
	}

	// Show no installations found
	if len(m.javaInstalls) == 0 {
		s += "\n\n"
		s += m.styles.Normal.Render("❌ Javaインストールが見つかりません")
		return s
	}

	// Show filter status
	if m.filterMode {
		s += "\n\n"
		s += m.styles.Header.Render("🔍 フィルター: ")
		s += m.styles.Selected.Render(m.filterText + "_")
		s += m.styles.Normal.Render(" (Escでキャンセル)")
	} else if m.filterText != "" {
		s += "\n\n"
		s += m.styles.Header.Render("🔍 フィルター: ")
		s += m.styles.Normal.Render(m.filterText)
		s += m.styles.Normal.Render(fmt.Sprintf(" (%d件表示)", len(m.filteredInstalls)))
	}

	// Show history if available
	if len(m.history) > 0 && !m.filterMode {
		s += "\n\n"
		s += m.styles.Header.Render("📜 最近使用したJava:")
		s += "\n"
		for i, item := range m.history {
			if i >= 3 { // Show only first 3
				break
			}
			s += m.styles.Normal.Render(fmt.Sprintf("   %d. %s", i+1, item))
			s += "\n"
		}
	}

	// Determine which list to show
	activeList := m.javaInstalls
	if m.filterText != "" {
		activeList = m.filteredInstalls
	}

	// Show Java installations
	s += "\n\n"
	s += m.styles.Header.Render("📋 利用可能なJavaインストール:")

	if len(activeList) == 0 && m.filterText != "" {
		s += "\n\n"
		s += m.styles.Normal.Render("❌ フィルターに一致するJavaが見つかりません")
	} else {
		for i, installation := range activeList {
			cursor := " "
			if m.cursor == i {
				cursor = "▶"
			}

			status := ""
			if installation.Current {
				status = m.styles.Current.Render(" (現在)")
			}

			line := fmt.Sprintf("%s %s%s", cursor, installation.Version, status)
			line += fmt.Sprintf("\n   📁 %s", installation.Home)

			if m.cursor == i {
				s += "\n" + m.styles.Selected.Render(line)
			} else {
				s += "\n" + m.styles.Normal.Render(line)
			}
			s += "\n"
		}
	}

	// Show preview if enabled and item selected
	if m.showPreview && len(activeList) > 0 && m.cursor < len(activeList) {
		selected := activeList[m.cursor]
		s += "\n"
		s += m.styles.Header.Render("🔍 プレビュー:")
		s += "\n"
		s += m.styles.Normal.Render(fmt.Sprintf("   バージョン: %s", selected.Version))
		s += "\n"
		s += m.styles.Normal.Render(fmt.Sprintf("   パス: %s", selected.Path))
		s += "\n"
		s += m.styles.Normal.Render(fmt.Sprintf("   ホーム: %s", selected.Home))
		if selected.Current {
			s += "\n"
			s += m.styles.Current.Render("   ✅ 現在アクティブ")
		}
	}

	// Show controls
	s += "\n\n"
	if m.filterMode {
		s += m.styles.Header.Render("💡 文字を入力してフィルター、Enterで確定、Escでキャンセル")
	} else {
		s += m.styles.Header.Render("💡 ↑/↓:移動 Enter:選択 r:復元 /:フィルター p:プレビュー q:終了")
	}

	return s
}
