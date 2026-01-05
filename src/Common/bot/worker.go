package bot

import (
	"time"

	"BotMatrix/common/types"
)

// StartWorkerTimeoutDetection starts worker timeout detection
func (m *Manager) StartWorkerTimeoutDetection() {
	m.LogInfo("[Worker Monitor] Starting timeout detection (60s interval)")
	for range NewTicker(60) {
		m.checkWorkerTimeouts()
	}
}

// checkWorkerTimeouts checks for worker heartbeat timeouts
func (m *Manager) checkWorkerTimeouts() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.Workers = CheckWorkerTimeouts(
		m.Workers,
		m.LogWarn,
		m.LogInfo,
		m.TrackWorkerDisconnection,
	)
}

// ForwardToWorkerWithLog forwards data to a worker with logging
func (m *Manager) ForwardToWorkerWithLog(data any, targetWorker *types.WorkerClient) error {
	targetWorker.Mutex.Lock()
	err := targetWorker.Conn.WriteJSON(data)
	targetWorker.HandledCount++
	targetWorker.Mutex.Unlock()

	if err != nil {
		m.LogError("[Worker Forward] Failed to send to worker %s: %v", targetWorker.ID, err)
		return err
	}
	return nil
}

// FallbackToRoundRobinWithLog falls back to round-robin worker selection with logging
func (m *Manager) FallbackToRoundRobinWithLog(data any) {
	if len(m.Workers) == 0 {
		m.LogWarn("[Fallback] No workers available for fallback")
		return
	}

	targetIndex := int(time.Now().UnixNano()) % len(m.Workers)
	worker := m.Workers[targetIndex]

	worker.Mutex.Lock()
	err := worker.Conn.WriteJSON(data)
	worker.HandledCount++
	worker.Mutex.Unlock()

	if err != nil {
		m.LogError("[Fallback] Failed to send to worker %s: %v", worker.ID, err)
	}
}

// CheckWorkerTimeouts checks for worker heartbeat timeouts
func CheckWorkerTimeouts(
	workers []*types.WorkerClient,
	logWarn func(string, ...any),
	logInfo func(string, ...any),
	trackWorkerDisconnection func(string, string, time.Duration),
) []*types.WorkerClient {
	now := time.Now()
	activeWorkers := make([]*types.WorkerClient, 0)

	for _, worker := range workers {
		worker.Mutex.Lock()
		lastHeartbeat := worker.LastHeartbeat
		if lastHeartbeat.IsZero() {
			lastHeartbeat = worker.Connected
		}

		timeoutDuration := now.Sub(lastHeartbeat)

		if timeoutDuration < 2*time.Minute {
			activeWorkers = append(activeWorkers, worker)
		} else {
			logWarn("[Worker Monitor] Worker %s heartbeat timeout after %v, closing connection", worker.ID, timeoutDuration)
			if worker.Conn != nil {
				worker.Conn.Close()
			}
			trackWorkerDisconnection(worker.ID, "heartbeat_timeout", timeoutDuration)
		}
		worker.Mutex.Unlock()
	}

	removedCount := len(workers) - len(activeWorkers)
	if removedCount > 0 {
		logInfo("[Worker Monitor] Removed %d timeout workers, remaining: %d", removedCount, len(activeWorkers))
	}

	return activeWorkers
}
