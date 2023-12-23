package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type modelSearch struct {
	tmdb          *client.TMDB
	focused       string
	searchInput   textinput.Model
	searchResults list.Model
	logContent    string
	ready         bool
	logViewport   viewport.Model
}

func (m modelSearch) Init() tea.Cmd {
	return nil
}

func (m modelSearch) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			switch m.focused {
			case "search":
				m.Search()
			case "result":
				selected := m.searchResults.Items()[m.searchResults.Index()].(Movie)
				m.Log(fmt.Sprintf("selected: %d - %s", m.searchResults.Index(), selected.Title()))
				return NewEdit(selected), nil
			}
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.initialModel(msg.Width, msg.Height)
		}
	}

	switch m.focused {
	case "search":
		m.searchInput, cmd = m.searchInput.Update(msg)
	case "result":
		m.searchResults, cmd = m.searchResults.Update(msg)
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *modelSearch) Log(msg string) {
	m.logContent = fmt.Sprintf("%s\n%s", m.logContent, msg)
	m.logViewport.SetContent(m.logContent)
	m.logViewport.GotoBottom()
}

func (m *modelSearch) Search() {
	m.Log("start search")
	movies, err := m.tmdb.Search(m.searchInput.Value())
	if err != nil {
		m.Log(fmt.Sprintf("error: %v", err))
		return
	}

	m.Log(fmt.Sprintf("found %d results", len(movies)))
	items := []list.Item{}
	for _, res := range movies {
		items = append(items, Movie{m: res})
	}

	m.searchResults.SetItems(items)
	m.focused = "result"
}

func (m modelSearch) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n---\n%s\n---\n%s", m.searchInput.View(), m.searchResults.View(), m.logViewport.View())
}

func (m *modelSearch) initialModel(width, height int) {

	si := textinput.New()
	si.Placeholder = "title"
	si.CharLimit = 156
	si.Width = 20
	m.searchInput = si
	m.searchInput.Focus()

	m.searchResults = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height-50)
	m.searchResults.Title = "Search results"
	m.searchResults.SetShowHelp(false)

	m.logViewport = viewport.New(width, 10)
	m.logViewport.SetContent(m.logContent)
	m.logViewport.KeyMap = viewport.KeyMap{}
	m.focused = "search"
	m.ready = true
}
