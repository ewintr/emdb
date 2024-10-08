package tui

import (
	"fmt"
	"strconv"
	"strings"

	"go-mod.ewintr.nl/emdb/storage"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tabReview struct {
	initialized    bool
	reviewRepo     *storage.ReviewRepository
	width          int
	height         int
	mode           string
	selectedReview storage.Review
	reviewViewport viewport.Model
	inputQuality   textinput.Model
	inputMentions  textarea.Model
	formFocus      int
	logger         *Logger
}

func NewTabReview(reviewRepo *storage.ReviewRepository, logger *Logger) (tea.Model, tea.Cmd) {
	reviewViewport := viewport.New(0, 0)
	//reviewViewport.KeyMap = viewport.KeyMap{}

	inputQuality := textinput.New()
	inputQuality.Prompt = ""
	inputQuality.Width = 50
	inputQuality.CharLimit = 500
	inputMentions := textarea.New()
	inputMentions.SetWidth(30)
	inputMentions.SetHeight(5)
	inputMentions.CharLimit = 500

	return &tabReview{
		reviewRepo:     reviewRepo,
		mode:           "view",
		reviewViewport: reviewViewport,
		inputQuality:   inputQuality,
		inputMentions:  inputMentions,
		logger:         logger,
	}, nil
}

func (m *tabReview) Init() tea.Cmd {
	return nil
}

func (m *tabReview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case TabSizeMsg:
		if !m.initialized {
			m.initialized = true
		}
		m.width = msg.Width
		m.height = msg.Height
		m.reviewViewport.Width = (m.width / 2) - 2
		m.reviewViewport.Height = m.height - 2
	case tea.KeyMsg:
		switch m.mode {
		case "edit":
			switch msg.String() {
			case "tab", "shift+tab", "up", "down":
				cmds = append(cmds, m.NavigateForm(msg.String())...)
			case "esc":
				m.mode = "view"
			case "enter":
				cmds = append(cmds, m.StoreReview())
			default:
				cmds = append(cmds, m.updateFormInputs(msg))
			}
		default:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			case "right", "tab":
				cmds = append(cmds, SelectNextTab())
			case "left", "shift+tab":
				cmds = append(cmds, SelectPrevTab())
			case "e":
				m.mode = "edit"
				m.formFocus = 0
				cmds = append(cmds, m.inputQuality.Focus())
			case "n":
				m.mode = "edit"
				m.formFocus = 0
				m.logger.Log("fetching next unrated review")
				cmds = append(cmds, m.inputQuality.Focus())
				cmds = append(cmds, FetchNextUnratedReview(m.reviewRepo))
			default:
				m.logger.Log(fmt.Sprintf("key: %s", msg.String()))
				m.reviewViewport, cmd = m.reviewViewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	case storage.Review:
		m.logger.Log(fmt.Sprintf("got review %s", msg.ID))
		m.selectedReview = msg
		review := strings.ReplaceAll(m.selectedReview.Review, "\n", "\n\n")
		review = lipgloss.NewStyle().Width((m.width / 2) - 2).Render(review)
		m.reviewViewport.SetContent(review)
		m.reviewViewport.GotoTop()
		m.UpdateForm()
	case ReviewStored:
		m.logger.Log(fmt.Sprintf("stored review %s", msg))
		cmds = append(cmds, m.inputQuality.Focus())
		cmds = append(cmds, FetchNextUnratedReview(m.reviewRepo))
	}

	return m, tea.Batch(cmds...)
}

func (m *tabReview) View() string {
	colReviewWidth := m.width / 2
	colRateWidth := m.width - colReviewWidth

	colReview := lipgloss.NewStyle().
		Width(colReviewWidth).
		Height(m.height).
		Padding(1).
		MaxHeight(m.height).
		Render(m.ViewReview())
	colRate := lipgloss.NewStyle().
		Width(colRateWidth).
		Height(m.height).
		Padding(1).
		Render(m.ViewForm())

	return lipgloss.JoinHorizontal(lipgloss.Top, colRate, colReview)
}

func (m *tabReview) UpdateForm() {
	//mentions := strings.Join(m.selectedReview.Mentions, ",")
	m.inputQuality.SetValue(fmt.Sprintf("%d", m.selectedReview.Quality))
	//m.inputMentions.SetValue(mentions)
}

func (m *tabReview) updateFormInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.formFocus {
	case 0:
		m.inputQuality, cmd = m.inputQuality.Update(msg)
	case 1:
		m.inputMentions, cmd = m.inputMentions.Update(msg)
	}

	return cmd
}

func (m *tabReview) NavigateForm(key string) []tea.Cmd {
	order := []string{"quality", "mentions"}

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
	case "quality":
		m.inputQuality.PromptStyle = focusedStyle
		m.inputQuality.TextStyle = focusedStyle
		cmds = append(cmds, m.inputQuality.Focus())
		m.inputMentions.Blur()
	case "mentions":
		cmds = append(cmds, m.inputMentions.Focus())
		m.inputQuality.Blur()
	}

	return cmds
}

func (m *tabReview) ViewForm() string {
	labels := "Quality:\nMentions:"
	fields := fmt.Sprintf("%s\n%s", m.inputQuality.View(), m.inputMentions.View())

	return lipgloss.JoinHorizontal(lipgloss.Left, labels, fields)
}

func (m *tabReview) ViewReview() string {
	return m.reviewViewport.View()
}

func (m *tabReview) StoreReview() tea.Cmd {
	return func() tea.Msg {
		quality, err := strconv.Atoi(m.inputQuality.Value())
		if err != nil {
			return err
		}
		//mentions := m.inputMentions.Value()

		m.selectedReview.Quality = quality
		//m.selectedReview.Mentions = strings.Split(mentions, ",")

		if err := m.reviewRepo.Store(m.selectedReview); err != nil {
			return err
		}

		return ReviewStored(m.selectedReview.ID)
	}
}

func FetchNextUnratedReview(reviewRepo *storage.ReviewRepository) tea.Cmd {
	return func() tea.Msg {
		review, err := reviewRepo.FindNextUnrated()
		if err != nil {
			return err
		}

		return review
	}
}
