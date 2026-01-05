package bot

import (
	"net/http"

	"BotMatrix/common/middleware"
)

// JWTMiddleware validates the JWT token in the Authorization header
func (m *Manager) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	mw := middleware.JWTMiddleware(
		m.ValidateToken,
		m.GetOrLoadUser,
	)
	return mw(next)
}

// AdminMiddleware ensures the user has admin privileges
func (m *Manager) AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	mw := middleware.AdminMiddleware(m.JWTMiddleware)
	return mw(next)
}
