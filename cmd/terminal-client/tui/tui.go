package tui

import (
	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	docStyle                 = lipgloss.NewStyle().Padding(1)
	colorNormalForeground    = lipgloss.ANSIColor(termenv.ANSIWhite)
	colorHighLightForeGround = lipgloss.ANSIColor(termenv.ANSIBrightWhite)
	windowStyle              = lipgloss.NewStyle().
					BorderForeground(colorHighLightForeGround).
					Foreground(colorNormalForeground).
					Padding(0, 1).
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

type TabSizeMsgType tea.WindowSizeMsg

func New(emdb *client.EMDB, tmdb *client.TMDB, logger *Logger) (*tea.Program, error) {
	logViewport := viewport.New(0, 0)
	logViewport.KeyMap = viewport.KeyMap{}

	m, _ := NewBaseModel(emdb, tmdb, logger)
	p := tea.NewProgram(m, tea.WithAltScreen())

	return p, nil
}
