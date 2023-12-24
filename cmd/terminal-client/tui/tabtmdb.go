package tui

import (
	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

//focused       string
//searchInput   textinput.Model
//searchResults list.Model
//movieLis

type tabTMDB struct {
	initialized bool
	results     list.Model
	tmdb        *client.TMDB
	logger      *Logger
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	m.results, cmd = m.results.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m tabTMDB) View() string {
	return "tmdb"
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
