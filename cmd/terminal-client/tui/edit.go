package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type modelEdit struct {
	movie       Movie
	focused     string
	ratingField textinput.Model
	commentFiel textinput.Model
}

func NewEdit(movie Movie) *modelEdit {
	m := &modelEdit{
		movie:       movie,
		focused:     "rating",
		ratingField: textinput.New(),
		commentFiel: textinput.New(),
	}
	m.ratingField.Placeholder = "Rating"
	m.ratingField.Width = 2
	m.ratingField.CharLimit = 2
	m.ratingField.Focus()

	return m
}

func (m modelEdit) Init() tea.Cmd {
	return nil
}

func (m modelEdit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab":
			switch m.focused {
			case "rating":
				m.focused = "comment"
				m.commentFiel.Focus()
			case "comment":
				m.focused = "rating"
				m.ratingField.Focus()
			}
		case "enter":
		}
	}

	switch m.focused {
	case "rating":
		m.ratingField, cmd = m.ratingField.Update(msg)
	case "comment":
		m.commentFiel, cmd = m.commentFiel.Update(msg)
	}

	return m, cmd
}

func (m modelEdit) View() string {
	return fmt.Sprintf("Title: \t%s\nSumary: \t%s\nRating: \t%s\nComment: \t%s\n",
		m.movie.Title(),
		m.movie.Description(),
		m.ratingField.View(),
		m.commentFiel.View(),
	)
}
