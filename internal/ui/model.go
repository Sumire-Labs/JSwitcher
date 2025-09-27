package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"javaswitcher/internal/java"
)

type Model struct {
	javaInstalls []java.Installation
	cursor       int
	selected     map[int]struct{}
	loading      bool
	loadingDone  bool
	err          error
	detector     *java.Detector
	switcher     *java.Switcher
	styles       Styles
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
		javaInstalls: []java.Installation{},
		selected:     make(map[int]struct{}),
		loading:      true,
		loadingDone:  false,
		detector:     java.NewDetector(),
		switcher:     java.NewSwitcher(),
		styles:       NewStyles(),
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.javaInstalls)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.javaInstalls) > 0 {
				selected := m.javaInstalls[m.cursor]
				return m, m.setJavaHome(selected.Home)
			}
		case "r", "R":
			// 前のJAVA_HOMEに復元
			return m, m.restorePreviousJavaHome()
		}

	// Mouse events disabled to prevent interference with CMD scrolling

	case JavaDetectedMsg:
		m.javaInstalls = msg.Installations
		m.loading = false
		m.loadingDone = true

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
	s += m.styles.Normal.Render("📦 バージョン: v0.1.0")
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

	// Show Java installations (appears below the credit section)
	s += "\n\n"
	s += m.styles.Header.Render("📋 利用可能なJavaインストール:")

	for i, installation := range m.javaInstalls {
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

	s += "\n"
	s += m.styles.Header.Render("💡 ↑/↓キーで移動、Enterで選択、'r'で復元、'q'で終了")

	return s
}