package movie

type Movie struct {
	ID           string   `json:"id"`
	TMDBID       int64    `json:"tmdbID"`
	IMDBID       string   `json:"imdbID"`
	Title        string   `json:"title"`
	EnglishTitle string   `json:"englishTitle"`
	Year         int      `json:"year"`
	Directors    []string `json:"directors"`
	WatchedOn    string   `json:"watchedOn"`
	Rating       int      `json:"rating"`
	Summary      string   `json:"summary"`
	Comment      string   `json:"comment"`
}

type MovieRepository interface {
	Store(movie *Movie) error
	FindOne(id string) (*Movie, error)
	FindAll() ([]*Movie, error)
	Delete(id string) error
}
