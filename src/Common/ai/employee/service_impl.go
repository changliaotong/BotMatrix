package employee

import (
	"BotMatrix/common/ai/b2b"
	"BotMatrix/common/ai/rag"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// --- DigitalEmployeeService Implementation ---

type EmployeeServiceImpl struct {
	db      *gorm.DB
	aiSvc   types.AIService
	factory *DigitalEmployeeFactory
}

var _ DigitalEmployeeService = (*EmployeeServiceImpl)(nil)

func NewEmployeeService(db *gorm.DB) *EmployeeServiceImpl {
	return &EmployeeServiceImpl{
		db:      db,
		factory: NewDigitalEmployeeFactory(db),
	}
}

func (s *EmployeeServiceImpl) Recruit(ctx context.Context, jobID uint, enterpriseID uint, name string, botID string) (*models.DigitalEmployee, error) {
	return s.factory.Recruit(ctx, RecruitParams{
		JobID:        jobID,
		EnterpriseID: enterpriseID,
		Name:         name,
		BotID:        botID,
	})
}

func (s *EmployeeServiceImpl) Fire(ctx context.Context, employeeID uint, reason string) error {
	return s.db.WithContext(ctx).Model(&models.DigitalEmployee{}).
		Where("id = ?", employeeID).
		Updates(map[string]interface{}{
			"status":     "terminated",
			"updated_at": time.Now(),
		}).Error
}

func (s *EmployeeServiceImpl) Transfer(ctx context.Context, employeeID uint, newJobID uint) error {
	// 1. Get current job relation and terminate it
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.EmployeeJobRelation{}).
		Where("employee_id = ? AND status = 'active'", employeeID).
		Updates(map[string]interface{}{
			"status": "transferred",
			"end_at": &now,
		}).Error; err != nil {
		return err
	}

	// 2. Create new job relation
	newRelation := models.EmployeeJobRelation{
		EmployeeID: employeeID,
		JobID:      newJobID,
		IsPrimary:  true,
		AssignedAt: now,
		Status:     "active",
	}
	if err := s.db.WithContext(ctx).Create(&newRelation).Error; err != nil {
		return err
	}
	
	// 3. Update employee title/dept cache
	var job models.DigitalJob
	if err := s.db.First(&job, newJobID).Error; err == nil {
		s.db.Model(&models.DigitalEmployee{}).Where("id = ?", employeeID).Updates(map[string]interface{}{
			"title":      job.Name,
			"department": job.Department,
		})
	}

	return nil
}

func (s *EmployeeServiceImpl) SetAIService(aiSvc types.AIService) {
	s.aiSvc = aiSvc
}

func (s *EmployeeServiceImpl) GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error) {
	var employee models.DigitalEmployee
	if err := s.db.Where("bot_id = ?", botID).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (s *EmployeeServiceImpl) RecordKpi(employeeID uint, metric string, score float64) error {
	log := models.DigitalEmployeeKpi{
		EmployeeID: employeeID,
		MetricName: metric,
		Score:      score,
	}
	if err := s.db.Create(&log).Error; err != nil {
		return err
	}

	var avgScore float64
	s.db.Model(&models.DigitalEmployeeKpi{}).
		Where("employee_id = ?", employeeID).
		Select("AVG(score)").
		Scan(&avgScore)

	err := s.db.Model(&models.DigitalEmployee{}).
		Where("id = ?", employeeID).
		Update("kpi_score", avgScore).Error

	if err != nil {
		return err
	}

	if avgScore < 85 {
		var recentEvolution int64
		s.db.Model(&models.DigitalEmployeeKpi{}).
			Where("employee_id = ? AND metric_name = ? AND created_at > ?", employeeID, "auto_evolution", time.Now().Add(-24*time.Hour)).
			Count(&recentEvolution)

		if recentEvolution == 0 {
			go s.AutoEvolve(employeeID)
		}
	}

	return nil
}

func (s *EmployeeServiceImpl) UpdateOnlineStatus(botID string, status string) error {
	return s.db.Model(&models.DigitalEmployee{}).
		Where("bot_id = ?", botID).
		Update("online_status", status).Error
}

func (s *EmployeeServiceImpl) ConsumeSalary(botID string, tokens int64) error {
	return s.db.Model(&models.DigitalEmployee{}).
		Where("bot_id = ?", botID).
		UpdateColumn("salary_token", gorm.Expr("salary_token + ?", tokens)).Error
}

func (s *EmployeeServiceImpl) CheckSalaryLimit(botID string) (bool, error) {
	var employee models.DigitalEmployee
	if err := s.db.Where("bot_id = ?", botID).First(&employee).Error; err != nil {
		return false, err
	}

	if employee.SalaryLimit > 0 && employee.SalaryToken > employee.SalaryLimit {
		return false, nil
	}

	return true, nil
}

func (s *EmployeeServiceImpl) UpdateSalary(botID string, salaryToken *int64, salaryLimit *int64) error {
	updates := make(map[string]interface{})
	if salaryToken != nil {
		updates["salary_token"] = *salaryToken
	}
	if salaryLimit != nil {
		updates["salary_limit"] = *salaryLimit
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.Model(&models.DigitalEmployee{}).
		Where("bot_id = ?", botID).
		Updates(updates).Error
}

func (s *EmployeeServiceImpl) AutoEvolve(employeeID uint) error {
	if s.aiSvc == nil {
		return fmt.Errorf("AI service not initialized")
	}

	var employee models.DigitalEmployee
	if err := s.db.Preload("Agent").First(&employee, employeeID).Error; err != nil {
		return err
	}

	if employee.AgentID == 0 {
		return fmt.Errorf("employee %d has no associated agent", employeeID)
	}

	var kpis []models.DigitalEmployeeKpi
	s.db.Where("employee_id = ?", employeeID).Order("created_at desc").Limit(10).Find(&kpis)

	if len(kpis) == 0 {
		return nil
	}

	var feedback string
	var totalScore float64
	for _, k := range kpis {
		totalScore += k.Score
		if k.Detail != "" {
			feedback += fmt.Sprintf("- [%s] %s: %s\n", k.CreatedAt.Format("2006-01-02"), k.MetricName, k.Detail)
		}
	}
	avgScore := totalScore / float64(len(kpis))

	if avgScore >= 95 && feedback == "" {
		return nil
	}

	clog.Info("开始数字员工自动进化", zap.Uint("id", employeeID), zap.String("name", employee.Name), zap.Float64("avg_score", avgScore))

	systemPrompt := `你是一个资深的 AI 提示词架构师。你的任务是根据数字员工的当前系统提示词和最近的 KPI 绩效反馈，优化其提示词。
数字员工信息：
- 姓名：{{.Name}}
- 职位：{{.Title}}
- 部门：{{.Department}}
- 简介：{{.Bio}}

当前系统提示词：
"""
{{.CurrentPrompt}}
"""

最近的绩效反馈与评分（平均分：{{.AvgScore}}）：
"""
{{.Feedback}}
"""

请分析反馈中的不足（如：专业度不够、回复太慢、语气生硬、未遵循规范等），并输出一个优化后的、更强大的系统提示词。
要求：
1. 保持原有的人设特征。
2. 针对性地解决反馈中提到的问题。
3. 增强对复杂场景的处理能力。
4. 只输出优化后的系统提示词内容，不要包含其他解释。`

	instruction := systemPrompt
	instruction = strings.ReplaceAll(instruction, "{{.Name}}", employee.Name)
	instruction = strings.ReplaceAll(instruction, "{{.Title}}", employee.Title)
	instruction = strings.ReplaceAll(instruction, "{{.Department}}", employee.Department)
	instruction = strings.ReplaceAll(instruction, "{{.Bio}}", employee.Bio)
	instruction = strings.ReplaceAll(instruction, "{{.CurrentPrompt}}", employee.Agent.Prompt)
	instruction = strings.ReplaceAll(instruction, "{{.AvgScore}}", fmt.Sprintf("%.2f", avgScore))
	instruction = strings.ReplaceAll(instruction, "{{.Feedback}}", feedback)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := s.aiSvc.Chat(ctx, employee.Agent.ModelID, []types.Message{
		{Role: types.RoleUser, Content: instruction},
	}, nil)

	if err != nil {
		return fmt.Errorf("AI optimization failed: %v", err)
	}

	newPrompt := ""
	if len(resp.Choices) > 0 {
		if content, ok := resp.Choices[0].Message.Content.(string); ok {
			newPrompt = strings.TrimSpace(content)
		}
	}

	if newPrompt == "" || newPrompt == employee.Agent.Prompt {
		return nil
	}

	if err := s.db.Model(&models.AIAgent{}).Where("id = ?", employee.AgentID).Update("prompt", newPrompt).Error; err != nil {
		return err
	}

	evolutionLog := models.DigitalEmployeeKpi{
		EmployeeID: employeeID,
		MetricName: "auto_evolution",
		Score:      avgScore,
		Detail:     fmt.Sprintf("提示词已自动优化。旧评分: %.2f。反馈摘要: %d 条记录已处理。", avgScore, len(kpis)),
	}
	s.db.Create(&evolutionLog)

	clog.Info("数字员工进化成功", zap.Uint("id", employeeID), zap.String("name", employee.Name))

	return nil
}

// --- CognitiveMemoryService Implementation ---

type CognitiveMemoryServiceImpl struct {
	db           *gorm.DB
	embeddingSvc rag.EmbeddingService
}

var _ CognitiveMemoryService = (*CognitiveMemoryServiceImpl)(nil)

func NewCognitiveMemoryService(db *gorm.DB) *CognitiveMemoryServiceImpl {
	return &CognitiveMemoryServiceImpl{
		db: db,
	}
}

func (s *CognitiveMemoryServiceImpl) SetEmbeddingService(svc any) {
	if es, ok := svc.(rag.EmbeddingService); ok {
		s.embeddingSvc = es
	}
}

func (s *CognitiveMemoryServiceImpl) GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemory, error) {
	var userMemories []models.CognitiveMemory
	var roleMemories []models.CognitiveMemory

	hasVector := false
	if query != "" && s.embeddingSvc != nil && s.db.Dialector.Name() == "postgres" {
		s.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVector)
	}

	if hasVector {
		vector, err := s.embeddingSvc.GenerateQueryEmbedding(ctx, query)
		if err == nil {
			vectorStr, _ := json.Marshal(vector)
			err = s.db.WithContext(ctx).
				Where("user_id = ? AND bot_id = ?", userID, botID).
				Order(fmt.Sprintf("embedding <=> '%s'", string(vectorStr))).
				Limit(10).
				Find(&userMemories).Error
			if err != nil {
				clog.Warn("[Memory] Vector search failed, falling back to keyword", zap.Error(err))
			}
		} else {
			clog.Warn("[Memory] Failed to generate embedding for query", zap.Error(err))
		}
	}

	if len(userMemories) == 0 {
		userQuery := s.db.WithContext(ctx).Where("user_id = ? AND bot_id = ?", userID, botID)
		if query != "" {
			userQuery = userQuery.Where("content LIKE ?", "%"+query+"%")
		}
		err := userQuery.Order("importance DESC, last_seen DESC").Limit(10).Find(&userMemories).Error
		if err != nil {
			clog.Error("[Memory] Failed to get user memories", zap.Error(err))
		}
	}

	roleQuery := s.db.WithContext(ctx).Where("user_id = '' AND bot_id = ?", botID)
	err := roleQuery.Order("importance DESC").Limit(5).Find(&roleMemories).Error
	if err != nil {
		clog.Error("[Memory] Failed to get role memories", zap.Error(err))
	}

	allMemories := append(roleMemories, userMemories...)
	return allMemories, nil
}

func (s *CognitiveMemoryServiceImpl) GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemory, error) {
	var memories []models.CognitiveMemory
	err := s.db.WithContext(ctx).
		Where("user_id = '' AND bot_id = ?", botID).
		Order("importance DESC").
		Find(&memories).Error
	return memories, err
}

func (s *CognitiveMemoryServiceImpl) SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemory, error) {
	var memories []models.CognitiveMemory
	db := s.db.WithContext(ctx).Where("bot_id = ?", botID)

	if query != "" {
		db = db.Where("content LIKE ?", "%"+query+"%")
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}

	err := db.Order("last_seen DESC").Limit(20).Find(&memories).Error
	return memories, err
}

func (s *CognitiveMemoryServiceImpl) SaveMemory(ctx context.Context, memory *models.CognitiveMemory) error {
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}
	memory.LastSeen = time.Now()

	if s.embeddingSvc != nil && memory.Content != "" {
		vec, err := s.embeddingSvc.GenerateEmbedding(ctx, memory.Content)
		if err == nil {
			vecJSON, _ := json.Marshal(vec)
			memory.Embedding = string(vecJSON)
		} else {
			clog.Warn("[Memory] Failed to generate embedding for memory", zap.Error(err))
		}
	}

	return s.db.WithContext(ctx).Save(memory).Error
}

func (s *CognitiveMemoryServiceImpl) ForgetMemory(ctx context.Context, memoryID uint) error {
	return s.db.WithContext(ctx).Delete(&models.CognitiveMemory{}, memoryID).Error
}

func (s *CognitiveMemoryServiceImpl) ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc types.AIService) error {
	if aiSvc == nil {
		return fmt.Errorf("AI service is required for consolidation")
	}

	var memories []models.CognitiveMemory
	query := s.db.WithContext(ctx).Where("bot_id = ?", botID)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id = '' OR user_id IS NULL")
	}

	err := query.Order("category, created_at ASC").Find(&memories).Error
	if err != nil {
		return err
	}

	if len(memories) < 10 {
		clog.Info("[Memory] Too few memories to consolidate", zap.String("bot_id", botID), zap.Int("count", len(memories)))
		return nil
	}

	prompt := "你是一个记忆 management 专家。以下是关于某个数字员工或其与用户交互的碎片化记忆片段。请将这些记忆进行逻辑合并、去重并提炼。\n"
	prompt += "规则：\n1. 合并相似或相关的片段（例如：‘喜欢苹果’和‘喜欢红富士’可以合并为‘喜欢各种苹果’）。\n2. 保持分类清晰。\n3. 提炼出更有深度的洞察，而不仅仅是堆砌事实。\n4. 格式：[类别] 提炼后的内容。\n\n记忆片段：\n"

	for _, m := range memories {
		cat := m.Category
		if cat == "" {
			cat = "general"
		}
		prompt += fmt.Sprintf("- [%s] %s\n", cat, m.Content)
	}

	msgs := []types.Message{
		{Role: types.RoleSystem, Content: prompt},
	}

	resp, err := aiSvc.Chat(ctx, 0, msgs, nil)
	if err != nil {
		return fmt.Errorf("AI chat failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("AI returned no choices")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || strings.TrimSpace(content) == "" {
		return fmt.Errorf("AI returned empty content")
	}

	lines := strings.Split(content, "\n")
	var newMemories []models.CognitiveMemory
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		category := "general"
		fact := line
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			idx := strings.Index(line, "]")
			category = line[1:idx]
			fact = strings.TrimSpace(line[idx+1:])
		}

		newMemories = append(newMemories, models.CognitiveMemory{
			UserID:     userID,
			BotID:      botID,
			Category:   category,
			Content:    fact,
			Importance: 3,
			LastSeen:   time.Now(),
		})
	}

	if len(newMemories) > 0 {
		return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			delQuery := tx.Where("bot_id = ?", botID)
			if userID != "" {
				delQuery = delQuery.Where("user_id = ?", userID)
			} else {
				delQuery = delQuery.Where("user_id = '' OR user_id IS NULL")
			}

			if err := delQuery.Delete(&models.CognitiveMemory{}).Error; err != nil {
				return err
			}

			for _, m := range newMemories {
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}

	return nil
}

func (s *CognitiveMemoryServiceImpl) LearnFromURL(ctx context.Context, botID string, url string, category string) error {
	clog.Info("[Memory] Learning from URL", zap.String("bot_id", botID), zap.String("url", url))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("URL returned status: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	filename := filepath.Base(url)
	if !strings.Contains(filename, ".") {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			filename = "index.html"
		} else if strings.Contains(contentType, "application/pdf") {
			filename = "doc.pdf"
		} else {
			filename = "content.txt"
		}
	}

	return s.LearnFromContent(ctx, botID, content, filename, category)
}

func (s *CognitiveMemoryServiceImpl) LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error {
	clog.Info("[Memory] Learning from content", zap.String("bot_id", botID), zap.String("filename", filename))

	var parser rag.ContentParser
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		parser = &rag.PDFParser{}
	case ".xlsx", ".xls":
		parser = &rag.ExcelParser{}
	case ".docx":
		parser = &rag.DocxParser{}
	case ".md", ".markdown":
		parser = &rag.MarkdownParser{MinSize: 50}
	case ".go", ".py", ".js", ".ts", ".java", ".c", ".cpp":
		parser = &rag.CodeParser{}
	case ".html", ".htm":
		re := regexp.MustCompile("<[^>]*>")
		stripped := re.ReplaceAllString(string(content), "")
		content = []byte(stripped)
		parser = &rag.TxtParser{MinSize: 50}
	default:
		parser = &rag.TxtParser{MinSize: 50}
	}

	chunks := parser.Parse(ctx, content)
	if len(chunks) == 0 {
		return fmt.Errorf("no content extracted from %s", filename)
	}

	for i, chunk := range chunks {
		mem := &models.CognitiveMemory{
			BotID:      botID,
			UserID:     "",
			Content:    chunk.Content,
			Category:   category,
			Importance: 3,
			Metadata:   fmt.Sprintf("Source: %s, Part: %d", filename, i+1),
		}
		if chunk.Title != "" {
			mem.Metadata += ", Title: " + chunk.Title
		}

		if err := s.SaveMemory(ctx, mem); err != nil {
			clog.Error("[Memory] Failed to save learned chunk", zap.Error(err), zap.Int("index", i))
		}
	}

	clog.Info("[Memory] Learning complete", zap.String("bot_id", botID), zap.Int("chunks", len(chunks)))
	return nil
}

// --- DigitalEmployeeTaskService Implementation ---

type DigitalEmployeeTaskServiceImpl struct {
	db         *gorm.DB
	aiSvc      types.AIService
	mcpManager types.MCPManagerInterface
}

var _ DigitalEmployeeTaskService = (*DigitalEmployeeTaskServiceImpl)(nil)

func NewDigitalEmployeeTaskService(db *gorm.DB, mcp types.MCPManagerInterface) *DigitalEmployeeTaskServiceImpl {
	return &DigitalEmployeeTaskServiceImpl{
		db:         db,
		mcpManager: mcp,
	}
}

func (s *DigitalEmployeeTaskServiceImpl) SetAIService(aiSvc types.AIService) {
	s.aiSvc = aiSvc
}

type TaskPlan struct {
	Steps []TaskStep `json:"steps"`
}

type TaskStep struct {
	Index            int    `json:"index"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Tool             string `json:"tool,omitempty"`
	RequiresApproval bool   `json:"requires_approval,omitempty"`
}

func (s *DigitalEmployeeTaskServiceImpl) CreateTask(ctx context.Context, task *models.DigitalEmployeeTask) error {
	return s.db.Create(task).Error
}

func (s *DigitalEmployeeTaskServiceImpl) UpdateTaskStatus(ctx context.Context, executionID string, status string, progress int) error {
	updates := map[string]any{
		"status":   status,
		"progress": progress,
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["end_time"] = &now
	}
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("execution_id = ?", executionID).
		Updates(updates).Error
}

func (s *DigitalEmployeeTaskServiceImpl) GetPendingTasks(ctx context.Context, employeeID uint) ([]*models.DigitalEmployeeTask, error) {
	var tasks []*models.DigitalEmployeeTask
	err := s.db.Where("assignee_id = ? AND status = ?", employeeID, "pending").Find(&tasks).Error
	return tasks, err
}

func (s *DigitalEmployeeTaskServiceImpl) GetTaskByExecutionID(ctx context.Context, executionID string) (*models.DigitalEmployeeTask, error) {
	var task models.DigitalEmployeeTask
	if err := s.db.WithContext(ctx).Where("execution_id = ?", executionID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *DigitalEmployeeTaskServiceImpl) AssignTask(ctx context.Context, executionID string, assigneeID uint) error {
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("execution_id = ?", executionID).
		Update("assignee_id", assigneeID).Error
}

func (s *DigitalEmployeeTaskServiceImpl) PlanTask(ctx context.Context, executionID string) error {
	task, err := s.GetTaskByExecutionID(ctx, executionID)
	if err != nil {
		return err
	}

	if s.aiSvc == nil {
		return fmt.Errorf("AI service not initialized")
	}

	s.UpdateTaskStatus(ctx, executionID, "planning", 10)

	prompt := fmt.Sprintf(`你是一个专业的数字员工任务规划器。
请为以下任务制定详细的执行计划：
任务标题: %s
任务描述: %s

要求：
1. 将任务分解为 3-5 个具体的步骤。
2. 每个步骤包含标题、详细描述。
3. 如果步骤需要使用工具（如搜索知识库、发送消息等），请注明。
4. 如果步骤涉及高风险操作（如资金划转、发送外部消息、修改核心配置），请将 "requires_approval" 设为 true。
5. 请以 JSON 格式返回，格式如下：
{"steps": [{"index": 1, "title": "步骤标题", "description": "步骤描述", "tool": "可选工具名称", "requires_approval": false}]}`, task.Title, task.Description)

	resp, err := s.aiSvc.Chat(ctx, 0, []types.Message{
		{Role: types.RoleSystem, Content: "你是一个专业的任务规划助手。"},
		{Role: types.RoleUser, Content: prompt},
	}, nil)

	if err != nil {
		s.UpdateTaskStatus(ctx, executionID, "failed", 0)
		return fmt.Errorf("AI planning failed: %v", err)
	}

	var plan TaskPlan
	var content string
	if len(resp.Choices) > 0 {
		if c, ok := resp.Choices[0].Message.Content.(string); ok {
			content = c
		}
	}

	if content == "" {
		s.UpdateTaskStatus(ctx, executionID, "failed", 0)
		return fmt.Errorf("AI planning returned empty content")
	}

	jsonStr := content
	if start := strings.Index(content, "{"); start != -1 {
		if end := strings.LastIndex(content, "}"); end != -1 && end > start {
			jsonStr = content[start : end+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return fmt.Errorf("failed to parse task plan: %v", err)
	}

	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("execution_id = ?", executionID).
		Updates(map[string]any{
			"plan_raw": content,
			"status":   "executing",
			"progress": 30,
		}).Error
}

func (s *DigitalEmployeeTaskServiceImpl) ExecuteTask(ctx context.Context, executionID string) error {
	task, err := s.GetTaskByExecutionID(ctx, executionID)
	if err != nil {
		return err
	}

	if task.PlanRaw == "" {
		return fmt.Errorf("task plan is empty, call PlanTask first")
	}

	var plan TaskPlan
	if err := json.Unmarshal([]byte(task.PlanRaw), &plan); err != nil {
		if strings.Contains(task.PlanRaw, "```json") {
			parts := strings.Split(task.PlanRaw, "```json")
			if len(parts) > 1 {
				jsonPart := strings.Split(parts[1], "```")[0]
				json.Unmarshal([]byte(strings.TrimSpace(jsonPart)), &plan)
			}
		}
	}

	if len(plan.Steps) == 0 {
		return fmt.Errorf("task plan has no steps")
	}

	if task.Status != "executing" {
		clog.Info("[TaskEngine] Starting execution", zap.String("execution_id", executionID), zap.String("title", task.Title))
		s.UpdateTaskStatus(ctx, executionID, "executing", 30)
	}

	var results []string
	if task.ResultRaw != "" {
		results = append(results, task.ResultRaw)
	}

	totalSteps := len(plan.Steps)
	startIndex := task.CurrentStepIndex

	for i := startIndex; i < totalSteps; i++ {
		step := plan.Steps[i]

		if step.RequiresApproval && task.Status != "approved" {
			s.db.Model(&models.DigitalEmployeeTask{}).
				Where("execution_id = ?", executionID).
				Updates(map[string]any{
					"status":             "pending_approval",
					"current_step_index": i,
					"result_raw":         strings.Join(results, "\n\n"),
				})
			fmt.Printf("[Task] Step %d requires approval, pausing execution\n", i+1)
			return nil
		}

		progress := 30 + int(float64(i+1)/float64(totalSteps)*60)
		clog.Info("[TaskEngine] Executing step",
			zap.String("execution_id", executionID),
			zap.Int("step", i+1),
			zap.Int("total", totalSteps),
			zap.String("step_title", step.Title))

		var stepResult string
		if step.Tool != "" && s.mcpManager != nil {
			toolRes, err := s.mcpManager.CallTool(ctx, step.Tool, map[string]any{
				"context":   task.Description,
				"objective": step.Description,
			})
			if err != nil {
				stepResult = fmt.Sprintf("Step %d Tool Error: %v", step.Index, err)
			} else {
				stepResult = fmt.Sprintf("Step %d Result: %v", step.Index, toolRes)
			}
		} else {
			resp, err := s.aiSvc.Chat(ctx, 0, []types.Message{
				{Role: types.RoleSystem, Content: "你是一个正在执行任务步骤的数字员工。"},
				{Role: types.RoleUser, Content: fmt.Sprintf("任务上下文: %s\n当前步骤: %s\n步骤描述: %s\n请完成该步骤并给出结果。",
					task.Description, step.Title, step.Description)},
			}, nil)
			if err != nil {
				stepResult = fmt.Sprintf("Step %d AI Error: %v", step.Index, err)
			} else {
				if len(resp.Choices) > 0 {
					if c, ok := resp.Choices[0].Message.Content.(string); ok {
						stepResult = c
					}
				}
			}
		}

		results = append(results, fmt.Sprintf("### %s\n%s", step.Title, stepResult))

		s.db.Model(&models.DigitalEmployeeTask{}).
			Where("execution_id = ?", executionID).
			Updates(map[string]any{
				"progress":           progress,
				"current_step_index": i + 1,
				"result_raw":         strings.Join(results, "\n\n"),
			})

		if task.Status == "approved" {
			task.Status = "executing"
		}
	}

	finalResult := strings.Join(results, "\n\n")
	return s.RecordTaskResult(ctx, executionID, finalResult, true)
}

func (s *DigitalEmployeeTaskServiceImpl) ExecuteStep(ctx context.Context, executionID string, stepIndex int) error {
	return nil
}

func (s *DigitalEmployeeTaskServiceImpl) ApproveTask(ctx context.Context, executionID string) error {
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("execution_id = ?", executionID).
		Update("status", "approved").Error
}

func (s *DigitalEmployeeTaskServiceImpl) CreateSubTask(ctx context.Context, parentExecutionID string, subTask *models.DigitalEmployeeTask) error {
	parent, err := s.GetTaskByExecutionID(ctx, parentExecutionID)
	if err != nil {
		return err
	}

	subTask.ParentTaskID = parent.ID
	if subTask.ExecutionID == "" {
		subTask.ExecutionID = fmt.Sprintf("sub-%s-%d", parentExecutionID, time.Now().UnixNano())
	}

	return s.CreateTask(ctx, subTask)
}

func (s *DigitalEmployeeTaskServiceImpl) RecordTaskResult(ctx context.Context, executionID string, result string, success bool) error {
	status := "completed"
	if !success {
		status = "failed"
		clog.Error("[TaskEngine] Task failed", zap.String("execution_id", executionID))
	} else {
		clog.Info("[TaskEngine] Task completed successfully", zap.String("execution_id", executionID))
	}
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.DigitalEmployeeTask{}).
		Where("execution_id = ?", executionID).
		Updates(map[string]any{
			"result_raw": result,
			"status":     status,
			"end_time":   &now,
			"progress":   100,
		}).Error
}

// --- DigitalEmployeeKPIService Implementation ---

type DigitalEmployeeKPIServiceImpl struct {
	db     *gorm.DB
	aiSvc  types.AIService
	b2bSvc b2b.B2BService
}

var _ DigitalEmployeeKPIService = (*DigitalEmployeeKPIServiceImpl)(nil)

func NewDigitalEmployeeKPIService(db *gorm.DB, aiSvc types.AIService) DigitalEmployeeKPIService {
	return &DigitalEmployeeKPIServiceImpl{
		db:    db,
		aiSvc: aiSvc,
	}
}

func (s *DigitalEmployeeKPIServiceImpl) SetB2BService(b2b b2b.B2BService) {
	s.b2bSvc = b2b
}

func (s *DigitalEmployeeKPIServiceImpl) CalculateKPI(ctx context.Context, employeeID uint) (float64, error) {
	var stats struct {
		TotalTasks     int64
		CompletedTasks int64
		FailedTasks    int64
		AvgTokenUsage  float64
		AvgDuration    float64
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.Model(&models.DigitalEmployeeTask{}).
		Where("assignee_id = ? AND created_at > ?", employeeID, thirtyDaysAgo).
		Count(&stats.TotalTasks)

	if stats.TotalTasks == 0 {
		return 100, nil
	}

	s.db.Model(&models.DigitalEmployeeTask{}).
		Where("assignee_id = ? AND status = ? AND created_at > ?", employeeID, "completed", thirtyDaysAgo).
		Count(&stats.CompletedTasks)

	var avgStats struct {
		AvgTokenUsage float64 `gorm:"column:avg_token_usage"`
		AvgDuration   float64 `gorm:"column:avg_duration"`
	}
	s.db.Model(&models.DigitalEmployeeTask{}).
		Where("assignee_id = ? AND status = ? AND created_at > ?", employeeID, "completed", thirtyDaysAgo).
		Select("COALESCE(AVG(token_usage), 0) as avg_token_usage, COALESCE(AVG(duration), 0) as avg_duration").
		Scan(&avgStats)

	stats.AvgTokenUsage = avgStats.AvgTokenUsage
	stats.AvgDuration = avgStats.AvgDuration

	clog.Debug("[KPI] Calculation stats",
		zap.Uint("employee_id", employeeID),
		zap.Int64("total", stats.TotalTasks),
		zap.Int64("completed", stats.CompletedTasks),
		zap.Float64("avg_token", stats.AvgTokenUsage),
		zap.Float64("avg_duration", stats.AvgDuration))

	completionRate := float64(stats.CompletedTasks) / float64(stats.TotalTasks)
	scoreCompletion := completionRate * 100

	scoreEfficiency := 0.0
	if stats.CompletedTasks > 0 {
		if stats.AvgDuration <= 0 {
			scoreEfficiency = 100.0
		} else if stats.AvgDuration <= 300 {
			scoreEfficiency = 100.0
		} else {
			scoreEfficiency = 100.0 * (300.0 / stats.AvgDuration)
		}
	}

	scoreCost := 0.0
	if stats.CompletedTasks > 0 {
		if stats.AvgTokenUsage <= 0 {
			scoreCost = 100.0
		} else if stats.AvgTokenUsage <= 5000 {
			scoreCost = 100.0
		} else {
			scoreCost = 100.0 * (5000.0 / stats.AvgTokenUsage)
		}
	}

	finalScore := (scoreCompletion * 0.6) + (scoreEfficiency * 0.2) + (scoreCost * 0.2)

	if err := s.db.Model(&models.DigitalEmployee{}).Where("id = ?", employeeID).Update("kpi_score", finalScore).Error; err != nil {
		return finalScore, err
	}

	return finalScore, nil
}

func (s *DigitalEmployeeKPIServiceImpl) OptimizeEmployee(ctx context.Context, employeeID uint) error {
	var emp models.DigitalEmployee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return err
	}

	if s.b2bSvc != nil {
		var dispatch models.DigitalEmployeeDispatch
		err := s.db.Where("employee_id = ? AND status = ?", employeeID, "approved").First(&dispatch).Error
		if err == nil {
			hasPerm, err := s.b2bSvc.CheckDispatchPermission(employeeID, dispatch.TargetEntID, "optimize")
			if err != nil {
				return fmt.Errorf("failed to check B2B permission: %w", err)
			}
			if !hasPerm {
				return fmt.Errorf("no permission to optimize dispatched employee (requires 'optimize' permission)")
			}
		}
	}

	var failedTasks []models.DigitalEmployeeTask
	s.db.Where("assignee_id = ? AND status = ?", employeeID, "failed").
		Order("created_at DESC").
		Limit(5).
		Find(&failedTasks)

	if len(failedTasks) == 0 {
		return nil
	}

	taskData, _ := json.Marshal(failedTasks)
	prompt := fmt.Sprintf(`你是一名数字员工绩效专家。
当前数字员工: %s (职位: %s)
最近失败的任务记录: %s
当前个人简介/人设: %s

请分析失败原因，并生成一个优化的“个人简介/Bio”或“执行策略建议”，以提高其后续任务成功率。
请直接输出优化后的 Bio 内容，不要包含其他解释文字。`,
		emp.Name, emp.Title, string(taskData), emp.Bio)

	var chatModel models.AIModel
	if err := s.db.Where("is_default = ?", true).First(&chatModel).Error; err != nil {
		if err := s.db.First(&chatModel).Error; err != nil {
			return fmt.Errorf("no AI models configured: %w", err)
		}
	}

	resp, err := s.aiSvc.Chat(ctx, chatModel.ID, []types.Message{
		{Role: types.RoleSystem, Content: "你是一名数字员工绩效专家。"},
		{Role: types.RoleUser, Content: prompt},
	}, nil)
	if err != nil {
		return err
	}

	optimizedBio, _ := resp.Choices[0].Message.Content.(string)

	clog.Info("[KPI] Employee optimization completed",
		zap.Uint("employee_id", employeeID),
		zap.String("name", emp.Name),
		zap.Uint("enterprise_id", emp.EnterpriseID))

	return s.db.Model(&models.DigitalEmployee{}).Where("id = ?", employeeID).Update("bio", optimizedBio).Error
}

func (s *DigitalEmployeeKPIServiceImpl) GetPerformanceReport(ctx context.Context, employeeID uint, days int) (string, error) {
	var emp models.DigitalEmployee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return "", err
	}

	score, _ := s.CalculateKPI(ctx, employeeID)

	var stats struct {
		Total         int64
		Completed     int64
		Failed        int64
		AvgTokenUsage float64
		AvgDuration   float64
	}
	startTime := time.Now().AddDate(0, 0, -days)
	s.db.Model(&models.DigitalEmployeeTask{}).Where("assignee_id = ? AND created_at > ?", employeeID, startTime).Count(&stats.Total)
	s.db.Model(&models.DigitalEmployeeTask{}).Where("assignee_id = ? AND status = ? AND created_at > ?", employeeID, "completed", startTime).Count(&stats.Completed)
	s.db.Model(&models.DigitalEmployeeTask{}).Where("assignee_id = ? AND status = ? AND created_at > ?", employeeID, "failed", startTime).Count(&stats.Failed)

	var avgStats struct {
		AvgTokenUsage float64 `gorm:"column:avg_token_usage"`
		AvgDuration   float64 `gorm:"column:avg_duration"`
	}
	s.db.Model(&models.DigitalEmployeeTask{}).
		Where("assignee_id = ? AND status = ? AND created_at > ?", employeeID, "completed", startTime).
		Select("COALESCE(AVG(token_usage), 0) as avg_token_usage, COALESCE(AVG(duration), 0) as avg_duration").
		Scan(&avgStats)

	stats.AvgTokenUsage = avgStats.AvgTokenUsage
	stats.AvgDuration = avgStats.AvgDuration

	report := fmt.Sprintf("### 数字员工绩效报告: %s\n", emp.Name)
	report += fmt.Sprintf("- **职位**: %s\n", emp.Title)
	report += fmt.Sprintf("- **当前 KPI 分数**: %.2f\n", score)
	report += fmt.Sprintf("- **任务统计 (过去 %d 天)**: 总数 %d | 成功 %d | 失败 %d\n", days, stats.Total, stats.Completed, stats.Failed)
	report += fmt.Sprintf("- **平均执行效率**: %.1f 秒/任务\n", stats.AvgDuration)
	report += fmt.Sprintf("- **平均资源消耗**: %.0f Token/任务\n", stats.AvgTokenUsage)
	report += fmt.Sprintf("- **累计消耗 Token**: %d / %d (预算)\n", emp.SalaryToken, emp.SalaryLimit)

	var tasks []models.DigitalEmployeeTask
	s.db.Where("assignee_id = ? AND status = ? AND plan_raw != ''", employeeID, "completed").Limit(20).Find(&tasks)
	toolCounts := make(map[string]int)
	for _, t := range tasks {
		var plan struct {
			Steps []struct {
				Tool string `json:"tool"`
			} `json:"steps"`
		}
		if err := json.Unmarshal([]byte(t.PlanRaw), &plan); err == nil {
			for _, s := range plan.Steps {
				if s.Tool != "" {
					toolCounts[s.Tool]++
				}
			}
		}
	}

	if len(toolCounts) > 0 {
		report += "\n**核心技能分布**:\n"
		for tool, count := range toolCounts {
			report += fmt.Sprintf("- `%s`: 使用 %d 次\n", tool, count)
		}
	}

	if score < 80 {
		report += "\n> **管理建议**: 该员工近期表现欠佳，建议执行 `OptimizeEmployee` 逻辑进行 AI 调优，或人工介入调整其权限与任务指派逻辑。"
	} else {
		report += "\n> **管理建议**: 该员工表现优异，建议维持当前配置。可考虑委派更具挑战性的跨部门任务。"
	}

	return report, nil
}
