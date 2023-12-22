package tui

import (
	"fmt"

	"ewintr.nl/emdb/movie"
)

type Movie struct {
	m movie.Movie
}

func (m Movie) FilterValue() string {
	return m.m.Title
}

func (m Movie) Title() string {
	return m.m.Title
}

func (m Movie) Description() string {
	return fmt.Sprintf("description: %s", m.m.Title)
}
