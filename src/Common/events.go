package common

import (
	"fmt"
	"time"
)

// BroadcastEvent broadcasts an event to all subscribers
func (m *Manager) BroadcastEvent(event interface{}) {
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

// BroadcastDockerEvent 向所有订阅者广播 Docker 事件
func (m *Manager) BroadcastDockerEvent(action, containerID, status string) {
	event := DockerEvent{
		Type:        "docker_event",
		Action:      action,
		ContainerID: containerID,
		Status:      status,
		Timestamp:   time.Now(),
	}

	m.BroadcastEvent(event)
}
