package common

import (
	"time"
)

// 启动Worker超时检测
func (m *Manager) StartWorkerTimeoutDetection() {
	m.LogInfo("[Worker Monitor] Starting timeout detection (60s interval)")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.checkWorkerTimeouts()
	}
}

// 检查Worker心跳超时
func (m *Manager) checkWorkerTimeouts() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	now := time.Now()
	activeWorkers := make([]*WorkerClient, 0)

	if len(m.Workers) > 0 {
		// m.LogDebug("[Worker Monitor] Checking timeouts for %d workers", len(m.Workers))
	}

	for _, worker := range m.Workers {
		worker.Mutex.Lock()
		lastHeartbeat := worker.LastHeartbeat
		if lastHeartbeat.IsZero() {
			lastHeartbeat = worker.Connected
		}

		timeoutDuration := now.Sub(lastHeartbeat)

		if timeoutDuration < 2*time.Minute {
			activeWorkers = append(activeWorkers, worker)
		} else {
			// 超时Worker，关闭连接
			m.LogWarn("[Worker Monitor] Worker %s heartbeat timeout after %v, closing connection",
				worker.ID, timeoutDuration)
			worker.Conn.Close()

			// 记录断开连接
			m.TrackWorkerDisconnection(worker.ID, "heartbeat_timeout", timeoutDuration)
		}
		worker.Mutex.Unlock()
	}

	removedCount := len(m.Workers) - len(activeWorkers)
	if removedCount > 0 {
		m.LogInfo("[Worker Monitor] Removed %d timeout workers, remaining: %d", removedCount, len(activeWorkers))
	}

	m.Workers = activeWorkers
}

// Worker消息转发（带详细日志）
func (m *Manager) forwardToWorkerWithLog(data any, targetWorker *WorkerClient) error {
	// m.LogDebug("[Worker Forward] Attempting to send to worker %s: %v", targetWorker.ID, data)

	targetWorker.Mutex.Lock()
	err := targetWorker.Conn.WriteJSON(data)
	targetWorker.HandledCount++
	targetWorker.Mutex.Unlock()

	if err != nil {
		m.LogError("[Worker Forward] Failed to send to worker %s: %v", targetWorker.ID, err)
		return err
	}

	// m.LogDebug("[Worker Forward] Successfully sent to worker %s", targetWorker.ID)
	return nil
}

// 带日志的fallback函数
func (m *Manager) fallbackToRoundRobinWithLog(data any) {
	m.LogDebug("[Fallback] Starting fallbackToRoundRobin with %d workers", len(m.Workers))

	if len(m.Workers) == 0 {
		m.LogWarn("[Fallback] No workers available for fallback")
		return
	}

	targetIndex := int(time.Now().UnixNano()) % len(m.Workers)
	worker := m.Workers[targetIndex]

	m.LogDebug("[Fallback] Trying worker %s at index %d", worker.ID, targetIndex)

	worker.Mutex.Lock()
	err := worker.Conn.WriteJSON(data)
	worker.HandledCount++
	worker.Mutex.Unlock()

	if err != nil {
		m.LogError("[Fallback] Worker %s failed: %v", worker.ID, err)
		go func(w *WorkerClient) {
			m.LogInfo("[Fallback] Removing failed worker %s", w.ID)
			m.removeWorkerWithLog(w)
		}(worker)

		// 尝试其他worker
		m.LogInfo("[Fallback] Trying other workers after %s failed", worker.ID)
		success := false
		for i, w := range m.Workers {
			if i == targetIndex {
				continue
			}
			m.LogDebug("[Fallback] Trying alternative worker %s", w.ID)
			w.Mutex.Lock()
			e := w.Conn.WriteJSON(data)
			w.Mutex.Unlock()
			if e == nil {
				m.LogInfo("[Fallback] Successfully used alternative worker %s", w.ID)
				success = true
				break
			} else {
				m.LogError("[Fallback] Alternative worker %s also failed: %v", w.ID, e)
			}
		}

		if !success {
			m.LogError("[Fallback] All workers failed, message dropped: %v", data)
		}
	} else {
		m.LogInfo("[Fallback] Successfully used worker %s", worker.ID)
	}

	m.LogDebug("[Fallback] Completed fallbackToRoundRobin")
}

// 带日志的removeWorker
func (m *Manager) removeWorkerWithLog(worker *WorkerClient) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.LogInfo("[Worker Remove] Removing worker %s from active list", worker.ID)

	newWorkers := make([]*WorkerClient, 0, len(m.Workers))
	for _, w := range m.Workers {
		if w != worker {
			newWorkers = append(newWorkers, w)
		}
	}

	removedCount := len(m.Workers) - len(newWorkers)
	if removedCount > 0 {
		m.LogInfo("[Worker Remove] Removed %d worker(s), remaining: %d", removedCount, len(newWorkers))
	}

	m.Workers = newWorkers
}

// findWorkerByID 根据ID查找Worker
func (m *Manager) FindWorkerByID(workerID string) *WorkerClient {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	for _, w := range m.Workers {
		if w.ID == workerID {
			return w
		}
	}
	return nil
}
