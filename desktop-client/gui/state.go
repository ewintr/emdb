package gui

import "fyne.io/fyne/v2/data/binding"

type State struct {
	Watched binding.StringList
	Log     binding.String
}

func NewState() *State {
	watched := binding.BindStringList(
		&[]string{"Item 1", "Item 2", "Item 3"},
	)

	return &State{
		Watched: watched,
		Log:     binding.NewString(),
	}
}
