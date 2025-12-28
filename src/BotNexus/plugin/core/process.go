package core

import (
	log "BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"BotNexus/plugin/policy"
)

func (pm *PluginManager) StartPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	if plugin.State == "running" {
		return nil
	}

	return pm.startPluginInstance(plugin)
}

func (pm *PluginManager) monitorPlugin(plugin *Plugin) {
	for {
		if plugin.State != "running" {
			return
		}

		_, err := os.FindProcess(plugin.Process.Pid)
		if err != nil {
			plugin.State = "crashed"
			pm.restartPlugin(plugin)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func (pm *PluginManager) restartPlugin(plugin *Plugin) {
	if plugin.RestartCount >= plugin.Config.MaxRestarts {
		log.Printf("plugin %s has reached max restarts", plugin.ID)
		return
	}

	if time.Since(plugin.LastRestart) < 5*time.Second {
		log.Printf("plugin %s restarted too recently, waiting...", plugin.ID)
		time.Sleep(5 * time.Second)
	}

	plugin.RestartCount++
	plugin.LastRestart = time.Now()
	pm.StartPlugin(plugin.ID)
}

func (pm *PluginManager) readPluginOutput(plugin *Plugin) {
	decoder := json.NewDecoder(plugin.Stdout)
	for {
		var resp ResponseMessage
		if err := decoder.Decode(&resp); err != nil {
			if plugin.State != "running" {
				return
			}
			log.Printf("plugin %s output error: %v", plugin.ID, err)
			return
		}

		pm.handlePluginResponse(plugin, &resp)
	}
}

func (pm *PluginManager) handlePluginResponse(plugin *Plugin, resp *ResponseMessage) {
	for _, action := range resp.Actions {
		// Special handling for cross-plugin skill calls
		if action.Type == "call_skill" {
			pm.routeSkillCall(plugin, action)
			continue
		}

		if !pm.isActionAllowed(plugin, action.Type) {
			log.Printf("plugin %s tried to execute forbidden action %s", plugin.ID, action.Type)
			continue
		}

		log.Printf("executing action %s from plugin %s", action.Type, plugin.ID)
	}
}

func (pm *PluginManager) isActionAllowed(plugin *Plugin, actionType string) bool {
	// 1. First check if the action is declared in the plugin's manifest
	declared := false
	for _, p := range plugin.Config.Permissions {
		if p == actionType {
			declared = true
			break
		}
	}

	if !declared {
		return false
	}

	// 2. Then check against the system's global policy (Center or Worker)
	// If RunOn is empty, default to worker policy for safety
	if len(plugin.Config.RunOn) == 0 {
		return policy.WorkerActionWhitelist[actionType]
	}

	for _, runOn := range plugin.Config.RunOn {
		switch runOn {
		case "center":
			if policy.CenterActionWhitelist[actionType] {
				return true
			}
		case "worker":
			if policy.WorkerActionWhitelist[actionType] {
				return true
			}
		}
	}

	return false
}

func (pm *PluginManager) routeSkillCall(source *Plugin, action Action) {
	targetID, ok := action.Payload["target_plugin"].(string)
	if !ok {
		log.Printf("Invalid skill call from %s: missing target_plugin", source.ID)
		return
	}

	skillName, ok := action.Payload["skill_name"].(string)
	if !ok {
		log.Printf("Invalid skill call from %s: missing skill_name", source.ID)
		return
	}

	payload, _ := action.Payload["payload"].(map[string]any)

	pm.mutex.Lock()
	targetPlugin, exists := pm.plugins[targetID]
	pm.mutex.Unlock()

	if !exists || targetPlugin.State != "running" {
		log.Printf("Skill call failed: target plugin %s not found or not running", targetID)
		return
	}

	// Inject as a new event to the target plugin
	event := EventMessage{
		ID:   fmt.Sprintf("skill_%d", time.Now().UnixNano()),
		Type: "event",
		Name: "skill_" + skillName,
		Payload: map[string]any{
			"caller_id": source.ID,
			"payload":   payload,
		},
	}

	encoder := json.NewEncoder(targetPlugin.Stdin)
	if err := encoder.Encode(event); err != nil {
		log.Printf("Failed to inject skill event to %s: %v", targetID, err)
	} else {
		log.Printf("Skill %s called: %s -> %s", skillName, source.ID, targetID)
	}
}

// DispatchEvent routes an incoming event to matching plugins
func (pm *PluginManager) DispatchEvent(event *EventMessage) {
	pm.mutex.Lock()
	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		if p.State == "running" {
			plugins = append(plugins, p)
		}
	}
	pm.mutex.Unlock()

	// 1. Check for Intent matches first if it's a message
	if event.Name == "on_message" {
		payload, ok := event.Payload.(map[string]any)
		if ok {
			text, _ := payload["text"].(string)
			if text != "" {
				pm.routeByIntent(text, event)
			}
		}
	}

	// 2. Broadcast to plugins that explicitly subscribe to this event name
	for _, p := range plugins {
		shouldSend := false
		for _, e := range p.Config.Events {
			if e == event.Name || e == "*" {
				shouldSend = true
				break
			}
		}

		if shouldSend {
			encoder := json.NewEncoder(p.Stdin)
			if err := encoder.Encode(event); err != nil {
				log.Printf("Failed to send event %s to plugin %s: %v", event.Name, p.ID, err)
			}
		}
	}
}

func (pm *PluginManager) routeByIntent(text string, originalEvent *EventMessage) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	type match struct {
		plugin *Plugin
		intent Intent
	}
	var matches []match

	for _, p := range pm.plugins {
		if p.State != "running" {
			continue
		}
		for _, intent := range p.Config.Intents {
			for _, kw := range intent.Keywords {
				if strings.Contains(strings.ToLower(text), strings.ToLower(kw)) {
					matches = append(matches, match{p, intent})
					break
				}
			}
		}
	}

	// Sort matches by priority (simple implementation)
	for _, m := range matches {
		intentEvent := EventMessage{
			ID:            fmt.Sprintf("intent_%d", time.Now().UnixNano()),
			Type:          "event",
			Name:          "intent_" + m.intent.Name,
			CorrelationID: originalEvent.ID,
			Payload: map[string]any{
				"original_text": text,
				"intent_name":   m.intent.Name,
				"source_event":  originalEvent,
			},
		}

		encoder := json.NewEncoder(m.plugin.Stdin)
		if err := encoder.Encode(intentEvent); err != nil {
			log.Printf("Failed to send intent %s to plugin %s: %v", m.intent.Name, m.plugin.ID, err)
		} else {
			log.Printf("Intent matched: '%s' -> Plugin %s (Intent: %s)", text, m.plugin.ID, m.intent.Name)
		}
	}
}

func (pm *PluginManager) StopPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	if plugin.State != "running" {
		return nil
	}

	if err := plugin.Process.Kill(); err != nil {
		return err
	}

	plugin.Process.Wait()
	plugin.State = "stopped"
	plugin.Process = nil
	plugin.Stdin.Close()
	plugin.Stdout.Close()

	return nil
}

func (pm *PluginManager) RestartPlugin(name string) error {
	if err := pm.StopPlugin(name); err != nil {
		return err
	}
	return pm.StartPlugin(name)
}

func (pm *PluginManager) ListPlugins() []*Plugin {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (pm *PluginManager) HotUpdatePlugin(name string, newVersion string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Create new plugin instance with updated version
	newPlugin := &Plugin{
		ID:      plugin.ID,
		Config:  plugin.Config,
		State:   "stopped",
		Version: newVersion,
	}

	// Start new plugin instance
	if err := pm.startPluginInstance(newPlugin); err != nil {
		return err
	}

	// Health check - wait for initial response
	if err := pm.healthCheckPlugin(newPlugin); err != nil {
		newPlugin.Process.Kill()
		return fmt.Errorf("health check failed: %v", err)
	}

	// Switch traffic - stop old plugin
	if plugin.State == "running" {
		if err := plugin.Process.Kill(); err != nil {
			return err
		}
		plugin.Process.Wait()
		plugin.State = "stopped"
		plugin.Process = nil
		plugin.Stdin.Close()
		plugin.Stdout.Close()
	}

	// Replace old plugin with new one
	pm.plugins[name] = newPlugin

	return nil
}

func (pm *PluginManager) startPluginInstance(plugin *Plugin) error {
	parts := strings.Fields(plugin.Config.EntryPoint)
	if len(parts) == 0 {
		return fmt.Errorf("empty entry point for plugin %s", plugin.ID)
	}

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0])
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	plugin.Process = cmd.Process
	plugin.Stdin = stdin.(*os.File)
	plugin.Stdout = stdout.(*os.File)
	plugin.State = "running"
	plugin.RestartCount = 0
	plugin.LastRestart = time.Now()

	go pm.monitorPlugin(plugin)
	go pm.readPluginOutput(plugin)

	return nil
}

func (pm *PluginManager) healthCheckPlugin(plugin *Plugin) error {
	// Send health check event
	healthEvent := EventMessage{
		ID:   "health-check-123",
		Type: "event",
		Name: "on_health_check",
		Payload: map[string]any{
			"timestamp": time.Now().Unix(),
		},
	}

	encoder := json.NewEncoder(plugin.Stdin)
	if err := encoder.Encode(healthEvent); err != nil {
		return err
	}

	// Wait for response with timeout
	responseChan := make(chan bool, 1)
	go func() {
		decoder := json.NewDecoder(plugin.Stdout)
		var resp ResponseMessage
		if err := decoder.Decode(&resp); err == nil && resp.OK {
			responseChan <- true
		}
	}()

	select {
	case <-responseChan:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("health check timeout")
	}
}
