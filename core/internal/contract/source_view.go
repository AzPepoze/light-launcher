package contract

// Where a source action is shown.
const (
	ActionScopeLibrary = "library"
	ActionScopeGame    = "game"
)

// SourceAction is a button a source adds to the Library. Method is the plugin method the host calls.
type SourceAction struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Icon   string `json:"icon,omitempty"`
	Method string `json:"method"`
	Scope  string `json:"scope,omitempty"`
}

// SourceCard names the game fields a Library card shows. Empty fields use the defaults: title and version.
type SourceCard struct {
	Title    string   `json:"title,omitempty"`
	Subtitle string   `json:"subtitle,omitempty"`
	Badges   []string `json:"badges,omitempty"`
}

// ActionRequest runs one source action. GameID is null for a Library action.
type ActionRequest struct {
	Source string `json:"source"`
	Action string `json:"action"`
	GameID *int64 `json:"gameId"`
}

// ActionResult is the plugin's answer to an action. Refresh asks the host to reload the Library list.
type ActionResult struct {
	Message string `json:"message"`
	Refresh bool   `json:"refresh"`
}
