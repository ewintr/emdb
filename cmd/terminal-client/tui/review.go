package tui

type Review struct {
	ID       string
	MovieID  string
	Source   string
	URL      string
	Review   string
	Quality  int
	Mentions []string
}

type ReviewStored string
