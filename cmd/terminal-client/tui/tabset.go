package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TabSet struct {
	active int
	order  []string
	title  map[string]string
	tabs   map[string]tea.Model
	size   TabSizeMsgType
}

func NewTabSet() *TabSet {
	return &TabSet{
		order: make([]string, 0),
		title: make(map[string]string),
		tabs:  make(map[string]tea.Model),
	}
}

func (t *TabSet) AddTab(name, title string, model tea.Model) {
	t.order = append(t.order, name)
	t.title[name] = title
	t.tabs[name] = model
}

func (t *TabSet) Next() {
	t.active++
	if t.active > len(t.order)-1 {
		t.active = 0
	}
}

func (t *TabSet) Previous() {
	t.active--
	if t.active < 0 {
		t.active = len(t.order) - 1
	}
}

func (t *TabSet) Select(name string) {
	for i, n := range t.order {
		if n == name {
			t.active = i
			return
		}
	}
}

func (t *TabSet) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg.(type) {
	case TabSizeMsgType:
		for _, name := range t.order {
			t.tabs[name], cmd = t.tabs[name].Update(msg)
			cmds = append(cmds, cmd)
		}
		t.size = msg.(TabSizeMsgType)
	//case ImportMovieMsg:
	//	t.Select("emdb")
	//	t.tabs["emdb"], cmd = t.tabs["emdb"].Update(msg)
	default:
		name := t.order[t.active]
		t.tabs[name], cmd = t.tabs[name].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (t *TabSet) View() string {
	var items []string
	for i, name := range t.order {
		if i == t.active {
			items = append(items, lipgloss.NewStyle().
				Foreground(colorHighLightForeGround).
				Render(fmt.Sprintf(" * %s ", t.title[name])))
			continue
		}

		items = append(items, lipgloss.NewStyle().
			Foreground(colorNormalForeground).
			Render(fmt.Sprintf("   %s ", t.title[name])))
	}
	menu := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	pane := t.tabs[t.order[t.active]].View()
	lipgloss.PlaceHorizontal(t.size.Width, lipgloss.Left, menu)

	return fmt.Sprintf("%s\n%s", menu, pane)
}
