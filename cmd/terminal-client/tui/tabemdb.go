package tui

import (
	"fmt"
	"strconv"
	"strings"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()
)

type UpdateForm tea.Msg
type StoredMovie struct{}

type tabEMDB struct {
	initialized bool
	emdb        *client.EMDB
	mode        string
	focused     string
	colWidth    int
	list        list.Model
	formLabels  []string
	formInputs  []textinput.Model
	formFocus   int
	logger      *Logger
}

func NewTabEMDB(emdb *client.EMDB, logger *Logger) (tea.Model, tea.Cmd) {
	del := list.NewDefaultDelegate()
	list := list.New([]list.Item{}, del, 0, 0)
	list.Title = "Movies"
	list.SetShowHelp(false)

	formLabels := []string{
		"Watched on",
		"Rating",
		"Comment",
	}
	formInputs := make([]textinput.Model, len(formLabels))
	for i := range formLabels {
		formInputs[i] = textinput.New()
		formInputs[i].Prompt = ""
		formInputs[i].Width = 50
		formInputs[i].CharLimit = 500
	}

	m := tabEMDB{
		focused:    "form",
		emdb:       emdb,
		logger:     logger,
		mode:       "view",
		list:       list,
		formLabels: formLabels,
		formInputs: formInputs,
	}

	logger.Log("search emdb...")
	return m, FetchMovieList(emdb)
}

func (m tabEMDB) Init() tea.Cmd {
	return nil
}

func (m tabEMDB) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case TabSizeMsg:
		if !m.initialized {
			m.initialized = true
		}
		m.colWidth = msg.Width / 2
		m.list.SetSize(m.colWidth, msg.Height)
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case Movies:
		m.logger.Log(fmt.Sprintf("found %d movies in in emdb", len(msg)))
		m.list.SetItems(msg.listItems())
		m.list.Select(0)
		m.UpdateForm()
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case StoredMovie:
		m.logger.Log("stored movie, fetching movie list")
		cmds = append(cmds, FetchMovieList(m.emdb))
	case tea.KeyMsg:
		switch m.mode {
		case "edit":
			switch msg.String() {
			case "tab", "shift+tab", "up", "down":
				cmds = append(cmds, m.NavigateForm(msg.String())...)
			case "enter":
				m.mode = "view"
				cmds = append(cmds, m.StoreMovie())
			default:
				cmds = append(cmds, m.updateFormInputs(msg))
			}
		default:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "right", "tab":
				cmds = append(cmds, SelectNextTab())
			case "left", "shift+tab":
				cmds = append(cmds, SelectPrevTab())
			case "up":
				m.list, cmd = m.list.Update(msg)
				m.UpdateForm()
				cmds = append(cmds, cmd)
			case "down":
				m.list, cmd = m.list.Update(msg)
				m.UpdateForm()
				cmds = append(cmds, cmd)
			case "e":
				m.mode = "edit"
				m.formFocus = 0
				m.formInputs[0].Focus()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *tabEMDB) updateFormInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.formInputs))
	for i := range m.formInputs {
		m.formInputs[i], cmds[i] = m.formInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m tabEMDB) View() string {
	colLeft := lipgloss.NewStyle().Width(m.colWidth).Render(m.list.View())
	colRight := lipgloss.NewStyle().Width(m.colWidth).Render(m.ViewForm())

	return lipgloss.JoinHorizontal(lipgloss.Top, colLeft, colRight)
}

func (m *tabEMDB) UpdateForm() {
	movie, ok := m.list.SelectedItem().(Movie)
	if !ok {
		return
	}
	m.Log(fmt.Sprintf("updating form with movie %s", movie.m.Title))
	m.formInputs[0].SetValue(movie.m.WatchedOn)
	m.formInputs[1].SetValue(fmt.Sprintf("%d", movie.m.Rating))
	m.formInputs[2].SetValue(movie.m.Comment)
}

func (m *tabEMDB) ViewForm() string {
	movie, ok := m.list.SelectedItem().(Movie)
	if !ok {
		return ""
	}

	labels := []string{
		"ID: ",
		"TMDBID: ",
		"IMDBID: ",
		"Title",
		"English title",
		"Year",
		"Directors",
		"Summary",
	}
	for _, l := range m.formLabels {
		labels = append(labels, fmt.Sprintf("%s: ", l))
	}

	fields := []string{
		movie.m.ID,
		fmt.Sprintf("%d", movie.m.TMDBID),
		movie.m.IMDBID,
		movie.m.Title,
		movie.m.EnglishTitle,
		fmt.Sprintf("%d", movie.m.Year),
		strings.Join(movie.m.Directors, ","),
		movie.m.Summary,
	}
	for _, f := range m.formInputs {
		fields = append(fields, f.View())
	}
	labelView := strings.Join(labels, "\n")
	fieldsView := strings.Join(fields, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, labelView, fieldsView)
}

func (m *tabEMDB) NavigateForm(key string) []tea.Cmd {
	var cmds []tea.Cmd
	if key == "up" || key == "shift+tab" {
		m.formFocus--
	} else {
		m.formFocus++
	}
	if m.formFocus > len(m.formInputs) {
		m.formFocus = 0
	}
	if m.formFocus < 0 {
		m.formFocus = len(m.formInputs)
	}
	for i := 0; i <= len(m.formInputs)-1; i++ {
		if i == m.formFocus {
			m.formInputs[i].PromptStyle = focusedStyle
			m.formInputs[i].TextStyle = focusedStyle
			cmds = append(cmds, m.formInputs[i].Focus())
			continue
		}
		m.formInputs[i].Blur()
		m.formInputs[i].PromptStyle = noStyle
		m.formInputs[i].TextStyle = noStyle
	}

	return cmds
}

func (m *tabEMDB) StoreMovie() tea.Cmd {
	return func() tea.Msg {
		updatedMovie := m.list.SelectedItem().(Movie)
		updatedMovie.m.WatchedOn = m.formInputs[0].Value()
		var err error
		if updatedMovie.m.Rating, err = strconv.Atoi(m.formInputs[1].Value()); err != nil {
			return fmt.Errorf("rating cannot be converted to an int: %w", err)
		}
		updatedMovie.m.Comment = m.formInputs[2].Value()
		if _, err := m.emdb.CreateMovie(updatedMovie.m); err != nil {
			return err
		}
		return StoredMovie{}
	}
}

func (m *tabEMDB) Log(s string) {
	m.logger.Log(s)
}

func FetchMovieList(emdb *client.EMDB) tea.Cmd {
	return func() tea.Msg {
		ems, err := emdb.GetMovies()
		if err != nil {
			return err
		}
		return Movies(ems)
	}
}
