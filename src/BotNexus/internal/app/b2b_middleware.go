package app

import (
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"net/http"
	"strings"
)

const B2BClaimsKey types.ContextKey = "b2b_claims"

// B2BMiddleware 允许标准 JWT 或 B2B 联邦身份认证
func (m *Manager) B2BMiddleware(next http.HandlerFunc) http.HandlerFunc {
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

		// 1. 尝试标准 JWT 认证
		claims, err := m.ValidateToken(tokenString)
		if err == nil {
			// 标准用户认证成功
			ctx := context.WithValue(r.Context(), types.UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 2. 尝试 B2B 联邦身份认证
		if m.B2BService != nil {
			ent, err := m.B2BService.VerifyB2BToken(tokenString)
			if err == nil {
				// B2B 认证成功
				ctx := context.WithValue(r.Context(), B2BClaimsKey, ent)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		utils.SendJSONResponse(w, false, utils.T(lang, "invalid_token|无效或已过期的Token"), nil)
	}
}
