package tui

import (
	"fmt"
	"strconv"
	"strings"

	"ewintr.nl/emdb/client"
	"ewintr.nl/emdb/model"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		"ID",
		"TMDB ID",
		"IMDB ID",
		"Title",
		"English Title",
		"Year",
		"Director",
		"Summary",
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
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case StoredMovie:
		m.logger.Log("stored movie, fetching movie list")
		cmds = append(cmds, FetchMovieList(m.emdb))
	case tea.KeyMsg:
		m.Log("key msg")
		switch m.mode {
		case "edit":
			m.Log("processing edit mode")
			switch msg.String() {
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" {
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
						cmds = append(cmds, m.formInputs[i].Focus())
						continue
					}
					m.formInputs[i].Blur()
				}
			case "enter":
				m.mode = "view"
				cmds = append(cmds, m.StoreMovie())
			default:
				cmds = append(cmds, m.updateFormInputs(msg))
			}
		default:
			m.Log("processing view mode")
			switch msg.String() {
			case "up":
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
				m.SetForm(m.list.SelectedItem().(Movie))
			case "down":
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
				m.SetForm(m.list.SelectedItem().(Movie))
			case "e":
				m.mode = "edit"
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
	labels := make([]string, len(m.formLabels))
	for i := range m.formLabels {
		labels[i] = fmt.Sprintf("%s: ", m.formLabels[i])
	}
	fields := make([]string, len(m.formLabels))
	for i := range m.formLabels {
		fields[i] = m.formInputs[i].View()
	}
	labelView := strings.Join(labels, "\n")
	fieldsView := strings.Join(fields, "\n")
	form := lipgloss.JoinHorizontal(lipgloss.Top, labelView, fieldsView)

	colLeft := lipgloss.NewStyle().Width(m.colWidth).Render(m.list.View())
	colRight := lipgloss.NewStyle().Width(m.colWidth).Render(form)

	return lipgloss.JoinHorizontal(lipgloss.Top, colLeft, colRight)
}

func (m *tabEMDB) SetForm(movie Movie) {
	m.formInputs[0].SetValue(movie.m.ID)
	m.formInputs[1].SetValue(fmt.Sprintf("%d", movie.m.TMDBID))
	m.formInputs[2].SetValue(movie.m.IMDBID)
	m.formInputs[3].SetValue(movie.m.Title)
	m.formInputs[4].SetValue(movie.m.EnglishTitle)
	m.formInputs[5].SetValue(fmt.Sprintf("%d", movie.m.Year))
	m.formInputs[6].SetValue(strings.Join(movie.m.Directors, ","))
	m.formInputs[7].SetValue(movie.m.Summary)
	m.formInputs[8].SetValue(fmt.Sprintf("%d", movie.m.Rating))
	m.formInputs[9].SetValue(movie.m.Comment)
}

func (m *tabEMDB) StoreMovie() tea.Cmd {
	return func() tea.Msg {
		tmdbId, err := strconv.Atoi(m.formInputs[1].Value())
		if err != nil {
			return fmt.Errorf("tmbID cannot be converted to an int: %w", err)
		}
		year, err := strconv.Atoi(m.formInputs[5].Value())
		if err != nil {
			return fmt.Errorf("year cannot be converted to an int: %w", err)
		}
		rating, err := strconv.Atoi(m.formInputs[8].Value())
		if err != nil {
			return fmt.Errorf("rating cannot be converted to an int: %w", err)
		}
		movie := Movie{
			m: model.Movie{
				ID:           m.formInputs[0].Value(),
				TMDBID:       int64(tmdbId),
				IMDBID:       m.formInputs[2].Value(),
				Title:        m.formInputs[3].Value(),
				EnglishTitle: m.formInputs[4].Value(),
				Year:         year,
				Directors:    strings.Split(m.formInputs[6].Value(), ","),
				Summary:      m.formInputs[7].Value(),
				Rating:       rating,
				Comment:      m.formInputs[9].Value(),
			},
		}
		m.Log(fmt.Sprintf("storing movie %s", movie.Title()))
		if _, err := m.emdb.AddMovie(movie.m); err != nil {
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
