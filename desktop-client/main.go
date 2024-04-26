package main

import (
	"code.ewintr.nl/emdb/desktop-client/backend"
	"code.ewintr.nl/emdb/desktop-client/gui"
)

func main() {
	b := backend.NewBackend()
	g := gui.New(b.In(), b.Out())
	g.Run()
}
