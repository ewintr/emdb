package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type FetchMoviesCmd tea.Cmd

func FetchMovieList() tea.Cmd {
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
