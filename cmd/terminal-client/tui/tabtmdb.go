package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type tabTMDB struct {
	emdb          *client.EMDB
	tmdb          *client.TMDB
	initialized   bool
	focused       string
	searchInput   textinput.Model
	searchResults list.Model
	logger        *Logger
}

func NewTabTMDB(emdb *client.EMDB, tmdb *client.TMDB, logger *Logger) (tea.Model, tea.Cmd) {
	m := tabTMDB{
		emdb:   emdb,
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
	case TabSizeMsg:
		if !m.initialized {
			m.initialModel(msg.Width, msg.Height)
		}
		m.initialized = true
		m.searchResults.SetSize(msg.Width, msg.Height)
	case TabResetMsg:
		m.searchInput.SetValue("")
		m.searchResults.SetItems([]list.Item{})
		m.searchInput.Focus()
		m.focused = "search"
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.focused {
			case "search":
				cmds = append(cmds, m.SearchTMDBCmd(m.searchInput.Value()))
				m.searchInput.Blur()
				m.Log("search tmdb...")
			case "result":
				movie := m.searchResults.SelectedItem().(Movie)
				cmds = append(cmds, m.ImportMovieCmd(movie), m.ResetCmd())
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

func (m *tabTMDB) SearchTMDBCmd(query string) tea.Cmd {
	return func() tea.Msg {
		tms, err := m.tmdb.Search(query)
		if err != nil {
			return err
		}
		return Movies(tms)
	}
}

func (m *tabTMDB) ImportMovieCmd(movie Movie) tea.Cmd {
	return func() tea.Msg {
		newMovie, err := m.emdb.AddMovie(movie.m)
		if err != nil {
			return err
		}

		return NewMovie(Movie{m: newMovie})
	}
}

func (m *tabTMDB) ResetCmd() tea.Cmd {
	return func() tea.Msg {
		return TabResetMsg("tmdb")
	}
}
