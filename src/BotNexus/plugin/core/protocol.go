package core

type EventMessage struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Payload any    `json:"payload"`
}

type Action struct {
	Type     string         `json:"type"`
	Target   string         `json:"target"`
	TargetID string         `json:"target_id"`
	Text     string         `json:"text"`
	Payload  map[string]any `json:"payload,omitempty"` // Added for rich actions like skill calls
}

type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

type Intent struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
	Priority int      `json:"priority"`
}

type UIComponent struct {
	Type     string `json:"type"`      // "panel", "button", "tab"
	Position string `json:"position"`  // "sidebar", "dashboard", "chat_action"
	Entry    string `json:"entry"`     // URL or HTML file path
	Title    string `json:"title"`
	Icon     string `json:"icon"`
}

type PluginConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Author      string        `json:"author"`
	EntryPoint  string        `json:"entry"`
	Permissions []string      `json:"permissions"`
	Events      []string      `json:"events"`
	Intents     []Intent      `json:"intents,omitempty"` // Added for AI/keyword routing
	UI          []UIComponent `json:"ui,omitempty"`      // Added for UI extensions
	RunOn       []string      `json:"run_on"`
	TimeoutMS   int           `json:"timeout_ms"`
	MaxRestarts int           `json:"max_restarts"`
	Signature   string        `json:"signature"`
}
