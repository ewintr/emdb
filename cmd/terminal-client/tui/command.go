package tui

import (
	"fmt"

	"ewintr.nl/emdb/client"
	tea "github.com/charmbracelet/bubbletea"
)

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
