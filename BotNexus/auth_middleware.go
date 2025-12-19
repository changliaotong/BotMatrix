package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// AuthKey is the key for the user claims in the request context
type AuthKey string

const (
	UserClaimsKey AuthKey = "user_claims"
)

// JWTMiddleware validates the JWT token in the Authorization header
func (m *Manager) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Check for token in query parameter for WebSocket or simple GET
			token := r.URL.Query().Get("token")
			if token != "" {
				authHeader = "Bearer " + token
			}
		}

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"success": false, "message": "Missing or invalid authorization header"}`)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenString)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"success": false, "message": "Invalid or expired token"}`)
			return
		}

		// 验证 SessionVersion 以支持多端同步和强制登出
		m.usersMutex.RLock()
		user, exists := m.users[claims.Username]
		m.usersMutex.RUnlock()

		// 如果内存中不存在，尝试从数据库加载
		if !exists {
			row := m.db.QueryRow("SELECT id, username, password_hash, is_admin, session_version, created_at, updated_at FROM users WHERE username = ?", claims.Username)
			var u User
			var createdAt, updatedAt string
			err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.SessionVersion, &createdAt, &updatedAt)
			if err == nil {
				user = &u
				exists = true

				// 更新内存缓存
				m.usersMutex.Lock()
				m.users[u.Username] = &u
				m.usersMutex.Unlock()
			}
		}

		if !exists || user.SessionVersion != claims.SessionVersion {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"success": false, "message": "Session expired or invalidated"}`)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminMiddleware ensures the user has admin privileges
func (m *Manager) AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return m.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
		if !ok || !claims.IsAdmin {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, `{"success": false, "message": "Admin privileges required"}`)
			return
		}
		next.ServeHTTP(w, r)
	})
}
