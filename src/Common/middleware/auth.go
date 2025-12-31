package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"BotMatrix/common/types"
	"BotMatrix/common/utils"
)

// ValidateTokenFunc is a function type for validating tokens
type ValidateTokenFunc func(token string) (*types.UserClaims, error)

// GetUserFunc is a function type for retrieving a user by username
type GetUserFunc func(username string) (*types.User, bool)

// JWTMiddleware validates the JWT token in the Authorization header
func JWTMiddleware(validateToken ValidateTokenFunc, getUser GetUserFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			lang := utils.GetLangFromRequest(r)
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				token := r.URL.Query().Get("token")
				if token != "" {
					authHeader = "Bearer " + token
				}
			}

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				utils.SendJSONResponse(w, false, utils.T(lang, "missing_auth_header|未提供认证信息"), nil)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateToken(tokenString)
			if err != nil {
				log.Printf("Token validation failed for %s: %v", r.URL.Path, err)
				w.WriteHeader(http.StatusUnauthorized)
				utils.SendJSONResponse(w, false, utils.T(lang, "invalid_token|无效或已过期的Token"), nil)
				return
			}

			user, exists := getUser(claims.Username)
			if !exists {
				log.Printf("[Auth] User %s not found in memory/DB during token validation for %s", claims.Username, r.URL.Path)
				w.WriteHeader(http.StatusUnauthorized)
				utils.SendJSONResponse(w, false, utils.T(lang, "session_expired|会话已过期或已被禁用"), nil)
				return
			}

			if user.SessionVersion != claims.SessionVersion {
				log.Printf("[Auth] Session version mismatch for user %s: claims version %d, current version %d (Path: %s)", claims.Username, claims.SessionVersion, user.SessionVersion, r.URL.Path)
				w.WriteHeader(http.StatusUnauthorized)
				utils.SendJSONResponse(w, false, utils.T(lang, "session_expired|会话已过期或已被禁用"), nil)
				return
			}

			ctx := context.WithValue(r.Context(), types.UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware(jwtMiddleware func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
			lang := utils.GetLangFromRequest(r)
			claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
			if !ok || !claims.IsAdmin {
				w.WriteHeader(http.StatusForbidden)
				utils.SendJSONResponse(w, false, utils.T(lang, "admin_required|需要管理员权限"), nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
