package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type emdbTab struct {
	ready  bool
	list   list.Model
	parent *baseModel
	logger *Logger
}

func NewEMDBTab(parent *baseModel, logger *Logger) tea.Model {
	m := emdbTab{}
	m.parent = parent
	m.logger = logger

	return m
}

func (m emdbTab) Init() tea.Cmd {
	return nil
}

func (m emdbTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TabSizeMsgType:
		if !m.ready {
			m.initialModel()
		}
		m.list.SetSize(msg.Width, msg.Height)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m emdbTab) View() string {
	return m.list.View()
}

func (m *emdbTab) initialModel() {
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = "Movies"
	m.list.SetShowHelp(false)
	m.refreshMovieList()
	m.ready = true
}

func (m *emdbTab) Log(s string) {
	m.logger.Log(s)
}

func (m *emdbTab) refreshMovieList() {
	m.Log("fetch emdb movies...")
	ems, err := m.parent.emdb.GetMovies()
	if err != nil {
		m.Log(err.Error())
	}
	items := make([]list.Item, len(ems))
	for i, em := range ems {
		items[i] = list.Item(Movie{m: em})
	}
	m.list.SetItems(items)
	m.Log(fmt.Sprintf("found %d movies in in emdb", len(items)))

}
