package tui

import (
	"ewintr.nl/emdb/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tabReview struct {
	initialized bool
	emdb        *client.EMDB
	width       int
	height      int
	logger      *Logger
}

func NewTabReview(emdb *client.EMDB, logger *Logger) (tea.Model, tea.Cmd) {
	return &tabReview{
		emdb:   emdb,
		logger: logger,
	}, nil
}

func (m *tabReview) Init() tea.Cmd {
	return nil
}

func (m *tabReview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case TabSizeMsg:
		if !m.initialized {
			m.initialized = true
		}
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "right", "tab":
			cmds = append(cmds, SelectNextTab())
		case "left", "shift+tab":
			cmds = append(cmds, SelectPrevTab())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *tabReview) View() string {
	return lipgloss.NewStyle().
		Width(m.width - 2).
		Height(m.height - 2).
		Padding(1).
		Render("Review")
}
