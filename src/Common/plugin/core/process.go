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

func (pm *PluginManager) StartPlugin(id string, version string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	versions, ok := pm.plugins[id]
	if !ok {
		return fmt.Errorf("plugin %s not found", id)
	}

	var plugin *Plugin
	if version == "" {
		plugin = versions[len(versions)-1]
	} else {
		for _, p := range versions {
			if p.Config.Version == version {
				plugin = p
				break
			}
		}
	}

	if plugin == nil {
		return fmt.Errorf("plugin %s version %s not found", id, version)
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
	pm.StartPlugin(plugin.ID, plugin.Config.Version)
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
	versions, exists := pm.plugins[targetID]
	pm.mutex.Unlock()

	if !exists || len(versions) == 0 {
		log.Printf("Skill call failed: target plugin %s not found", targetID)
		return
	}

	// Always route to the latest version for skills for now
	targetPlugin := versions[len(versions)-1]
	if targetPlugin.State != "running" {
		log.Printf("Skill call failed: target plugin %s (v%s) is not running", targetID, targetPlugin.Config.Version)
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
	// Get candidate plugins with version selection (Canary)
	targetPlugins := make([]*Plugin, 0)
	for id, versions := range pm.plugins {
		if len(versions) == 0 {
			continue
		}
		p := pm.selectVersion(event, versions)
		if p != nil && p.State == "running" {
			targetPlugins = append(targetPlugins, p)
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
	for _, p := range targetPlugins {
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

// selectVersion implements Canary routing logic
func (pm *PluginManager) selectVersion(event *EventMessage, versions []*Plugin) *Plugin {
	if len(versions) == 1 {
		return versions[0]
	}

	stable := versions[len(versions)-1]
	var canary *Plugin

	for _, v := range versions {
		if v.Config.CanaryWeight > 0 {
			canary = v
			break
		}
	}

	if canary == nil {
		return stable
	}

	seed := event.CorrelationId
	if seed == "" {
		if payload, ok := event.Payload.(map[string]any); ok {
			if from, ok := payload["from"].(string); ok {
				seed = from
			}
		}
	}

	if seed == "" {
		seed = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	h := 0
	for i := 0; i < len(seed); i++ {
		h = 31*h + int(seed[i])
	}
	if h < 0 {
		h = -h
	}

	if h%100 < canary.Config.CanaryWeight {
		return canary
	}
	return stable
}

func (pm *PluginManager) routeByIntent(text string, originalEvent *EventMessage) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	type match struct {
		plugin *Plugin
		intent Intent
	}
	var matches []match

	for _, versions := range pm.plugins {
		if len(versions) == 0 {
			continue
		}
		p := pm.selectVersion(originalEvent, versions)
		if p == nil || p.State != "running" {
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

	for _, m := range matches {
		intentEvent := EventMessage{
			ID:            fmt.Sprintf("intent_%d", time.Now().UnixNano()),
			Type:          "event",
			Name:          "intent_" + m.intent.Name,
			CorrelationId: originalEvent.ID,
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

func (pm *PluginManager) StopPlugin(id string, version string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	versions, ok := pm.plugins[id]
	if !ok {
		return fmt.Errorf("plugin %s not found", id)
	}

	var plugin *Plugin
	for _, p := range versions {
		if version == "" || p.Config.Version == version {
			plugin = p
			break
		}
	}

	if plugin == nil {
		return fmt.Errorf("plugin %s version %s not found", id, version)
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

func (pm *PluginManager) RestartPlugin(id string, version string) error {
	if err := pm.StopPlugin(id, version); err != nil {
		return err
	}
	return pm.StartPlugin(id, version)
}

func (pm *PluginManager) HotUpdatePlugin(id string, newVersion string) error {
	pm.mutex.Lock()
	versions, ok := pm.plugins[id]
	if !ok {
		pm.mutex.Unlock()
		return fmt.Errorf("plugin %s not found", id)
	}
	pm.mutex.Unlock()

	if err := pm.StartPlugin(id, newVersion); err != nil {
		return fmt.Errorf("failed to start new version: %v", err)
	}

	for _, p := range versions {
		if p.Config.Version != newVersion {
			pm.StopPlugin(id, p.Config.Version)
		}
	}

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
	plugin.LastRestart = time.Now()

	go pm.monitorPlugin(plugin)
	go pm.readPluginOutput(plugin)

	log.Printf("Started plugin %s (v%s) with PID %d", plugin.ID, plugin.Config.Version, plugin.Process.Pid)
	return nil
}
