package tui

import (
	"fmt"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/charmbracelet/bubbles/list"
)

type Movie struct {
	m moviestore.Movie
}

func (m Movie) FilterValue() string {
	return m.m.Title
}

func (m Movie) Title() string {
	return fmt.Sprintf("%s (%d)", m.m.Title, m.m.Year)
}

func (m Movie) Description() string {
	return fmt.Sprintf("%s", m.m.Summary)
}

type Movies []moviestore.Movie

func (ms Movies) listItems() []list.Item {
	items := []list.Item{}
	for _, m := range ms {
		items = append(items, Movie{m: m})
	}
	return items
}
