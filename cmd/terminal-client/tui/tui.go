package tui

import (
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

type LogMsg tea.Msg

func (l *Logger) SetProgram(p *tea.Program) {
	l.p = p
}

func (l *Logger) Log(s string) {
	l.Lines = append(l.Lines, s)
}

type TabSizeMsgType tea.WindowSizeMsg

func New(conf Config, logger *Logger) (*tea.Program, error) {
	tabs := []string{"Erik's movie database", "The movie database"}
	tmdb, err := client.NewTMDB(conf.TMDBAPIKey)
	if err != nil {
		return nil, err
	}
	m := baseModel{
		config: conf,
		emdb:   client.NewEMDB(conf.EMDBBaseURL, conf.EMDBAPIKey),
		tmdb:   tmdb,
		Tabs:   tabs,
		logger: logger,
	}
	m.TabContent = NewEMDBTab(&m, m.logger)

	p := tea.NewProgram(m, tea.WithAltScreen())

	return p, nil
}
