package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type emdbTab struct {
	initialized bool
	list        list.Model
	emdb        *client.EMDB
	logger      *Logger
}

func NewEMDBTab(emdb *client.EMDB, logger *Logger) (tea.Model, tea.Cmd) {
	list := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Movies"
	list.SetShowHelp(false)

	m := emdbTab{
		emdb:   emdb,
		logger: logger,
		list:   list,
	}

	return m, FetchMovieList(emdb, logger)
}

func (m emdbTab) Init() tea.Cmd {
	return nil
}

func (m emdbTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TabSizeMsgType:
		if !m.initialized {
			//cmds = append(cmds, FetchMovieList(m.emdb, m.logger))
			m.initialized = true
		}
		m.list.SetSize(msg.Width, msg.Height)
	case Movies:
		m.list.SetItems(msg.listItems())
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m emdbTab) View() string {
	return m.list.View()
}

func (m *emdbTab) Log(s string) {
	m.logger.Log(s)
}

func FetchMovieList(emdb *client.EMDB, logger *Logger) tea.Cmd {
	return func() tea.Msg {
		logger.Log("fetching emdb movies...")
		ems, err := emdb.GetMovies()
		if err != nil {
			logger.Log(err.Error())
		}

		//m.list.SetItems(items)
		logger.Log(fmt.Sprintf("found %d movies in in emdb", len(ems)))

		return Movies(ems)
	}
}
