package tui

//focused       string
//searchInput   textinput.Model
//searchResults list.Model
//movieList   list.Model

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
