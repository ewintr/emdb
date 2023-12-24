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
	m.Log(fmt.Sprintf("%v", msg))

	switch msg := msg.(type) {
	case TabSizeMsgType:
		if !m.initialized {
			m.Log(fmt.Sprintf("tmdb initialized. focused: %s", m.focused))
			m.initialModel(msg.Width, msg.Height)
		}
		m.initialized = true
		m.searchResults.SetSize(msg.Width, msg.Height-10)
	case tea.KeyMsg:
		switch msg.String() {
		}
	}

	m.Log(fmt.Sprintf("focused: %s", m.focused))
	switch m.focused {
	case "search":
		m.Log("search")
		m.searchInput, cmd = m.searchInput.Update(msg)
	case "result":
		m.Log("result")
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

//func (m *model) Search() {
//	m.Log("start search")
//	movies, err := m.tmdb.Search(m.searchInput.Value())
//	if err != nil {
//		m.Log(fmt.Sprintf("error: %v", err))
//		return
//	}
//
//	m.Log(fmt.Sprintf("found %d results", len(movies)))
//	items := []list.Item{}
//	for _, res := range movies {
//		items = append(items, Movie{m: res})
//	}
//
//	m.searchResults.SetItems(items)
//	m.focused = "result"
//}
