package gui

import (
	"strings"

	"code.ewintr.nl/emdb/desktop-client/backend"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	a   fyne.App
	w   fyne.Window
	in  chan backend.State
	out chan backend.Command
	s   *State
}

func New(out chan backend.Command, in chan backend.State) *GUI {
	a := app.New()
	w := a.NewWindow("EMDB")
	w.Resize(fyne.NewSize(800, 600))

	g := &GUI{
		a:   a,
		w:   w,
		in:  in,
		out: out,
		s:   NewState(),
	}

	g.SetContent()

	return g
}

func (g *GUI) Run() {
	go func() {
		for s := range g.in {
			g.Update(s)
		}
	}()
	g.w.ShowAndRun()
}

func (g *GUI) Update(bs backend.State) {
	// watched
	watched := make([]string, 0, len(bs.Watched))
	for _, m := range bs.Watched {
		watched = append(watched, m.Title)
	}
	g.s.Watched.Set(watched)

	// log
	g.s.Log.Set(strings.Join(bs.Log, "\n"))

}

func (g *GUI) SetContent() {
	label1 := widget.NewLabel("Label 1")
	label2 := widget.NewLabel("Label 2")
	value2 := widget.NewLabel("Something")

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	form := container.New(layout.NewFormLayout(), label1, input, label2, value2)

	button := widget.NewButton("Save", func() {
		g.out <- backend.Command{
			Name: backend.CommandAdd,
			Args: map[string]any{
				backend.ArgName: input.Text,
			},
		}
	})

	grid := container.NewBorder(nil, button, nil, nil, form)

	logLines := container.NewVScroll(widget.NewLabelWithData(g.s.Log))
	//logLines.ScrollToBottom()

	list := widget.NewListWithData(g.s.Watched,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	tabs := container.NewAppTabs(
		container.NewTabItem("Watched", list),
		container.NewTabItem("TheMovieDB", grid),
		container.NewTabItem("Log", logLines),
	)

	g.w.SetContent(tabs)
}
