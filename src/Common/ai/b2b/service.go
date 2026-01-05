package b2b

import (
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"bytes"
	"context"
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
	db       *gorm.DB
	kb       types.KnowledgeBase // 通过依赖注入获取知识库
	localEnt *models.EnterpriseGORM
	client   *http.Client
	circuits map[uint]*circuitState
	mu       sync.RWMutex
}

type circuitState struct {
	failureCount int
	lastFailure  time.Time
	status       string // "closed", "open", "half-open"
}

const (
	MaxFailures     = 5
	CircuitOpenTime = 30 * time.Second
	MaxRetries      = 3
	RetryWaitTime   = 1 * time.Second
)

var _ B2BService = (*B2BServiceImpl)(nil)

func NewB2BService(db *gorm.DB, kb types.KnowledgeBase) *B2BServiceImpl {
	return &B2BServiceImpl{
		db: db,
		kb: kb,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		circuits: make(map[uint]*circuitState),
	}
}

func (s *B2BServiceImpl) checkCircuit(targetEntID uint) error {
	s.mu.RLock()
	state, ok := s.circuits[targetEntID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	if state.status == "open" {
		if time.Since(state.lastFailure) > CircuitOpenTime {
			// 尝试进入半开状态
			s.mu.Lock()
			state.status = "half-open"
			s.mu.Unlock()
			return nil
		}
		return fmt.Errorf("circuit breaker is open for target enterprise %d", targetEntID)
	}

	return nil
}

func (s *B2BServiceImpl) recordFailure(targetEntID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.circuits[targetEntID]
	if !ok {
		state = &circuitState{status: "closed"}
		s.circuits[targetEntID] = state
	}

	state.failureCount++
	state.lastFailure = time.Now()

	if state.failureCount >= MaxFailures {
		state.status = "open"
		clog.Warn("[B2B] Circuit breaker opened", zap.Uint("target_ent_id", targetEntID))
	}
}

func (s *B2BServiceImpl) recordSuccess(targetEntID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.circuits[targetEntID]
	if !ok {
		return
	}

	state.failureCount = 0
	state.status = "closed"
}

// Connect 建立企业间 B2B 连接
func (s *B2BServiceImpl) Connect(sourceEntCode, targetEntCode string) error {
	var sourceEnt models.EnterpriseGORM
	if err := s.db.Where("code = ?", sourceEntCode).First(&sourceEnt).Error; err != nil {
		return fmt.Errorf("source enterprise not found: %w", err)
	}

	// 1. 获取目标企业的公共 MCP 端点作为握手入口
	var targetEnt models.EnterpriseGORM
	if err := s.db.Where("code = ?", targetEntCode).First(&targetEnt).Error; err != nil {
		return fmt.Errorf("target enterprise not found: %w", err)
	}

	var apiServer models.MCPServerGORM
	if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", targetEnt.ID, "global", "active").First(&apiServer).Error; err != nil {
		return fmt.Errorf("target enterprise has no public MCP endpoint for handshake: %w", err)
	}

	// 2. 构造握手请求
	handshakeURL := strings.TrimSuffix(apiServer.Endpoint, "/") + "/api/b2b/handshake"
	challenge := fmt.Sprintf("handshake_%d", time.Now().UnixNano())

	// 使用私钥签名 challenge
	signature, err := s.signData(sourceEnt.PrivateKey, challenge)
	if err != nil {
		return fmt.Errorf("failed to sign handshake challenge: %w", err)
	}

	reqBody := HandshakeRequest{
		SourceEntCode: sourceEntCode,
		Challenge:     challenge,
		Signature:     signature,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 3. 发起握手 HTTP 请求
	req, _ := http.NewRequest("POST", handshakeURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send handshake request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("handshake failed with status: %d", resp.StatusCode)
	}

	var res HandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode handshake response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("handshake rejected by target")
	}

	// 4. 验证目标企业的响应签名
	if err := s.verifyData(targetEnt.PublicKey, res.Acceptance, res.Signature); err != nil {
		return fmt.Errorf("failed to verify target handshake signature: %w", err)
	}

	// 5. 创建或更新连接记录
	var conn models.B2BConnectionGORM
	err = s.db.Where("source_ent_id = ? AND target_ent_id = ?", sourceEnt.ID, targetEnt.ID).First(&conn).Error
	if err == gorm.ErrRecordNotFound {
		conn = models.B2BConnectionGORM{
			SourceEntID:  sourceEnt.ID,
			TargetEntID:  targetEnt.ID,
			Status:       "active",
			AuthProtocol: "jwt",
		}
		return s.db.Create(&conn).Error
	} else if err == nil {
		conn.Status = "active"
		return s.db.Save(&conn).Error
	}

	return err
}

// HandleHandshake 处理来自外部企业的握手请求
func (s *B2BServiceImpl) HandleHandshake(req HandshakeRequest) (*HandshakeResponse, error) {
	// 1. 获取来源企业信息
	var sourceEnt models.EnterpriseGORM
	if err := s.db.Where("code = ?", req.SourceEntCode).First(&sourceEnt).Error; err != nil {
		return nil, fmt.Errorf("source enterprise not found: %w", err)
	}

	// 2. 验证签名
	if err := s.verifyData(sourceEnt.PublicKey, req.Challenge, req.Signature); err != nil {
		return nil, fmt.Errorf("invalid handshake signature: %w", err)
	}

	// 3. 获取本地企业信息 (假设当前服务属于某个企业，这里需要动态获取或配置)
	var localEnt models.EnterpriseGORM
	// 临时方案：获取 ID 为 1 的企业作为本地企业 (通常是系统默认企业)
	if err := s.db.First(&localEnt, 1).Error; err != nil {
		return nil, fmt.Errorf("local enterprise not found: %w", err)
	}

	// 4. 创建反向连接记录 (建立双向信任)
	var conn models.B2BConnectionGORM
	err := s.db.Where("source_ent_id = ? AND target_ent_id = ?", localEnt.ID, sourceEnt.ID).First(&conn).Error
	if err == gorm.ErrRecordNotFound {
		conn = models.B2BConnectionGORM{
			SourceEntID:  localEnt.ID,
			TargetEntID:  sourceEnt.ID,
			Status:       "active",
			AuthProtocol: "jwt",
		}
		s.db.Create(&conn)
	} else if err == nil {
		conn.Status = "active"
		s.db.Save(&conn)
	}

	// 5. 构造响应
	acceptance := "accepted_" + req.Challenge
	signature, err := s.signData(localEnt.PrivateKey, acceptance)
	if err != nil {
		return nil, fmt.Errorf("failed to sign handshake response: %w", err)
	}

	return &HandshakeResponse{
		Success:    true,
		TargetCode: localEnt.Code,
		Acceptance: acceptance,
		Signature:  signature,
	}, nil
}

// SearchLocalKnowledge 执行本地知识库搜索
func (s *B2BServiceImpl) SearchLocalKnowledge(query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error) {
	if s.kb == nil {
		return nil, fmt.Errorf("knowledge base not initialized")
	}

	return s.kb.Search(context.Background(), query, limit, filter)
}

// SearchMeshKnowledge 在全网（Mesh）范围内搜索知识
func (s *B2BServiceImpl) SearchMeshKnowledge(query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error) {
	// 1. 获取本地结果
	localResults, err := s.SearchLocalKnowledge(query, limit, filter)
	if err != nil {
		clog.Warn("[Mesh] Local knowledge search failed", zap.Error(err))
	}

	allResults := localResults

	// 2. 获取所有活跃的 B2B 连接
	var connections []models.B2BConnectionGORM
	if err := s.db.Where("status = ?", "active").Find(&connections).Error; err != nil {
		return allResults, nil
	}

	// 3. 并发向已连接的企业发起搜索请求
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.B2BConnectionGORM) {
			defer wg.Done()

			// 调用远程企业的 search_knowledge 工具
			args := map[string]any{
				"query": query,
				"limit": limit,
			}

			// 注意：远程调用通过 CallRemoteTool 路由
			resp, err := s.CallRemoteTool(c.SourceEntID, c.TargetEntID, "search_knowledge", args)
			if err != nil {
				clog.Warn(fmt.Sprintf("[Mesh] Failed to query remote knowledge from ent %d: %v", c.TargetEntID, err))
				return
			}

			// 解析响应
			if mcpResp, ok := resp.(map[string]any); ok {
				if content, ok := mcpResp["content"].([]any); ok && len(content) > 0 {
					if first, ok := content[0].(map[string]any); ok {
						if text, ok := first["text"].(string); ok {
							// 获取目标企业信息以标记来源
							var targetEnt models.EnterpriseGORM
							s.db.First(&targetEnt, c.TargetEntID)

							mu.Lock()
							allResults = append(allResults, types.DocChunk{
								ID:      fmt.Sprintf("mesh_%d_%s", c.TargetEntID, utils.GenerateRandomToken(8)),
								Content: text,
								Source:  fmt.Sprintf("Mesh:%s", targetEnt.Name),
								Score:   0.85,
							})
							mu.Unlock()
						}
					}
				}
			}
		}(conn)
	}

	wg.Wait()
	return allResults, nil
}

func (s *B2BServiceImpl) CheckDispatchPermission(employeeID, targetOrgID uint, action string) (bool, error) {
	// 简单实现，后期可扩展为复杂的 RBAC/Mesh 策略
	return true, nil
}

func (s *B2BServiceImpl) signData(privateKeyStr, data string) (string, error) {
	claims := jwt.MapClaims{
		"data": data,
		"exp":  time.Now().Add(time.Minute * 5).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(privateKeyStr))
}

func (s *B2BServiceImpl) verifyData(publicKeyStr, data, signature string) error {
	token, err := jwt.Parse(signature, func(token *jwt.Token) (interface{}, error) {
		return []byte(publicKeyStr), nil
	})

	if err != nil || !token.Valid {
		return fmt.Errorf("invalid signature")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["data"] != data {
		return fmt.Errorf("data mismatch in signature")
	}

	return nil
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

	// 4. 调用远程企业的 im_send_message 工具
	args := map[string]any{
		"platform":  "mesh",
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

func (s *B2BServiceImpl) checkSkillSharing(sourceEntID, targetEntID uint, skillName string) error {
	if skillName == "im_send_message" {
		return nil
	}

	var sharing models.B2BSkillSharingGORM
	err := s.db.Where("source_ent_id = ? AND target_ent_id = ? AND skill_name = ?",
		targetEntID, sourceEntID, skillName).First(&sharing).Error

	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("skill '%s' is not shared between these enterprises", skillName)
	}
	if err != nil {
		return fmt.Errorf("failed to check skill sharing: %w", err)
	}

	if !sharing.IsActive {
		return fmt.Errorf("skill '%s' sharing is currently inactive", skillName)
	}

	if sharing.Status != "approved" {
		return fmt.Errorf("skill '%s' sharing is in status '%s' (not approved)", skillName, sharing.Status)
	}

	return nil
}

func (s *B2BServiceImpl) CallRemoteTool(sourceEntID, targetEntID uint, toolName string, arguments map[string]any) (any, error) {
	if err := s.checkCircuit(targetEntID); err != nil {
		return nil, err
	}

	if err := s.checkSkillSharing(sourceEntID, targetEntID, toolName); err != nil {
		return nil, fmt.Errorf("b2b skill authorization failed: %w", err)
	}

	var apiServer models.MCPServerGORM
	if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", targetEntID, "global", "active").First(&apiServer).Error; err != nil {
		return nil, fmt.Errorf("target enterprise has no public MCP endpoint: %w", err)
	}

	callURL := apiServer.Endpoint
	if !strings.Contains(callURL, "/api/mcp/v1/tools/call") {
		callURL = strings.TrimSuffix(callURL, "/") + "/api/mcp/v1/tools/call"
	}

	reqBody := map[string]any{
		"server_id": "mesh_bridge",
		"tool_name": toolName,
		"arguments": arguments,
	}
	jsonBody, _ := json.Marshal(reqBody)

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		token, err := s.generateB2BToken(sourceEntID, targetEntID)
		if err != nil {
			return nil, err
		}

		req, _ := http.NewRequest("POST", callURL, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to remote peer: %w", err)
			clog.Warn("[B2B] Request failed, retrying...", zap.Int("attempt", i+1), zap.Error(err))
			time.Sleep(RetryWaitTime * time.Duration(i+1))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errRes struct {
				Message string `json:"message"`
			}
			json.NewDecoder(resp.Body).Decode(&errRes)
			resp.Body.Close()
			lastErr = fmt.Errorf("remote error (status %d): %s", resp.StatusCode, errRes.Message)

			if resp.StatusCode >= 500 {
				clog.Warn("[B2B] Remote server error, retrying...", zap.Int("attempt", i+1), zap.Int("status", resp.StatusCode))
				time.Sleep(RetryWaitTime * time.Duration(i+1))
				continue
			}

			s.recordFailure(targetEntID)
			return nil, lastErr
		}

		var result struct {
			Success bool `json:"success"`
			Data    any  `json:"data"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if decodeErr != nil {
			lastErr = fmt.Errorf("failed to decode remote response: %w", decodeErr)
			s.recordFailure(targetEntID)
			return nil, lastErr
		}

		s.recordSuccess(targetEntID)
		return result.Data, nil
	}

	s.recordFailure(targetEntID)
	return nil, fmt.Errorf("failed after %d retries: %v", MaxRetries, lastErr)
}

func (s *B2BServiceImpl) generateB2BToken(sourceEntID, targetEntID uint) (string, error) {
	var sourceEnt models.EnterpriseGORM
	if err := s.db.First(&sourceEnt, sourceEntID).Error; err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"iss": sourceEnt.Code,
		"sub": fmt.Sprintf("ent_%d", targetEntID),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sourceEnt.PrivateKey))
}

// VerifyIdentity 验证企业身份
func (s *B2BServiceImpl) VerifyIdentity(entCode string, signature string) bool {
	var ent models.EnterpriseGORM
	if err := s.db.Where("code = ?", entCode).First(&ent).Error; err != nil {
		return false
	}

	token, err := jwt.Parse(signature, func(token *jwt.Token) (interface{}, error) {
		return []byte(ent.PublicKey), nil
	})

	return err == nil && token.Valid
}

// VerifyB2BToken 验证 B2B JWT 令牌并返回企业信息
func (s *B2BServiceImpl) VerifyB2BToken(tokenString string) (*models.EnterpriseGORM, error) {
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

	var ent models.EnterpriseGORM
	if err := s.db.Where("code = ?", entCode).First(&ent).Error; err != nil {
		return nil, fmt.Errorf("enterprise %s not found: %w", entCode, err)
	}

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
		Scope:    "global",
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
	localServers, err := s.DiscoverEndpoints(query)
	if err != nil {
		clog.Error("[Mesh] Local discovery failed", zap.Error(err))
	}

	allServers := localServers

	var connections []models.B2BConnectionGORM
	if err := s.db.Where("status = ?", "active").Find(&connections).Error; err != nil {
		return allServers, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.B2BConnectionGORM) {
			defer wg.Done()

			var targetEnt models.EnterpriseGORM
			if err := s.db.First(&targetEnt, c.TargetEntID).Error; err != nil {
				return
			}

			var apiServer models.MCPServerGORM
			if err := s.db.Where("owner_id = ? AND scope = ? AND status = ?", c.TargetEntID, "global", "active").First(&apiServer).Error; err != nil {
				return
			}

			discoverURL := apiServer.Endpoint
			if !strings.Contains(discoverURL, "/api/mesh/discover") {
				discoverURL = strings.TrimSuffix(discoverURL, "/") + "/api/mesh/discover"
			}
			discoverURL += "?q=" + query

			token, _ := s.generateB2BToken(c.SourceEntID, c.TargetEntID)

			req, _ := http.NewRequest("GET", discoverURL, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := s.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var res struct {
					Success bool                   `json:"success"`
					Data    []models.MCPServerGORM `json:"data"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.Success {
					mu.Lock()
					allServers = append(allServers, res.Data...)
					mu.Unlock()
				}
			}
		}(conn)
	}

	wg.Wait()
	return allServers, nil
}

// RequestSkillSharing 请求技能共享
func (s *B2BServiceImpl) RequestSkillSharing(fromEntID, targetEntID uint, skillName string) error {
	sharing := models.B2BSkillSharingGORM{
		SourceEntID: targetEntID, // 提供方是目标企业
		TargetEntID: fromEntID,   // 使用方是发起请求的企业
		SkillName:   skillName,
		Status:      "pending",
		IsActive:    true,
	}
	return s.db.Create(&sharing).Error
}

// ApproveSkillSharing 审批技能共享
func (s *B2BServiceImpl) ApproveSkillSharing(sharingID uint, status string) error {
	return s.db.Model(&models.B2BSkillSharingGORM{}).Where("id = ?", sharingID).Update("status", status).Error
}

// ListSkillSharings 列出技能共享
func (s *B2BServiceImpl) ListSkillSharings(entID uint, role string) ([]models.B2BSkillSharingGORM, error) {
	var sharings []models.B2BSkillSharingGORM
	db := s.db
	if role == "source" {
		db = db.Where("source_ent_id = ?", entID)
	} else if role == "target" {
		db = db.Where("target_ent_id = ?", entID)
	} else {
		db = db.Where("source_ent_id = ? OR target_ent_id = ?", entID, entID)
	}
	err := db.Find(&sharings).Error
	return sharings, err
}

// DispatchEmployee 外派员工
func (s *B2BServiceImpl) DispatchEmployee(employeeID, sourceEntID, targetEntID uint, permissions []string) error {
	permJSON, _ := json.Marshal(permissions)
	dispatch := models.DigitalEmployeeDispatchGORM{
		EmployeeID:  employeeID,
		SourceEntID: sourceEntID,
		TargetEntID: targetEntID,
		Permissions: string(permJSON),
		Status:      "pending",
		DispatchAt:  time.Now(),
	}
	return s.db.Create(&dispatch).Error
}

// ApproveDispatch 审批外派
func (s *B2BServiceImpl) ApproveDispatch(dispatchID uint, status string) error {
	return s.db.Model(&models.DigitalEmployeeDispatchGORM{}).Where("id = ?", dispatchID).Update("status", status).Error
}

// ListDispatchedEmployees 列出外派记录
func (s *B2BServiceImpl) ListDispatchedEmployees(entID uint, role string) ([]models.DigitalEmployeeDispatchGORM, error) {
	var dispatches []models.DigitalEmployeeDispatchGORM
	db := s.db
	if role == "source" {
		db = db.Where("source_ent_id = ?", entID)
	} else if role == "target" {
		db = db.Where("target_ent_id = ?", entID)
	} else {
		db = db.Where("source_ent_id = ? OR target_ent_id = ?", entID, entID)
	}
	err := db.Find(&dispatches).Error
	return dispatches, err
}
