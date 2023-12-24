package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type baseModel struct {
	Tabs        []string
	TabContent  tea.Model
	activeTab   int
	initialized bool
	logViewport viewport.Model
	windowSize  tea.WindowSizeMsg
	contentSize tea.WindowSizeMsg
	tabSize     tea.WindowSizeMsg
}

func NewBaseModel() (tea.Model, tea.Cmd) {
	logViewport := viewport.New(0, 0)
	logViewport.KeyMap = viewport.KeyMap{}

	m := baseModel{
		Tabs:        []string{"Erik's movie database", "The movie database"},
		logViewport: logViewport,
	}
	m.setSize()

	return m, nil
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "right", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.initialized {
			m.TabContent, cmd = NewTabEMDB()
			cmds = append(cmds, cmd)
			m.initialized = true
		}
		logger.Log(fmt.Sprintf("new window size: %dx%d", msg.Width, msg.Height))
		m.setSize()
		tabSize := TabSizeMsgType{
			Width:  m.contentSize.Width,
			Height: m.contentSize.Height,
		}
		m.TabContent, cmd = m.TabContent.Update(tabSize)
		cmds = append(cmds, cmd)
		logger.Log("done with resize")
		//case FetchMoviesCmd:
		//	cmds = append(cmds, FetchMovieList(m.emdb, m.logger))
	}

	m.TabContent, cmd = m.TabContent.Update(msg)
	cmds = append(cmds, cmd)

	m.logViewport.SetContent(logger.Content())
	m.logViewport.GotoBottom()
	m.logViewport, cmd = m.logViewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m baseModel) View() string {
	if !m.initialized {
		return "\n  Initializing..."
	}

	doc := strings.Builder{}
	doc.WriteString(m.renderMenu())
	doc.WriteString("\n")
	doc.WriteString(m.renderTabContent())
	doc.WriteString("\n")
	doc.WriteString(m.renderLog())
	return docStyle.Render(doc.String())
}

func (m *baseModel) renderMenu() string {
	var items []string
	for i, t := range m.Tabs {
		if i == m.activeTab {
			items = append(items, lipgloss.NewStyle().
				Foreground(colorHighLightForeGround).
				Render(fmt.Sprintf(" * %s ", t)))
			continue
		}

		items = append(items, lipgloss.NewStyle().
			Foreground(colorNormalForeground).
			Render(fmt.Sprintf("   %s ", t)))
	}

	return lipgloss.PlaceHorizontal(m.contentSize.Width, lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, items...))
}

func (m *baseModel) renderTabContent() string {
	content := m.TabContent.View()
	return windowStyle.Width(m.contentSize.Width).Height(m.contentSize.Height).Render(content)
}

func (m *baseModel) renderLog() string {
	return windowStyle.Width(m.contentSize.Width).Height(logLineCount).Render(m.logViewport.View())
}

func (m *baseModel) setSize() {
	logHeight := logLineCount + docStyle.GetVerticalFrameSize()
	menuHeight := 1

	m.contentSize.Width = m.windowSize.Width - windowStyle.GetHorizontalFrameSize() - docStyle.GetHorizontalFrameSize()
	m.contentSize.Height = m.windowSize.Height - windowStyle.GetVerticalFrameSize() - docStyle.GetVerticalFrameSize() - logHeight - menuHeight

	m.logViewport.Width = m.contentSize.Width
	m.logViewport.Height = logLineCount
}
