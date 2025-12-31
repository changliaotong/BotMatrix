package session

import (
	"context"
	"time"
)

// SessionStore defines the interface for distributed session management.
type SessionStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// SessionKey generates a standard session key for a user in a group.
func SessionKey(pluginID, groupId, userId string) string {
	return "session:" + pluginID + ":" + groupId + ":" + userId
}

// CorrelationKey generates a key for tracking distributed Ask interactions.
func CorrelationKey(correlationID string) string {
	return "correlation:" + correlationID
}
