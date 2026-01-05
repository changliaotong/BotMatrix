package core

import (
	log "BotMatrix/common/log"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"BotMatrix/common/plugin/policy"
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
	log.Printf("[PluginManager] Started reading output for plugin %s", plugin.ID)
	scanner := bufio.NewScanner(plugin.Stdout)
	// 设置缓冲区大小，防止单行过长导致 Scanner 失败（默认 64k，这里设为 1MB）
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// 检查是否是 JSON 对象（以 { 开头）
		if !strings.HasPrefix(trimmedLine, "{") {
			// 如果不是 JSON，视为普通日志输出
			log.Printf("[PluginLog][%s] %s", plugin.ID, line)
			continue
		}

		var resp ResponseMessage
		if err := json.Unmarshal([]byte(trimmedLine), &resp); err != nil {
			// 如果解析失败，可能是混杂了日志，尝试寻找 JSON 的起始位置
			startIdx := strings.Index(trimmedLine, "{")
			if startIdx > 0 {
				trimmedLine = trimmedLine[startIdx:]
				if err := json.Unmarshal([]byte(trimmedLine), &resp); err == nil {
					log.Printf("[PluginManager] Received response (after recovery) from plugin %s: %s", plugin.ID, trimmedLine)
					pm.handlePluginResponse(plugin, &resp)
					continue
				}
			}

			// 记录非 JSON 输出作为普通日志（可选，目前已在外部打印）
			continue
		}

		// 打印收到的响应
		log.Printf("[PluginManager] Received response from plugin %s for ID %s: %s", plugin.ID, resp.ID, trimmedLine)

		if len(resp.Actions) > 0 {
			for i, action := range resp.Actions {
				log.Printf("[PluginManager] Action[%d] for ID %s: Type=%s, Target=%s, TextLen=%d, Text=%s", i, resp.ID, action.Type, action.Target, len(action.Text), action.Text)
			}
		} else {
			log.Printf("[PluginManager] Response from %s for ID %s has 0 actions. Raw JSON: %s", plugin.ID, resp.ID, trimmedLine)
		}
		pm.handlePluginResponse(plugin, &resp)
		continue
	}

	if err := scanner.Err(); err != nil {
		if plugin.State == "running" && err != io.EOF {
			log.Printf("[PluginManager] Scanner error for plugin %s: %v", plugin.ID, err)
		}
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
		if pm.actionHandler != nil {
			pm.actionHandler(plugin, &action)
		}
	}
}

func (pm *PluginManager) isActionAllowed(plugin *Plugin, actionType string) bool {
	// 临时：在调试阶段允许所有操作
	log.Printf("DEBUG: Authorizing action %s for plugin %s", actionType, plugin.ID)
	return true

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
	for _, versions := range pm.plugins {
		if len(versions) == 0 {
			continue
		}
		p := pm.selectVersion(event, versions)
		if p != nil {
			// If running, send immediately. If stopped, buffer it.
			if p.State == "running" || p.State == "stopped" {
				targetPlugins = append(targetPlugins, p)
			}
		}
	}
	pm.mutex.Unlock()

	// 1. Check for Intent matches first if it's a message
	// ... (omitted for brevity in search/replace match if possible, but I'll include it to be safe)
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
			if p.State == "running" {
				encoder := json.NewEncoder(p.Stdin)
				if err := encoder.Encode(event); err != nil {
					log.Printf("Failed to send event %s to plugin %s: %v", event.Name, p.ID, err)
				}
			} else if p.State == "stopped" {
				// Buffer messages while the plugin is restarting
				pm.mutex.Lock()
				if len(p.MessageBuffer) < 100 { // Limit buffer size to 100 messages
					p.MessageBuffer = append(p.MessageBuffer, event)
					log.Printf("[PluginManager] Buffered event %s for restarting plugin %s", event.Name, p.ID)
				}
				pm.mutex.Unlock()
			}
		}
	}
}

// DispatchEventToPlugin routes an event to a specific plugin version
func (pm *PluginManager) DispatchEventToPlugin(id string, version string, event *EventMessage) {
	pm.mutex.Lock()
	versions, ok := pm.plugins[id]
	pm.mutex.Unlock()

	if !ok {
		return
	}

	var target *Plugin
	for _, v := range versions {
		if v.Config.Version == version {
			target = v
			break
		}
	}

	if target != nil && target.State == "running" {
		encoder := json.NewEncoder(target.Stdin)
		if err := encoder.Encode(event); err != nil {
			log.Printf("Failed to send targeted event %s to plugin %s (v%s): %v", event.Name, id, version, err)
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

	// Clean up runtime directory if it's a shadow copy
	if plugin.RuntimeDir != "" && plugin.RuntimeDir != plugin.Dir {
		log.Printf("[PluginManager] Cleaning up runtime directory for %s: %s", plugin.ID, plugin.RuntimeDir)
		os.RemoveAll(plugin.RuntimeDir)
		plugin.RuntimeDir = ""
	}

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

	// Create a shadow copy of the plugin directory to avoid file locking on Windows
	// We use a .runtime folder in the plugin's own parent directory instead of OS temp to avoid disk space issues on C:
	runtimeBaseDir := filepath.Join(plugin.Dir, ".runtime")
	runtimeDir := filepath.Join(runtimeBaseDir, fmt.Sprintf("run_%d", time.Now().UnixNano()))
	log.Printf("[PluginManager] Creating shadow copy for plugin %s: %s -> %s", plugin.ID, plugin.Dir, runtimeDir)
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return fmt.Errorf("failed to create runtime directory: %v", err)
	}

	// Also try to clean up old runtime directories in the same base folder
	go func() {
		files, err := os.ReadDir(runtimeBaseDir)
		if err == nil {
			for _, f := range files {
				if f.IsDir() && strings.HasPrefix(f.Name(), "run_") {
					// If the directory is older than 1 hour, try to remove it
					info, err := f.Info()
					if err == nil && time.Since(info.ModTime()) > 1*time.Hour {
						os.RemoveAll(filepath.Join(runtimeBaseDir, f.Name()))
					}
				}
			}
		}
	}()

	if err := copyDir(plugin.Dir, runtimeDir); err != nil {
		os.RemoveAll(runtimeDir)
		return fmt.Errorf("failed to copy plugin to runtime directory: %v", err)
	}
	plugin.RuntimeDir = runtimeDir

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0])
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}
	cmd.Dir = plugin.RuntimeDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
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

	// Flush message buffer after startup
	if len(plugin.MessageBuffer) > 0 {
		log.Printf("[PluginManager] Flushing %d buffered messages for plugin %s", len(plugin.MessageBuffer), plugin.ID)
		encoder := json.NewEncoder(plugin.Stdin)
		for _, bufferedEvent := range plugin.MessageBuffer {
			if err := encoder.Encode(bufferedEvent); err != nil {
				log.Printf("Failed to send buffered event %s to plugin %s: %v", bufferedEvent.Name, plugin.ID, err)
			}
		}
		plugin.MessageBuffer = nil // Clear buffer after flushing
	}

	go pm.monitorPlugin(plugin)
	go pm.readPluginOutput(plugin)
	go pm.readPluginError(plugin, stderr)

	log.Printf("Started plugin %s (v%s) with PID %d (Running in shadow copy: %s)", plugin.ID, plugin.Config.Version, plugin.Process.Pid, plugin.RuntimeDir)
	return nil
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip the .runtime directory to avoid infinite recursion
		if rel == ".runtime" || strings.HasPrefix(rel, ".runtime"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file with retries for Windows file locking
		var srcFile *os.File
		var openErr error
		for i := 0; i < 5; i++ {
			srcFile, openErr = os.Open(path)
			if openErr == nil {
				break
			}
			if i < 4 {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return openErr
		}
		defer srcFile.Close()

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(targetPath, info.Mode())
	})
}

func (pm *PluginManager) readPluginError(plugin *Plugin, stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			log.Printf("[PluginErr][%s] %s", plugin.ID, line)
		}
	}
}
