package tui

import (
	"ewintr.nl/emdb/cmd/terminal-client/clients"
	tea "github.com/charmbracelet/bubbletea"
)

func New(tmdb *clients.TMDB) *tea.Program {
	m := modelSearch{
		tmdb: tmdb,
	}
	return tea.NewProgram(m, tea.WithAltScreen())
}
