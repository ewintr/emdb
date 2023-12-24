package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type tabTMDB struct {
	initialized   bool
	focused       string
	searchInput   textinput.Model
	searchResults list.Model
	tmdb          *client.TMDB
	logger        *Logger
}

func NewTabTMDB(tmdb *client.TMDB, logger *Logger) (tea.Model, tea.Cmd) {
	m := tabTMDB{
		tmdb:   tmdb,
		logger: logger,
	}

	return m, nil
}

func (m tabTMDB) Init() tea.Cmd {
	return nil
}

func (m tabTMDB) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TabSizeMsgType:
		if !m.initialized {
			m.initialModel(msg.Width, msg.Height)
		}
		m.initialized = true
		m.searchResults.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.focused {
			case "search":
				cmds = append(cmds, SearchTMDB(m.tmdb, m.searchInput.Value()))
				m.searchInput.Blur()
				m.Log("search tmdb...")
			}
		}
	case Movies:
		m.Log(fmt.Sprintf("found %d movies in in tmdb", len(msg)))
		m.searchResults.SetItems(msg.listItems())
		m.focused = "result"
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

func (m tabTMDB) View() string {
	return fmt.Sprintf("%s\n%s\n", m.searchInput.View(), m.searchResults.View())
}

func (m *tabTMDB) Log(s string) {
	m.logger.Log(s)
}

func (m *tabTMDB) initialModel(width, height int) {
	si := textinput.New()
	si.Placeholder = "title"
	si.CharLimit = 156
	si.Width = 20
	m.searchInput = si
	m.searchInput.Focus()

	m.searchResults = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height-50)
	m.searchResults.Title = "Search results"
	m.searchResults.SetShowHelp(false)

	m.focused = "search"
}

func SearchTMDB(tmdb *client.TMDB, query string) tea.Cmd {
	return func() tea.Msg {
		tms, err := tmdb.Search(query)
		if err != nil {
			return err
		}
		return Movies(tms)
	}
}
