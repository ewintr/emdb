package backend

import (
	"fmt"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/storage"
)

type Backend struct {
	s         *State
	in        chan Command
	out       chan State
	logLines  []string
	movieRepo *storage.MovieRepository
	tmdb      *client.TMDB
}

func NewBackend(movieRepo *storage.MovieRepository, tmdb *client.TMDB) *Backend {
	b := &Backend{
		s:         NewState(),
		in:        make(chan Command),
		out:       make(chan State),
		logLines:  make([]string, 0),
		movieRepo: movieRepo,
		tmdb:      tmdb,
	}
	go b.Run()

	b.in <- Command{Name: CommandRefreshWatched}

	return b
}

func (b *Backend) Out() chan State {
	return b.out
}

func (b *Backend) In() chan Command {
	return b.in
}

func (b *Backend) Run() {
	for cmd := range b.in {
		switch cmd.Name {
		case CommandAdd:
			newName, ok := cmd.Args[ArgName]
			if ok {
				newNameStr, ok := newName.(string)
				if ok {
					//b.s.Watched.Append(newNameStr)
					b.Log("Item added: " + newNameStr)
				}
			}
		case CommandRefreshWatched:
			b.RefreshWatched()
		default:
			b.Error(fmt.Errorf("unknown command: %s", cmd.Name))
		}

		b.out <- *b.s
	}
}

func (b *Backend) RefreshWatched() {
	watched, err := b.movieRepo.FindAll()
	if err != nil {
		b.Error(fmt.Errorf("could not refresh watched: %w", err))
	}
	b.s.Watched = watched
}

func (b *Backend) Error(err error) {
	b.Log(fmt.Sprintf("ERROR: %s", err))
}

func (b *Backend) Log(msg string) {
	b.logLines = append(b.logLines, msg)
	b.s.Log = b.logLines
}
