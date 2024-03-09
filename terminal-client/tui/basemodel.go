package tui

import (
	"fmt"
	"strings"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/job"
	"code.ewintr.nl/emdb/storage"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type baseModel struct {
	movieRepo   *storage.MovieRepositoryPG
	reviewRepo  *storage.ReviewRepositoryPG
	jobQueue    *job.JobQueue
	tmdb        *client.TMDB
	tabs        *TabSet
	initialized bool
	logger      *Logger
	logViewport viewport.Model
	windowSize  tea.WindowSizeMsg
	contentSize tea.WindowSizeMsg
}

func NewBaseModel(movieRepo *storage.MovieRepositoryPG, reviewRepo *storage.ReviewRepositoryPG, jobQueue *job.JobQueue, tmdb *client.TMDB, logger *Logger) (tea.Model, tea.Cmd) {
	logViewport := viewport.New(0, 0)
	logViewport.KeyMap = viewport.KeyMap{}

	m := baseModel{
		movieRepo:   movieRepo,
		reviewRepo:  reviewRepo,
		jobQueue:    jobQueue,
		tmdb:        tmdb,
		tabs:        NewTabSet(),
		logViewport: logViewport,
		logger:      logger,
	}
	m.setSize()

	return m, nil
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case NextTabSelected:
		m.tabs.Next()
	case PrevTabSelected:
		m.tabs.Previous()
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.initialized {
			var emdbTab, tmdbTab tea.Model
			emdbTab, cmd = NewTabEMDB(m.movieRepo, m.logger)
			cmds = append(cmds, cmd)
			tmdbTab, cmd = NewTabTMDB(m.movieRepo, m.jobQueue, m.tmdb, m.logger)
			cmds = append(cmds, cmd)
			reviewTab, cmd := NewTabReview(m.reviewRepo, m.logger)
			cmds = append(cmds, cmd)
			m.tabs.AddTab("emdb", "Watched movies", emdbTab)
			m.tabs.AddTab("review", "Review", reviewTab)
			m.tabs.AddTab("tmdb", "TMDB", tmdbTab)
			m.initialized = true
		}
		m.Log(fmt.Sprintf("new window size: %dx%d", msg.Width, msg.Height))
		m.setSize()
		tabSize := TabSizeMsg{
			Width:  m.contentSize.Width,
			Height: m.contentSize.Height,
		}
		cmds = append(cmds, m.tabs.Update(tabSize))
	case NewMovie:
		m.Log(fmt.Sprintf("imported movie %s", msg.m.Title))
		m.tabs.Select("emdb")
		cmds = append(cmds, FetchMovieList(m.movieRepo))
	case error:
		m.Log(fmt.Sprintf("ERROR: %s", msg.Error()))
	default:
		cmds = append(cmds, m.tabs.Update(msg))
	}

	m.logViewport.SetContent(strings.Join(m.logger.Lines, "\n"))
	m.logViewport.GotoBottom()
	m.logViewport, cmd = m.logViewport.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m *baseModel) Log(msg string) {
	m.logger.Log(msg)
}

func (m baseModel) View() string {
	if !m.initialized {
		return "\n  Initializing..."
	}

	logWindow := windowStyle.
		Width(m.contentSize.Width).
		Height(logLineCount).
		//Background(lipgloss.ANSIColor(termenv.ANSIYellow)).
		Render(m.logViewport.View())

	return fmt.Sprintf("%s\n%s", m.tabs.View(), logWindow)
}

func (m *baseModel) setSize() {
	logHeight := logLineCount
	menuHeight := 1

	m.contentSize.Width = m.windowSize.Width - windowStyle.GetHorizontalFrameSize()
	m.contentSize.Height = m.windowSize.Height - windowStyle.GetVerticalFrameSize() - logHeight - menuHeight
	//m.Log(fmt.Sprintf("contentheight: %d = windowheight %d - windowframeheight %d -  logheight %d - menuheight %d", m.contentSize.Height, m.windowSize.Height, windowStyle.GetVerticalFrameSize(), logHeight, menuHeight))

	m.logViewport.Width = m.contentSize.Width
	m.logViewport.Height = logLineCount
}
