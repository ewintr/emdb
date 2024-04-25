package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	a fyne.App
	w fyne.Window
}

func New() *GUI {
	a := app.New()
	w := a.NewWindow("EMDB")

	w.SetContent(Layout())
	w.Resize(fyne.NewSize(800, 600))

	return &GUI{
		a: a,
		w: w,
	}
}

func (g *GUI) Run() {
	g.w.ShowAndRun()
}

func Layout() fyne.CanvasObject {
	label1 := widget.NewLabel("Label 1")
	label2 := widget.NewLabel("Label 2")
	value2 := widget.NewLabel("Something")

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	form := container.New(layout.NewFormLayout(), label1, input, label2, value2)

	button := widget.NewButton("Save", func() {
		log.Println("Content was:", input.Text)
	})

	grid := container.NewBorder(nil, button, nil, nil, form)

	logLines := container.NewVScroll(widget.NewLabel("Log\n\n\n\n\n\n\n\nhoi\n\n\n\n\n\na lot"))
	//logLines.ScrollToBottom()

	tabs := container.NewAppTabs(
		container.NewTabItem("Watched", widget.NewLabel("Watched")),
		container.NewTabItem("TheMovieDB", grid),
		container.NewTabItem("Log", logLines),
	)

	return tabs
}
