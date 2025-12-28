package core

type EventMessage struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

type Action struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	TargetID string `json:"target_id"`
	Text     string `json:"text"`
}

type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

type PluginConfig struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	APIVersion      string   `json:"api_version"`
	Version         string   `json:"version"`
	EntryPoint      string   `json:"entry_point"`
	RunOn           []string `json:"run_on"`
	Capabilities    []string `json:"capabilities"`
	Actions         []string `json:"actions"`
	TimeoutMS       int      `json:"timeout_ms"`
	MaxConcurrency  int      `json:"max_concurrency"`
	MaxRestarts     int      `json:"max_restarts"`
	Signature       string   `json:"signature"`
	PluginLevel     string   `json:"plugin_level"`
	Source          string   `json:"source"`
}
