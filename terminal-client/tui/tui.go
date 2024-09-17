package tui

import (
	"go-mod.ewintr.nl/emdb/client"
	"go-mod.ewintr.nl/emdb/job"
	"go-mod.ewintr.nl/emdb/storage"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	colorNormalForeground    = lipgloss.ANSIColor(termenv.ANSIWhite)
	colorHighLightForeGround = lipgloss.ANSIColor(termenv.ANSIBrightWhite)
	windowStyle              = lipgloss.NewStyle().
					BorderForeground(colorHighLightForeGround).
					Foreground(colorNormalForeground).
					Border(lipgloss.NormalBorder(), true)
	logLineCount = 10
)

type Logger struct {
	p     *tea.Program
	Lines []string
}

func NewLogger() *Logger {
	return &Logger{
		Lines: make([]string, 0),
	}
}

func (l *Logger) SetProgram(p *tea.Program) { l.p = p }
func (l *Logger) Log(s string)              { l.Lines = append(l.Lines, s) }

type NewMovie Movie

type NextTabSelected struct{}

func SelectNextTab() tea.Cmd {
	return func() tea.Msg {
		return NextTabSelected{}
	}
}

type PrevTabSelected struct{}

func SelectPrevTab() tea.Cmd {
	return func() tea.Msg {
		return PrevTabSelected{}
	}
}

func New(movieRepo *storage.MovieRepository, reviewRepo *storage.ReviewRepository, jobQueue *job.JobQueue, tmdb *client.TMDB, logger *Logger) (*tea.Program, error) {
	logViewport := viewport.New(0, 0)
	logViewport.KeyMap = viewport.KeyMap{}

	m, _ := NewBaseModel(movieRepo, reviewRepo, jobQueue, tmdb, logger)
	p := tea.NewProgram(m, tea.WithAltScreen())

	return p, nil
}
