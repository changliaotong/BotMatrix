package common

import (
	"log"
	"time"
)

// BroadcastDockerEvent 向所有订阅者广播 Docker 事件
func (m *Manager) BroadcastDockerEvent(action, containerID, status string) {
	event := DockerEvent{
		Type:        "docker_event",
		Action:      action,
		ContainerID: containerID,
		Status:      status,
		Timestamp:   time.Now(),
	}

	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for _, sub := range m.Subscribers {
		go func(s *Subscriber, e DockerEvent) {
			s.Mutex.Lock()
			defer s.Mutex.Unlock()
			err := s.Conn.WriteJSON(e)
			if err != nil {
				log.Printf("[SUBSCRIBER] Failed to send docker event to subscriber: %v", err)
			}
		}(sub, event)
	}
}
