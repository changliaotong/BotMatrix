package core

import (
	log "BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"BotMatrix/common/plugin/policy"
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
		if !pm.isActionAllowed(plugin, action.Type) {
			log.Printf("plugin %s tried to execute forbidden action %s", plugin.ID, action.Type)
			continue
		}

		log.Printf("executing action %s from plugin %s", action.Type, plugin.ID)
	}
}

func (pm *PluginManager) isActionAllowed(plugin *Plugin, actionType string) bool {
	// First check plugin's own allowed actions
	for _, allowed := range plugin.Config.Actions {
		if allowed == actionType {
			// Then check against the appropriate policy whitelist
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
		}
	}
	return false
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
	cmd := exec.Command(plugin.Config.EntryPoint)
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
		Payload: map[string]interface{}{
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
