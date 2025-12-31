package types

import (
	"sync"
	"time"
)

// ConnectionStats tracks connection lifecycle statistics
type ConnectionStats struct {
	TotalBotConnections       int64                    `json:"total_bot_connections"`
	TotalWorkerConnections    int64                    `json:"total_worker_connections"`
	BotConnectionDurations    map[string]time.Duration `json:"bot_connection_durations"`    // bot_id -> duration
	WorkerConnectionDurations map[string]time.Duration `json:"worker_connection_durations"` // worker_id -> duration
	BotDisconnectReasons      map[string]int64         `json:"bot_disconnect_reasons"`      // reason -> count
	WorkerDisconnectReasons   map[string]int64         `json:"worker_disconnect_reasons"`   // reason -> count
	LastBotActivity           map[string]time.Time     `json:"last_bot_activity"`           // bot_id -> last activity
	LastWorkerActivity        map[string]time.Time     `json:"last_worker_activity"`        // worker_id -> last activity
	Mutex                     sync.RWMutex
}

// BotStatDetail represents detailed stats for a bot
type BotStatDetail struct {
	Sent     int64            `json:"sent"`
	Received int64            `json:"received"`
	Users    map[string]int64 `json:"users"`  // UserID -> Count
	Groups   map[string]int64 `json:"groups"` // GroupID -> Count
	LastMsg  time.Time        `json:"last_msg"`
}

type ProcInfo struct {
	Pid    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"`
}
