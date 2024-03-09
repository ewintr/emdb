package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TabSizeMsg tea.WindowSizeMsg
type TabResetMsg string

type TabSet struct {
	active int
	order  []string
	title  map[string]string
	tabs   map[string]tea.Model
	size   TabSizeMsg
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
	case TabSizeMsg:
		for _, name := range t.order {
			t.tabs[name], cmd = t.tabs[name].Update(msg)
			cmds = append(cmds, cmd)
		}
		t.size = msg.(TabSizeMsg)
	case TabResetMsg:
		name := string(msg.(TabResetMsg))
		t.tabs[name], cmd = t.tabs[name].Update(msg)
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
	menu = lipgloss.PlaceHorizontal(t.size.Width, lipgloss.Left, menu)

	pane := lipgloss.NewStyle().
		//	Background(lipgloss.ANSIColor(termenv.ANSIBlue)).
		Render(t.tabs[t.order[t.active]].View())
	pane = lipgloss.PlaceVertical(t.size.Height, lipgloss.Top, pane)

	return fmt.Sprintf("%s\n%s", menu, pane)
}
