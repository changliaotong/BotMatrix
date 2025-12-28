package core

type EventMessage struct {
	ID            string `json:"id"`
	Type          string `json:"type"` // "event", "request", "response"
	Name          string `json:"name"` // "on_message", "on_group_message", etc.
	CorrelationId string `json:"correlation_id,omitempty"`
	Payload       any    `json:"payload"`
}

type Action struct {
	Type     string         `json:"type"` // "send_text", "send_image", "call_skill", etc.
	Target   string         `json:"target,omitempty"`
	TargetID string         `json:"target_id,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
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
	Type     string `json:"type"`     // "panel", "button", "tab"
	Position string `json:"position"` // "sidebar", "dashboard", "chat_action"
	Entry    string `json:"entry"`    // URL or HTML file path
	Title    string `json:"title"`
	Icon     string `json:"icon"`
}

type PluginConfig struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Version        string        `json:"version"`
	EntryPoint     string        `json:"entry_point"`
	RunOn          []string      `json:"run_on"`      // "center", "worker"
	Permissions    []string      `json:"permissions"` // List of allowed actions
	Events         []string      `json:"events"`      // List of events to subscribe to
	Intents        []Intent      `json:"intents"`
	UI             []UIComponent `json:"ui,omitempty"`
	MaxRestarts    int           `json:"max_restarts"`
	CanaryWeight   int           `json:"canary_weight,omitempty"` // 0-100
	Signature      string        `json:"signature,omitempty"`
}
