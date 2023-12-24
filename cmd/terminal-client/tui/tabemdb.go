package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type tabEMDB struct {
	initialized bool
	list        list.Model
}

func NewTabEMDB() (tea.Model, tea.Cmd) {
	list := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Movies"
	list.SetShowHelp(false)

	m := tabEMDB{
		list: list,
	}

	return m, FetchMovieList()
}

func (m tabEMDB) Init() tea.Cmd {
	return nil
}

func (m tabEMDB) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m tabEMDB) View() string {
	return m.list.View()
}
