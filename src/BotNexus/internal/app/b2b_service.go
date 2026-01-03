package app

import (
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// B2BServiceImpl 实现 B2BService 接口
type B2BServiceImpl struct {
	db      *gorm.DB
	manager *Manager
	client  *http.Client
}

func NewB2BService(db *gorm.DB, m *Manager) *B2BServiceImpl {
	return &B2BServiceImpl{
		db:      db,
		manager: m,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Connect 建立企业间 B2B 连接
func (s *B2BServiceImpl) Connect(sourceEntCode, targetEntCode string) error {
	var sourceEnt models.EnterpriseGORM
	if err := s.db.Where("code = ?", sourceEntCode).First(&sourceEnt).Error; err != nil {
		return fmt.Errorf("source enterprise not found: %w", err)
	}

	// 在实际场景中，这里应该通过网络请求向 targetEnt 发起握手请求
	// 目前简化为直接在本地数据库创建连接记录 (假设是单机模拟或通过共享库)
	var targetEnt models.EnterpriseGORM
	if err := s.db.Where("code = ?", targetEntCode).First(&targetEnt).Error; err != nil {
		return fmt.Errorf("target enterprise not found: %w", err)
	}

	conn := models.B2BConnectionGORM{
		SourceEntID:  sourceEnt.ID,
		TargetEntID:  targetEnt.ID,
		Status:       "active",
		AuthProtocol: "jwt",
	}

	return s.db.Create(&conn).Error
}

// SendCrossEnterpriseMessage 发送跨企业数字员工消息
func (s *B2BServiceImpl) SendCrossEnterpriseMessage(fromEmployeeID, toEmployeeID string, msg string) error {
	// 1. 获取发送者员工信息
	var fromEmp models.DigitalEmployeeGORM
	if err := s.db.Where("employee_id = ?", fromEmployeeID).First(&fromEmp).Error; err != nil {
		return err
	}

	// 2. 获取接收者员工信息
	var toEmp models.DigitalEmployeeGORM
	if err := s.db.Where("employee_id = ?", toEmployeeID).First(&toEmp).Error; err != nil {
		return err
	}

	// 3. 检查企业间连接状态
	var conn models.B2BConnectionGORM
	if err := s.db.Where("source_ent_id = ? AND target_ent_id = ? AND status = ?", fromEmp.EnterpriseID, toEmp.EnterpriseID, "active").First(&conn).Error; err != nil {
		return fmt.Errorf("no active B2B connection between enterprises: %w", err)
	}

	// 4. 调用远程企业的 im_send_message 工具 (假设目标企业暴露了此 MCP 工具)
	// 在 Global Agent Mesh 中，数字员工的通信被抽象为 MCP 工具调用
	args := map[string]any{
		"platform":  "mesh", // 特殊平台标识
		"target_id": toEmployeeID,
		"content":   msg,
		"from_id":   fromEmployeeID,
	}

	_, err := s.CallRemoteTool(fromEmp.EnterpriseID, toEmp.EnterpriseID, "im_send_message", args)
	if err != nil {
		return fmt.Errorf("failed to send cross-enterprise message via mesh: %w", err)
	}

	clog.Info(fmt.Sprintf("[B2B] Message sent from %s to %s via Mesh", fromEmployeeID, toEmployeeID))
	return nil
}

func (s *B2BServiceImpl) CallRemoteTool(sourceEntID, targetEntID uint, toolName string, arguments map[string]any) (any, error) {
	// 1. 获取目标企业的公共 MCP 端点
	var apiServer models.MCPServerGORM
	if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", targetEntID, "global", "active").First(&apiServer).Error; err != nil {
		return nil, fmt.Errorf("target enterprise has no public MCP endpoint: %w", err)
	}

	// 2. 构造远程调用 URL
	callURL := apiServer.Endpoint
	if !strings.Contains(callURL, "/api/mcp/v1/tools/call") {
		callURL = strings.TrimSuffix(callURL, "/") + "/api/mcp/v1/tools/call"
	}

	// 3. 准备请求体
	reqBody := map[string]any{
		"server_id": "mesh_bridge",
		"tool_name": toolName,
		"arguments": arguments,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 4. 生成认证 Token
	token, err := s.generateB2BToken(sourceEntID, targetEntID)
	if err != nil {
		return nil, err
	}

	// 5. 发起 HTTP 请求
	req, _ := http.NewRequest("POST", callURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote peer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errRes struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errRes)
		return nil, fmt.Errorf("remote error (status %d): %s", resp.StatusCode, errRes.Message)
	}

	var result struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode remote response: %w", err)
	}

	return result.Data, nil
}

// VerifyIdentity 验证企业身份
func (s *B2BServiceImpl) VerifyIdentity(entCode string, signature string) bool {
	var ent models.EnterpriseGORM
	if err := s.db.Where("code = ?", entCode).First(&ent).Error; err != nil {
		return false
	}

	// 验证 JWT 签名 (使用企业的公钥)
	token, err := jwt.Parse(signature, func(token *jwt.Token) (interface{}, error) {
		return []byte(ent.PublicKey), nil
	})

	return err == nil && token.Valid
}

// VerifyB2BToken 验证 B2B JWT 令牌并返回企业信息
func (s *B2BServiceImpl) VerifyB2BToken(tokenString string) (*models.EnterpriseGORM, error) {
	// 1. 解析不带签名的 token 以获取 issuer (企业代码)
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse unverified token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	entCode, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("missing issuer (iss) in token")
	}

	// 2. 从数据库查找企业及其公钥
	var ent models.EnterpriseGORM
	if err := s.db.Where("code = ?", entCode).First(&ent).Error; err != nil {
		return nil, fmt.Errorf("enterprise %s not found: %w", entCode, err)
	}

	// 3. 使用企业公钥验证签名
	verifiedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ent.PublicKey), nil
	})

	if err != nil || !verifiedToken.Valid {
		return nil, fmt.Errorf("invalid token signature: %w", err)
	}

	return &ent, nil
}

// RegisterEndpoint 注册企业公开的 MCP 端点
func (s *B2BServiceImpl) RegisterEndpoint(entID uint, name, endpointType, url string) error {
	server := models.MCPServerGORM{
		Name:     name,
		Type:     endpointType,
		Endpoint: url,
		Scope:    "global", // 公开端点设为 global
		OwnerID:  entID,
		Status:   "active",
	}
	return s.db.Create(&server).Error
}

// DiscoverEndpoints 发现公开的 MCP 端点
func (s *B2BServiceImpl) DiscoverEndpoints(query string) ([]models.MCPServerGORM, error) {
	var servers []models.MCPServerGORM
	db := s.db.Where("scope = ? AND status = ?", "global", "active")
	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}
	if err := db.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

// DiscoverMeshEndpoints 在全网（Mesh）范围内发现端点
func (s *B2BServiceImpl) DiscoverMeshEndpoints(query string) ([]models.MCPServerGORM, error) {
	// 1. 获取本地端点
	localServers, err := s.DiscoverEndpoints(query)
	if err != nil {
		clog.Error("[Mesh] Local discovery failed", zap.Error(err))
	}

	allServers := localServers

	// 2. 获取所有活跃的 B2B 连接
	var connections []models.B2BConnectionGORM
	if err := s.db.Where("status = ?", "active").Find(&connections).Error; err != nil {
		return allServers, nil // 如果获取连接失败，仅返回本地结果
	}

	// 3. 并发向已连接的企业发起搜索请求
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.B2BConnectionGORM) {
			defer wg.Done()

			// 获取目标企业信息以得到其 API 端点（这里简化逻辑：假设目标企业也有一个公开的 MCP 端点作为 API 入口）
			var targetEnt models.EnterpriseGORM
			if err := s.db.First(&targetEnt, c.TargetEntID).Error; err != nil {
				return
			}

			// 查找目标企业的 API 端点 (类型为 'mesh' 或 'mcp')
			var apiServer models.MCPServerGORM
			if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", c.TargetEntID, "global", "active").First(&apiServer).Error; err != nil {
				return
			}

			// 构造联邦搜索请求 URL (GET /api/mesh/discover?q=...)
			// 注意：这里调用的是对方的 Mesh Discover 接口
			discoverURL := apiServer.Endpoint
			if !strings.Contains(discoverURL, "/api/mesh/discover") {
				// 简单的 URL 拼接
				discoverURL = strings.TrimSuffix(discoverURL, "/") + "/api/mesh/discover"
			}
			discoverURL += "?q=" + query

			// 生成认证 Token
			token, _ := s.generateB2BToken(c.SourceEntID, c.TargetEntID)

			req, _ := http.NewRequest("GET", discoverURL, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := s.client.Do(req)
			if err != nil {
				clog.Warn(fmt.Sprintf("[Mesh] Failed to query remote peer %d: %v", c.TargetEntID, err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var result struct {
					Success bool                   `json:"success"`
					Data    []models.MCPServerGORM `json:"data"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Success {
					mu.Lock()
					// 避免重复并标记来源
					for _, srv := range result.Data {
						srv.Description = fmt.Sprintf("[Remote:%s] %s", targetEnt.Name, srv.Description)
						allServers = append(allServers, srv)
					}
					mu.Unlock()
				}
			}
		}(conn)
	}

	wg.Wait()
	return allServers, nil
}

// generateB2BToken 生成基于 JWT 的跨域令牌
func (s *B2BServiceImpl) generateB2BToken(sourceID, targetID uint) (string, error) {
	var sourceEnt models.EnterpriseGORM
	if err := s.db.First(&sourceEnt, sourceID).Error; err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"iss": sourceEnt.Code,
		"sub": "b2b_communication",
		"aud": fmt.Sprintf("ent_%d", targetID),
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	// 优先使用私钥签名，如果没有则回退到公钥 (测试用)
	signingKey := []byte(sourceEnt.PrivateKey)
	if len(signingKey) == 0 {
		signingKey = []byte(sourceEnt.PublicKey)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}
