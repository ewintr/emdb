package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type tabEMDB struct {
	initialized bool
	list        list.Model
	emdb        *client.EMDB
	logger      *Logger
}

func NewTabEMDB(emdb *client.EMDB, logger *Logger) (tea.Model, tea.Cmd) {
	list := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Movies"
	list.SetShowHelp(false)

	m := tabEMDB{
		emdb:   emdb,
		logger: logger,
		list:   list,
	}

	logger.Log("search emdb...")
	return m, FetchMovieList(emdb)
}

func (m tabEMDB) Init() tea.Cmd {
	return nil
}

func (m tabEMDB) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TabSizeMsg:
		if !m.initialized {
			m.initialized = true
		}
		m.list.SetSize(msg.Width, msg.Height)
	case Movies:
		m.logger.Log(fmt.Sprintf("found %d movies in in emdb", len(msg)))
		m.list.SetItems(msg.listItems())
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m tabEMDB) View() string {
	return m.list.View()
}

func (m *tabEMDB) Log(s string) {
	m.logger.Log(s)
}

func FetchMovieList(emdb *client.EMDB) tea.Cmd {
	return func() tea.Msg {
		ems, err := emdb.GetMovies()
		if err != nil {
			return err
		}
		return Movies(ems)
	}
}
