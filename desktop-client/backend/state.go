package backend

import (
	"go-mod.ewintr.nl/emdb/storage"
)

type State struct {
	Watched []storage.Movie
	Log     []string
}

func NewState() *State {
	return &State{
		Watched: make([]storage.Movie, 0),
		Log:     make([]string, 0),
	}
}
