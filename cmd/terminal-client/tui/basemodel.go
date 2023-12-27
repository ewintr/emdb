package tui

import (
	"fmt"
	"strings"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type baseModel struct {
	emdb        *client.EMDB
	tmdb        *client.TMDB
	tabs        *TabSet
	initialized bool
	logger      *Logger
	logViewport viewport.Model
	windowSize  tea.WindowSizeMsg
	contentSize tea.WindowSizeMsg
}

func NewBaseModel(emdb *client.EMDB, tmdb *client.TMDB, logger *Logger) (tea.Model, tea.Cmd) {
	logViewport := viewport.New(0, 0)
	logViewport.KeyMap = viewport.KeyMap{}

	m := baseModel{
		emdb:        emdb,
		tmdb:        tmdb,
		tabs:        NewTabSet(),
		logViewport: logViewport,
		logger:      logger,
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
	case NextTabSelected:
		m.tabs.Next()
	case PrevTabSelected:
		m.tabs.Previous()
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.initialized {
			var emdbTab, tmdbTab tea.Model
			emdbTab, cmd = NewTabEMDB(m.emdb, m.logger)
			cmds = append(cmds, cmd)
			tmdbTab, cmd = NewTabTMDB(m.emdb, m.tmdb, m.logger)
			cmds = append(cmds, cmd)
			m.tabs.AddTab("emdb", "Watched movies", emdbTab)
			m.tabs.AddTab("tmdb", "TMDB", tmdbTab)
			m.initialized = true
		}
		m.Log(fmt.Sprintf("new window size: %dx%d", msg.Width, msg.Height))
		m.setSize()
		tabSize := TabSizeMsg{
			Width:  m.contentSize.Width,
			Height: m.contentSize.Height,
		}
		cmds = append(cmds, m.tabs.Update(tabSize))
	case NewMovie:
		m.Log(fmt.Sprintf("imported movie %s", msg.m.Title))
		m.tabs.Select("emdb")
		cmds = append(cmds, FetchMovieList(m.emdb))
	case error:
		m.Log(fmt.Sprintf("ERROR: %s", msg.Error()))
	default:
		cmds = append(cmds, m.tabs.Update(msg))
	}

	m.logViewport.SetContent(strings.Join(m.logger.Lines, "\n"))
	m.logViewport.GotoBottom()
	m.logViewport, cmd = m.logViewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m *baseModel) Log(msg string) {
	m.logger.Log(msg)
}

func (m baseModel) View() string {
	if !m.initialized {
		return "\n  Initializing..."
	}

	logWindow := windowStyle.Width(m.contentSize.Width).Height(logLineCount).Render(m.logViewport.View())
	return docStyle.Render(fmt.Sprintf("%s\n%s", m.tabs.View(), logWindow))
}

func (m *baseModel) setSize() {
	logHeight := logLineCount + docStyle.GetVerticalFrameSize()
	menuHeight := 1

	m.contentSize.Width = m.windowSize.Width - windowStyle.GetHorizontalFrameSize() - docStyle.GetHorizontalFrameSize()
	m.contentSize.Height = m.windowSize.Height - windowStyle.GetVerticalFrameSize() - docStyle.GetVerticalFrameSize() - logHeight - menuHeight

	m.logViewport.Width = m.contentSize.Width
	m.logViewport.Height = logLineCount
}
