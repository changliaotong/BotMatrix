package bot

import (
	"fmt"
	"time"

	"BotMatrix/common/types"
)

// BroadcastEvent broadcasts an event to all subscribers
func (m *Manager) BroadcastEvent(event any) {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for _, sub := range m.Subscribers {
		sub.Mutex.Lock()
		err := sub.Conn.WriteJSON(event)
		sub.Mutex.Unlock()
		if err != nil {
			fmt.Printf("Failed to send event to subscriber: %v\n", err)
		}
	}
}

// BroadcastDockerEvent broadcasts a Docker event to all subscribers
func (m *Manager) BroadcastDockerEvent(action, containerID, status string) {
	event := types.DockerEvent{
		Type:        "docker_event",
		Action:      action,
		ContainerID: containerID,
		Status:      status,
		Timestamp:   time.Now(),
	}

	m.BroadcastEvent(event)
}

// BroadcastEventStatic broadcasts an event to a given map of subscribers
func BroadcastEventStatic(subscribers map[any]*types.Subscriber, event any) {
	for _, sub := range subscribers {
		sub.Mutex.Lock()
		err := sub.Conn.WriteJSON(event)
		sub.Mutex.Unlock()
		if err != nil {
			fmt.Printf("Failed to send event to subscriber: %v\n", err)
		}
	}
}

// BroadcastDockerEventStatic broadcasts a Docker event to a given map of subscribers
func BroadcastDockerEventStatic(subscribers map[any]*types.Subscriber, action, containerID, status string) {
	event := types.DockerEvent{
		Type:        "docker_event",
		Action:      action,
		ContainerID: containerID,
		Status:      status,
		Timestamp:   time.Now(),
	}

	BroadcastEventStatic(subscribers, event)
}
