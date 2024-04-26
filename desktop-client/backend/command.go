package backend

const (
	CommandAdd = "add"

	ArgName = "name"
)

type CommandName string

type Command struct {
	Name string
	Args map[string]any
}
