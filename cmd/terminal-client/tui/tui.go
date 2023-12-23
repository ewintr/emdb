package tui

import (
	"fmt"
	"strings"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
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
	logLineCount = 5
)

func New(conf Config) (*tea.Program, error) {
	tabs := []string{"Erik's movie database", "The movie database"}
	tabContent := []string{"Emdb", "TMDB"}

	tmdb, err := client.NewTMDB(conf.TMDBAPIKey)
	if err != nil {
		return nil, err
	}
	m := baseModel{
		config:     conf,
		emdb:       client.NewEMDB(conf.EMDBBaseURL, conf.EMDBAPIKey),
		tmdb:       tmdb,
		Tabs:       tabs,
		TabContent: tabContent,
	}
	return tea.NewProgram(m, tea.WithAltScreen()), nil
}

type baseModel struct {
	config     Config
	emdb       *client.EMDB
	tmdb       *client.TMDB
	Tabs       []string
	TabContent []string
	activeTab  int
	//focused       string
	//searchInput   textinput.Model
	//searchResults list.Model
	movieList   list.Model
	logContent  string
	ready       bool
	logViewport viewport.Model
	windowSize  tea.WindowSizeMsg
	contentSize tea.WindowSizeMsg
	tabSize     tea.WindowSizeMsg
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "right", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.ready {
			m.initialModel()
		}
		//m.Log(fmt.Sprintf("new window size: %dx%d", msg.Width, msg.Height))
		m.setSizes()

	}

	//switch m.focused {
	//case "search":
	//	m.searchInput, cmd = m.searchInput.Update(msg)
	//case "result":
	//	m.searchResults, cmd = m.searchResults.Update(msg)
	//}
	m.logViewport, cmd = m.logViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *baseModel) Log(msg string) {
	m.logContent = fmt.Sprintf("%s\n%s", m.logContent, msg)
	m.logViewport.SetContent(m.logContent)
	m.logViewport.GotoBottom()
}

//func (m *model) Search() {
//	m.Log("start search")
//	movies, err := m.tmdb.Search(m.searchInput.Value())
//	if err != nil {
//		m.Log(fmt.Sprintf("error: %v", err))
//		return
//	}
//
//	m.Log(fmt.Sprintf("found %d results", len(movies)))
//	items := []list.Item{}
//	for _, res := range movies {
//		items = append(items, Movie{m: res})
//	}
//
//	m.searchResults.SetItems(items)
//	m.focused = "result"
//}

func (m baseModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	doc := strings.Builder{}
	doc.WriteString(m.renderMenu())
	doc.WriteString("\n")
	doc.WriteString(m.renderTabContent())
	doc.WriteString("\n")
	doc.WriteString(m.renderLog())
	return docStyle.Render(doc.String())
}

func (m *baseModel) renderMenu() string {
	var items []string
	for i, t := range m.Tabs {
		if i == m.activeTab {
			items = append(items, lipgloss.NewStyle().
				Foreground(colorHighLightForeGround).
				Render(fmt.Sprintf(" * %s ", t)))
			continue
		}

		items = append(items, lipgloss.NewStyle().
			Foreground(colorNormalForeground).
			Render(fmt.Sprintf("   %s ", t)))
	}

	return lipgloss.PlaceHorizontal(m.contentSize.Width, lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, items...))
}

func (m *baseModel) renderTabContent() string {
	content := m.TabContent[m.activeTab]
	switch m.activeTab {
	case 0:
		content = m.movieList.View()
	case 1:
		content = "tmdb"
	}

	return windowStyle.Width(m.contentSize.Width).Height(m.contentSize.Height).Render(content)
}

func (m *baseModel) renderLog() string {
	return windowStyle.Width(m.contentSize.Width).Height(logLineCount).Render(m.logViewport.View())
}

func (m *baseModel) initialModel() {
	m.movieList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.movieList.Title = "Movies"
	m.movieList.SetShowHelp(false)

	m.logViewport = viewport.New(0, 0)
	m.logViewport.KeyMap = viewport.KeyMap{}

	m.setSizes()
	m.refreshMovieList()

	m.ready = true
}

func (m *baseModel) setSizes() {
	logHeight := logLineCount + docStyle.GetVerticalFrameSize()
	menuHeight := 1

	m.contentSize.Width = m.windowSize.Width - windowStyle.GetHorizontalFrameSize() - docStyle.GetHorizontalFrameSize()
	m.contentSize.Height = m.windowSize.Height - windowStyle.GetVerticalFrameSize() - docStyle.GetVerticalFrameSize() - logHeight - menuHeight

	m.movieList.SetSize(m.contentSize.Width, m.contentSize.Height)
	m.logViewport.Width = m.contentSize.Width
	m.logViewport.Height = logLineCount
}

func (m *baseModel) refreshMovieList() {
	m.Log("fetch emdb movies...")
	ems, err := m.emdb.GetMovies()
	if err != nil {
		m.Log(err.Error())
	}
	items := make([]list.Item, len(ems))
	for i, em := range ems {
		items[i] = list.Item(Movie{m: em})
	}
	m.movieList.SetItems(items)
	m.Log(fmt.Sprintf("found %d movies in in emdb", len(items)))
}
