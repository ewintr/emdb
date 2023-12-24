package tui

import (
	"fmt"
	"os"
	"strings"

	"ewintr.nl/emdb/client"
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
	logLineCount = 5

	emdb   *client.EMDB
	tmdb   *client.TMDB
	logger = NewLogger()
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

func (l *Logger) SetProgram(p *tea.Program) {
	l.p = p
}

func (l *Logger) Log(s string) {
	l.Lines = append(l.Lines, s)
}

func (l *Logger) Content() string {
	if l.Lines == nil {
		return "logger not initialized"
	}

	return strings.Join(l.Lines, "\n")
}

type TabSizeMsgType tea.WindowSizeMsg

func New(emdb *client.EMDB, tmdb *client.TMDB) (*tea.Program, error) {
	emdb = emdb
	tmdb = tmdb

	fmt.Printf("emdb: %v\n", emdb)
	os.Exit(0)

	m, _ := NewBaseModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	logger.SetProgram(p)

	return p, nil
}
