package backend

const (
	CommandAdd            = "add"
	CommandRefreshWatched = "refreshWatched"

	ArgName = "name"
)

type CommandName string

type Command struct {
	Name string
	Args map[string]any
}
