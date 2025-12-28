package common

import (
	"context"
	"time"
)

// SessionStore defines the interface for distributed session management.
// This allows BotWorker instances to be stateless while maintaining user context.
type SessionStore interface {
	// Set stores a value for a given key with an optional expiration.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get retrieves a value for a given key.
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete removes a value for a given key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the store.
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
