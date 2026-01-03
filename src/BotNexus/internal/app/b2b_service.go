package app

import (
	"BotMatrix/common/ai"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	// 4. 生成跨域认证 Token (JWT)
	token, err := s.generateB2BToken(fromEmp.EnterpriseID, toEmp.EnterpriseID)
	if err != nil {
		return fmt.Errorf("failed to generate B2B token: %w", err)
	}

	clog.Info(fmt.Sprintf("[B2B] Sending message from %s to %s | Token: %s...", fromEmployeeID, toEmployeeID, token[:10]))

	// 5. 模拟发送 (在实际分布式架构中，这里会调用远程 BotMatrix 的 API)
	// TODO: 调用远程 MCP Endpoint

	return nil
}

// CallRemoteTool 调用远程企业的 MCP 工具
func (s *B2BServiceImpl) CallRemoteTool(fromEntID uint, targetEntID uint, toolName string, arguments map[string]any) (any, error) {
	// 1. 获取目标企业信息
	var targetEnt models.EnterpriseGORM
	if err := s.db.First(&targetEnt, targetEntID).Error; err != nil {
		return nil, fmt.Errorf("target enterprise not found: %w", err)
	}

	// 2. 检查 B2B 连接
	var conn models.B2BConnectionGORM
	if err := s.db.Where("source_ent_id = ? AND target_ent_id = ? AND status = ?", fromEntID, targetEntID, "active").First(&conn).Error; err != nil {
		return nil, fmt.Errorf("no active B2B connection: %w", err)
	}

	// 3. 寻找目标端点 (假设目标企业有一个名为 'main' 的公开端点)
	var server models.MCPServerGORM
	if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", targetEntID, "global", "active").First(&server).Error; err != nil {
		return nil, fmt.Errorf("no public MCP endpoint found for enterprise %d", targetEntID)
	}

	// 4. 生成 JWT
	token, err := s.generateB2BToken(fromEntID, targetEntID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate B2B token: %w", err)
	}

	// 5. 构造请求 (假设端点是 SSE, 我们直接调用其 /tools/call)
	// 在实际 mesh 中，端点 URL 可能需要拼接
	callURL := server.Endpoint
	if !bytes.Contains([]byte(callURL), []byte("/tools/call")) {
		// 简单的 URL 拼接逻辑，实际应更健壮
		if bytes.HasSuffix([]byte(callURL), []byte("/")) {
			callURL += "tools/call"
		} else {
			callURL += "/tools/call"
		}
	}

	reqBody := ai.MCPCallToolRequest{
		Name:      toolName,
		Arguments: arguments,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", callURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	clog.Info(fmt.Sprintf("[Mesh] Calling remote tool %s at %s", toolName, callURL))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("remote call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote call returned status %d", resp.StatusCode)
	}

	var mcpResp ai.MCPCallToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if mcpResp.IsError {
		errorMsg := "Unknown error"
		if len(mcpResp.Content) > 0 {
			errorMsg = mcpResp.Content[0].Text
		}
		return nil, fmt.Errorf("remote tool error: %s", errorMsg)
	}

	return mcpResp, nil
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
