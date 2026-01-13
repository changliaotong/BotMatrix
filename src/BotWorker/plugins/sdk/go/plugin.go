package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EventType defines the type of message from core
type EventType string

const (
	TypeEvent EventType = "event"
)

// EventMessage represents an event received from the core
type EventMessage struct {
	ID            string         `json:"id"`
	Type          EventType      `json:"type"`
	Name          string         `json:"name"`
	CorrelationID string         `json:"correlation_id,omitempty"` // Link back to an Ask request
	Payload       map[string]any `json:"payload"`
}

// Action represents an action to be performed
type Action struct {
	Type          string         `json:"type"`
	Target        string         `json:"target"`
	TargetID      string         `json:"target_id"`
	Text          string         `json:"text"`
	CorrelationID string         `json:"correlation_id,omitempty"` // For cross-instance tracking
	Payload       map[string]any `json:"payload,omitempty"`
}

// ResponseMessage represents a response to the core
type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
	Error   string   `json:"error,omitempty"`
}

// Context provides helper methods for handlers
type Context struct {
	Event   *EventMessage
	Actions []Action
	Args    []string          // For command arguments
	Params  map[string]string // For named parameters
	Result  string            // For string-based message propagation
	plugin  *Plugin
	mu      sync.Mutex
}

// Reply adds a send_message action to the response (immediate construction)
func (c *Context) Reply(text string) {
	c.CallAction("send_message", map[string]any{
		"text": text,
	})
}

// SendText is an alias for Reply for consistency
func (c *Context) SendText(text string) {
	c.Reply(text)
}

// SendImage sends an image message
func (c *Context) SendImage(url string) {
	c.CallAction("send_image", map[string]any{
		"url": url,
	})
}

// AddAction adds a custom action
func (c *Context) AddAction(actionType string, payload map[string]any) {
	c.CallAction(actionType, payload)
}

// Ask sends a prompt and waits for the user's next message
func (c *Context) Ask(prompt string, timeout time.Duration) (*Context, error) {
	// Generate a unique correlation ID for this interaction
	correlationID := fmt.Sprintf("ask_%s_%d", c.Event.ID, time.Now().UnixNano())

	c.mu.Lock()
	c.Actions = append(c.Actions, Action{
		Type:          "send_message",
		Target:        c.Event.Payload["from"].(string),
		TargetID:      c.Event.Payload["group_id"].(string),
		Text:          prompt,
		CorrelationID: correlationID, // Tell core to track this
	})
	c.mu.Unlock()

	ch := make(chan *Context, 1)
	c.plugin.mu.Lock()
	if c.plugin.waitingSessions == nil {
		c.plugin.waitingSessions = make(map[string]chan *Context)
	}
	c.plugin.waitingSessions[correlationID] = ch
	c.plugin.mu.Unlock()

	defer func() {
		c.plugin.mu.Lock()
		delete(c.plugin.waitingSessions, correlationID)
		c.plugin.mu.Unlock()
	}()

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

// CallSkill calls a skill exported by another plugin
func (c *Context) CallSkill(pluginID, skillName string, payload map[string]any) (*Context, error) {
	correlationID := fmt.Sprintf("skill_%s_%d", skillName, time.Now().UnixNano())

	c.mu.Lock()
	skillPayload := map[string]any{
		"plugin_id":      pluginID,
		"skill":          skillName,
		"correlation_id": correlationID,
	}
	for k, v := range payload {
		skillPayload[k] = v
	}

	c.Actions = append(c.Actions, Action{
		Type:    "call_skill",
		Payload: skillPayload,
	})
	c.mu.Unlock()

	ch := make(chan *Context, 1)
	c.plugin.mu.Lock()
	if c.plugin.waitingSessions == nil {
		c.plugin.waitingSessions = make(map[string]chan *Context)
	}
	c.plugin.waitingSessions[correlationID] = ch
	c.plugin.mu.Unlock()

	defer func() {
		c.plugin.mu.Lock()
		delete(c.plugin.waitingSessions, correlationID)
		c.plugin.mu.Unlock()
	}()

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(10 * time.Second):
		return nil, context.DeadlineExceeded
	}
}

// Delete deletes a message by ID
func (c *Context) Delete(messageId string) {
	c.CallAction("delete_message", map[string]any{
		"message_id": messageId,
	})
}

// Kick removes a user from a group
func (c *Context) Kick(groupId, userId string) {
	c.CallAction("kick_user", map[string]any{
		"group_id": groupId,
		"user_id":  userId,
	})
}

// CallAction is the "Escape Hatch" to send any action supported by the core
func (c *Context) CallAction(actionType string, payload map[string]any) {
	// Permission check
	if actionType != "call_skill" && !c.plugin.HasPermission(actionType) {
		fmt.Fprintf(os.Stderr, "Permission denied: Action '%s' is not declared in plugin.json\n", actionType)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	from := ""
	if f, ok := c.Event.Payload["from"].(string); ok {
		from = f
	}

	groupID := ""
	if g, ok := c.Event.Payload["group_id"].(string); ok {
		groupID = g
	}

	action := Action{
		Type:     actionType,
		Target:   from,
		TargetID: groupID,
		Payload:  payload,
	}

	// Legacy support for 'text' field in some actions
	if text, ok := payload["text"].(string); ok {
		action.Text = text
	}

	c.Actions = append(c.Actions, action)
}

// Handler is a function that handles an event message
type Handler func(ctx *Context) error

// Middleware defines a function that wraps a Handler
type Middleware func(next Handler) Handler

// Plugin represents a bot plugin
type Plugin struct {
	handlers        map[string]Handler
	middlewares     []Middleware
	waitingSessions map[string]chan *Context
	config          map[string]any
	mu              sync.RWMutex
}

// NewPlugin creates a new plugin instance
func NewPlugin() *Plugin {
	p := &Plugin{
		handlers:        make(map[string]Handler),
		waitingSessions: make(map[string]chan *Context),
	}
	p.loadConfig("plugin.json")
	return p
}

func (p *Plugin) loadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	var config map[string]any
	if err := json.NewDecoder(file).Decode(&config); err == nil {
		p.config = config
	}
}

func (p *Plugin) HasPermission(action string) bool {
	if p.config == nil {
		return true // Legacy mode
	}

	// Essential built-in actions are always allowed
	switch action {
	case "send_message", "send_image", "storage.get", "storage.set":
		return true
	}

	actions, ok := p.config["actions"].([]any)
	if !ok {
		return true
	}
	for _, a := range actions {
		if s, ok := a.(string); ok && s == action {
			return true
		}
	}
	return false
}

// Use adds middlewares to the plugin
func (p *Plugin) Use(m ...Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.middlewares = append(p.middlewares, m...)
}

// On registers a handler for a specific event name
func (p *Plugin) On(eventName string, handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[eventName] = handler
}

// OnMessage registers a handler for "on_message" events
func (p *Plugin) OnMessage(handler Handler) {
	p.On("on_message", handler)
}

// OnIntent registers a handler for a specific intent name
func (p *Plugin) OnIntent(intentName string, handler Handler) {
	p.On("intent_"+intentName, handler)
}

// Command registers a handler for a specific command (e.g., "/echo")
func (p *Plugin) Command(cmd string, handler Handler) {
	p.OnMessage(func(ctx *Context) error {
		text, _ := ctx.Event.Payload["text"].(string)
		if strings.HasPrefix(text, cmd+" ") || text == cmd {
			ctx.Args = strings.Fields(strings.TrimPrefix(text, cmd))
			return handler(ctx)
		}
		return nil
	})
}

// RegexCommand registers a handler using regular expressions
func (p *Plugin) RegexCommand(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	p.OnMessage(func(ctx *Context) error {
		text, _ := ctx.Event.Payload["text"].(string)
		match := re.FindStringSubmatch(text)
		if match != nil {
			ctx.Args = match
			ctx.Params = make(map[string]string)
			groupNames := re.SubexpNames()
			for i, name := range groupNames {
				if i != 0 && name != "" {
					ctx.Params[name] = match[i]
				}
			}
			return handler(ctx)
		}
		return nil
	})
}

// ExportSkill allows a plugin to expose a function that can be called by other plugins
func (p *Plugin) ExportSkill(name string, handler Handler) {
	p.On("skill_"+name, handler)
}

// Run starts the plugin event loop
func (p *Plugin) Run() {
	decoder := json.NewDecoder(os.Stdin)
	// Use a channel to serialize output to stdout to avoid interleaved JSON
	outputChan := make(chan ResponseMessage, 100)

	// Start output worker
	go func() {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		for resp := range outputChan {
			if err := encoder.Encode(resp); err != nil {
				fmt.Fprintf(os.Stderr, "[SDK] Error encoding response: %v\n", err)
			}
		}
	}()

	for {
		var msg EventMessage
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "[SDK] Error decoding message: %v\n", err)
			continue
		}

		if msg.Type == TypeEvent {
			go p.handleEvent(&msg, outputChan)
		}
	}
	close(outputChan)
}

func (p *Plugin) handleEvent(msg *EventMessage, outputChan chan<- ResponseMessage) {
	// 1. Check by CorrelationID first (The most reliable way in distributed systems)
	if msg.CorrelationID != "" {
		p.mu.RLock()
		ch, ok := p.waitingSessions[msg.CorrelationID]
		p.mu.RUnlock()

		if ok {
			ch <- &Context{Event: msg, plugin: p}
			outputChan <- ResponseMessage{ID: msg.ID, OK: true}
			return
		}
	}

	// 2. Fallback to session key (for local backward compatibility)
	if msg.Name == "on_message" {
		from, _ := msg.Payload["from"].(string)
		groupID, _ := msg.Payload["group_id"].(string)
		sessionKey := fmt.Sprintf("%s:%s", groupID, from)

		p.mu.RLock()
		ch, ok := p.waitingSessions[sessionKey]
		p.mu.RUnlock()

		if ok {
			ch <- &Context{Event: msg, plugin: p}
			outputChan <- ResponseMessage{ID: msg.ID, OK: true}
			return
		}
	}

	p.mu.RLock()
	handler, ok := p.handlers[msg.Name]
	middlewares := p.middlewares
	p.mu.RUnlock()

	if !ok {
		outputChan <- ResponseMessage{ID: msg.ID, OK: true}
		return
	}

	// Wrap handler with middlewares in reverse order
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}

	ctx := &Context{Event: msg, plugin: p}

	// Recover from panics in handlers
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[SDK] Panic in handler %s: %v\n", msg.Name, r)
			outputChan <- ResponseMessage{ID: msg.ID, OK: false, Error: fmt.Sprintf("panic: %v", r)}
		}
	}()

	err := finalHandler(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[SDK] Handler error for %s: %v\n", msg.Name, err)
		outputChan <- ResponseMessage{ID: msg.ID, OK: false, Error: err.Error()}
	} else {
		// Auto-convert string result to send_message action
		if ctx.Result != "" {
			ctx.Reply(ctx.Result)
		}
		outputChan <- ResponseMessage{ID: msg.ID, OK: true, Actions: ctx.Actions}
	}
}
