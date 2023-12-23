package tui

import (
	"fmt"
	"strings"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
			m.Log("switch to next tab")
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "shift+tab":
			m.Log("switch to previous tab")
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			m.windowSize = msg
			m.Log(fmt.Sprintf("new window size: %dx%d", msg.Width, msg.Height))
			m.initialModel(msg.Width, msg.Height)
		}
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

	contentWidth := m.windowSize.Width - docStyle.GetHorizontalFrameSize() - docStyle.GetHorizontalFrameSize()
	m.Log(fmt.Sprintf("content width: %d", contentWidth))
	doc := strings.Builder{}
	doc.WriteString(m.renderMenu(contentWidth))
	doc.WriteString("\n")
	doc.WriteString(m.renderTabContent(contentWidth))
	doc.WriteString("\n")
	doc.WriteString(m.renderLog(contentWidth))
	return docStyle.Render(doc.String())
}

func (m *baseModel) renderMenu(width int) string {
	var items []string
	for i, t := range m.Tabs {
		if i == m.activeTab {
			items = append(items, lipgloss.NewStyle().
				Foreground(colorHighLightForeGround).
				//		Background(lipgloss.ANSIColor(termenv.ANSIBlack)).
				Render(fmt.Sprintf(" * %s ", t)))
			continue
		}

		items = append(items, lipgloss.NewStyle().
			Foreground(colorNormalForeground).
			//		Background(lipgloss.ANSIColor(termenv.ANSIBlack)).
			Render(fmt.Sprintf("   %s ", t)))
	}

	return lipgloss.PlaceHorizontal(width, lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, items...))
}

func (m *baseModel) renderTabContent(width int) string {
	content := m.TabContent[m.activeTab]

	return windowStyle.Width(width).Render(content)
}

func (m *baseModel) renderLog(width int) string {
	return windowStyle.Width(width).Render(m.logViewport.View())
}

func (m *baseModel) initialModel(width, height int) {

	si := textinput.New()
	si.Placeholder = "title"
	si.CharLimit = 156
	si.Width = 20
	//m.searchInput = si
	//m.searchInput.Focus()
	//
	//m.searchResults = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height-50)
	//m.searchResults.Title = "Search results"
	//m.searchResults.SetShowHelp(false)

	m.Log("fetch emdb movies")
	ems, err := m.emdb.GetMovies()
	if err != nil {
		m.Log(err.Error())
	}
	items := make([]list.Item, len(ems))
	for i, em := range ems {
		items[i] = list.Item(Movie{m: em})
	}
	m.Log(fmt.Sprintf("found %d movies in in emdb", len(items)))

	m.movieList = list.New(items, list.NewDefaultDelegate(), width, height-10)
	m.movieList.Title = "Movies"
	m.movieList.SetShowHelp(false)

	m.logViewport = viewport.New(width, 10)
	m.logViewport.SetContent(m.logContent)
	m.logViewport.KeyMap = viewport.KeyMap{}
	//m.focused = "search"
	m.ready = true
}
