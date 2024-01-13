package tui

import (
	"fmt"
	"strconv"
	"strings"

	"ewintr.nl/emdb/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
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
	initialized    bool
	emdb           *client.EMDB
	mode           string
	focused        string
	colWidth       int
	colHeight      int
	list           list.Model
	formLabels     []string
	inputWatchedOn textinput.Model
	inputRating    textinput.Model
	inputComment   textarea.Model
	formFocus      int
	logger         *Logger
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

	inputWatchedOn := textinput.New()
	inputWatchedOn.Prompt = ""
	inputWatchedOn.Width = 50
	inputWatchedOn.CharLimit = 500
	inputRating := textinput.New()
	inputRating.Prompt = ""
	inputRating.Width = 50
	inputRating.CharLimit = 500
	inputComment := textarea.New()
	inputComment.SetWidth(50)
	inputComment.SetHeight(3)
	inputComment.CharLimit = 500

	m := tabEMDB{
		focused:        "form",
		emdb:           emdb,
		logger:         logger,
		mode:           "view",
		list:           list,
		formLabels:     formLabels,
		inputWatchedOn: inputWatchedOn,
		inputRating:    inputRating,
		inputComment:   inputComment,
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
		m.colHeight = msg.Height
		m.list.SetSize(m.colWidth-4, msg.Height-4)
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case Movies:
		m.logger.Log(fmt.Sprintf("found %d movies in in emdb", len(msg)))
		m.list.SetItems(msg.listItems())
		m.list.Select(len(msg.listItems()) - 1)
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
				m.inputWatchedOn.PromptStyle = focusedStyle
				m.inputWatchedOn.TextStyle = focusedStyle
				cmds = append(cmds, m.inputWatchedOn.Focus())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m tabEMDB) View() string {
	colLeft := lipgloss.NewStyle().
		Width(m.colWidth - 2).
		Height(m.colHeight - 2).
		Padding(1).
		Render(m.list.View())
	colRight := lipgloss.NewStyle().
		Width(m.colWidth - 2).
		Height(m.colHeight - 2).
		Padding(1).
		Render(m.ViewForm())

	return lipgloss.JoinHorizontal(lipgloss.Top, colLeft, colRight)
}

func (m *tabEMDB) UpdateForm() {
	movie, ok := m.list.SelectedItem().(Movie)
	if !ok {
		return
	}
	m.inputWatchedOn.SetValue(movie.m.WatchedOn)
	m.inputRating.SetValue(fmt.Sprintf("%d", movie.m.Rating))
	m.inputComment.SetValue(movie.m.Comment)
	m.Log(fmt.Sprintf("showing movie %s", movie.m.ID))
}

func (m *tabEMDB) updateFormInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.formFocus {
	case 0:
		m.inputWatchedOn, cmd = m.inputWatchedOn.Update(msg)
	case 1:
		m.inputRating, cmd = m.inputRating.Update(msg)
	case 2:
		m.inputComment, cmd = m.inputComment.Update(msg)
	}
	return cmd
}

func (m *tabEMDB) NavigateForm(key string) []tea.Cmd {
	order := []string{"Watched on", "Rating", "Comment"}

	var cmds []tea.Cmd
	if key == "up" || key == "shift+tab" {
		m.formFocus--
	} else {
		m.formFocus++
	}
	if m.formFocus >= len(order) {
		m.formFocus = 0
	}
	if m.formFocus < 0 {
		m.formFocus = len(order) - 1
	}

	switch order[m.formFocus] {
	case "Watched on":
		m.inputWatchedOn.PromptStyle = focusedStyle
		m.inputWatchedOn.TextStyle = focusedStyle
		cmds = append(cmds, m.inputWatchedOn.Focus())
		m.inputRating.Blur()
		m.inputComment.Blur()
	case "Rating":
		m.inputRating.PromptStyle = focusedStyle
		m.inputRating.TextStyle = focusedStyle
		cmds = append(cmds, m.inputRating.Focus())
		m.inputWatchedOn.Blur()
		m.inputComment.Blur()
	case "Comment":
		cmds = append(cmds, m.inputComment.Focus())
		m.inputWatchedOn.Blur()
		m.inputRating.Blur()
	}

	return cmds
}

func (m *tabEMDB) ViewForm() string {
	movie, ok := m.list.SelectedItem().(Movie)
	if !ok {
		return ""
	}

	labels := []string{
		"Title: ",
		"English title: ",
		"Year: ",
		"Directors: ",
		"Summary: ",
	}
	for _, l := range m.formLabels {
		labels = append(labels, fmt.Sprintf("%s: ", l))
	}

	fields := []string{
		movie.m.Title,
		movie.m.EnglishTitle,
		fmt.Sprintf("%d", movie.m.Year),
		strings.Join(movie.m.Directors, ","),
		movie.m.Summary,
	}

	fields = append(fields, m.inputWatchedOn.View(), m.inputRating.View(), m.inputComment.View())

	labelView := strings.Join(labels, "\n")
	fieldsView := strings.Join(fields, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, labelView, fieldsView)
}

func (m *tabEMDB) StoreMovie() tea.Cmd {
	return func() tea.Msg {
		updatedMovie := m.list.SelectedItem().(Movie)
		updatedMovie.m.WatchedOn = m.inputWatchedOn.Value()
		var err error
		if updatedMovie.m.Rating, err = strconv.Atoi(m.inputRating.Value()); err != nil {
			return fmt.Errorf("rating cannot be converted to an int: %w", err)
		}
		updatedMovie.m.Comment = m.inputComment.Value()
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
