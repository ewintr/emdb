package backend

import "strings"

type Backend struct {
	s   *State
	c   chan Command
	log []string
}

func NewBackend() *Backend {
	b := &Backend{
		s:   NewState(),
		c:   make(chan Command),
		log: make([]string, 0),
	}
	go b.Run()

	return b
}

func (b *Backend) Out() *State {
	return b.s
}

func (b *Backend) In() chan Command {
	return b.c
}

func (b *Backend) Run() {
	for cmd := range b.c {
		switch cmd.Name {
		case CommandAdd:
			newName, ok := cmd.Args[ArgName]
			if ok {
				newNameStr, ok := newName.(string)
				if ok {
					b.s.Watched.Append(newNameStr)
					b.Log("Item added: " + newNameStr)
				}
			}
		}
	}
}

func (b *Backend) Log(msg string) {
	b.log = append(b.log, msg)
	b.s.Log.Set(strings.Join(b.log, "\n"))
}
