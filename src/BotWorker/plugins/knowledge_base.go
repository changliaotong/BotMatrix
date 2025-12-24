package plugins

import (
	"BotMatrix/common"
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type KnowledgeBasePlugin struct {
	db              *sql.DB
	officialGroupID string
}

func NewKnowledgeBasePlugin(database *sql.DB, officialGroupID string) *KnowledgeBasePlugin {
	return &KnowledgeBasePlugin{
		db:              database,
		officialGroupID: officialGroupID,
	}
}

func (p *KnowledgeBasePlugin) Name() string {
	return "knowledge_base"
}

func (p *KnowledgeBasePlugin) Description() string {
	return common.T("", "knowledge_base_desc|知识库插件，支持自定义问答")
}

func (p *KnowledgeBasePlugin) Version() string {
	return "1.0.0"
}

func (p *KnowledgeBasePlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "knowledge_base_no_db|知识库插件：未配置数据库，功能已禁用"))
		return
	}

	log.Println(common.T("", "knowledge_base_loaded|知识库插件已加载"))

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "knowledge_base") {
			return nil
		}

		if p.handleTeach(robot, event) {
			return nil
		}

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "knowledge_base") {
			return nil
		}

		if p.isTeachPattern(event.RawMessage) {
			return nil
		}

		if p.handleAsk(robot, event) {
			return nil
		}

		return nil
	})
}

func (p *KnowledgeBasePlugin) isTeachPattern(text string) bool {
	if text == "" {
		return false
	}
	s := strings.TrimSpace(text)
	re := regexp.MustCompile(`(?s)^[问Qq][:： ]+(.+?)[\s]+[答Aa][:： ]+(.+)$`)
	return re.MatchString(s)
}

func (p *KnowledgeBasePlugin) handleTeach(robot plugin.Robot, event *onebot.Event) bool {
	if event == nil {
		return false
	}

	text := strings.TrimSpace(event.RawMessage)
	if text == "" {
		return false
	}

	re := regexp.MustCompile(`(?s)^[问Qq][:： ]+(.+?)[\s]+[答Aa][:： ]+(.+)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 3 {
		return false
	}

	questionRaw := strings.TrimSpace(matches[1])
	answerRaw := strings.TrimSpace(matches[2])
	if questionRaw == "" || answerRaw == "" {
		p.sendMessage(robot, event, common.T("", "knowledge_base_empty_qa|问题或回答不能为空哦！"))
		return true
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", event.UserID)

	points, err := db.GetPoints(p.db, userIDStr)
	if err != nil {
		log.Printf(common.T("", "knowledge_base_get_points_failed|获取用户积分失败"), err)
		p.sendMessage(robot, event, common.T("", "knowledge_base_query_points_failed|查询积分失败，请稍后再试"))
		return true
	}

	if points < 0 {
		p.sendMessage(robot, event, common.T("", "knowledge_base_negative_points|你的积分不足，无法进行教学"))
		return true
	}

	containsSensitive := p.containsSensitive(questionRaw) || p.containsSensitive(answerRaw)
	containsAd := p.containsAdvertisement(questionRaw) || p.containsAdvertisement(answerRaw)

	deduct := -10
	if containsSensitive {
		deduct = -50
	}
	if containsAd {
		deduct = -100
	}

	needReview := containsSensitive || containsAd

	isPrivileged := false
	if GlobalDB != nil && groupIDStr != "" && userIDStr != "" {
		if ok, err := db.IsSuperAdmin(GlobalDB, groupIDStr, userIDStr); err == nil && ok {
			isPrivileged = true
		}
		if !isPrivileged {
			if ok, err := db.IsUserInGroupWhitelist(GlobalDB, groupIDStr, userIDStr); err == nil && ok {
				isPrivileged = true
			}
		}
	}

	status := "approved"
	if needReview && !isPrivileged {
		status = "pending"
	}

	normalized := NormalizeQuestion(questionRaw)
	if normalized == "" {
		p.sendMessage(robot, event, common.T("", "knowledge_base_invalid_format|无效的问题格式"))
		return true
	}

	if deduct != 0 {
		if err := db.AddPoints(p.db, userIDStr, deduct, common.T("", "knowledge_base_teach_reason|知识库教学消耗"), "teach"); err != nil {
			log.Printf(common.T("", "knowledge_base_deduct_points_failed|扣除积分失败"), err)
			p.sendMessage(robot, event, common.T("", "knowledge_base_points_failed_not_effective|操作失败：积分服务不可用"))
			return true
		}
	}

	q := &db.Question{
		GroupID:            groupIDStr,
		QuestionRaw:        questionRaw,
		QuestionNormalized: normalized,
		Status:             status,
		CreatedBy:          userIDStr,
		SourceGroupID:      groupIDStr,
	}

	createdQuestion, err := db.CreateQuestion(p.db, q)
	if err != nil {
		log.Printf(common.T("", "knowledge_base_create_question_failed|创建问题记录失败"), err)
		p.sendMessage(robot, event, common.T("", "knowledge_base_save_question_failed|保存问题失败，请重试"))
		return true
	}

	answer := &db.Answer{
		QuestionID: createdQuestion.ID,
		Answer:     answerRaw,
		Status:     status,
		CreatedBy:  userIDStr,
	}

	_, err = db.AddAnswer(p.db, answer)
	if err != nil {
		log.Printf(common.T("", "knowledge_base_create_answer_failed|创建回答记录失败"), err)
		p.sendMessage(robot, event, common.T("", "knowledge_base_save_answer_failed|保存回答失败，请重试"))
		return true
	}

	reply := common.T("", "knowledge_base_teach_success|教学成功！我已经记住啦。")
	if needReview && !isPrivileged {
		reply = common.T("", "knowledge_base_teach_pending|教学已提交，内容包含敏感词或广告，需审核后生效。")
	}

	p.sendMessage(robot, event, reply)
	return true
}

func (p *KnowledgeBasePlugin) handleAsk(robot plugin.Robot, event *onebot.Event) bool {
	if event == nil {
		return false
	}

	text := event.RawMessage
	if text == "" {
		if msg, ok := event.Message.(string); ok {
			text = msg
		}
	}
	if text == "" {
		return false
	}

	s := strings.TrimSpace(text)
	if s == "" {
		return false
	}

	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "／") {
		return false
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", event.UserID)

	clean := s
	isAt := IsAtMe(event)
	if isAt {
		selfAt := fmt.Sprintf("[CQ:at,qq=%d]", event.SelfID)
		clean = strings.ReplaceAll(clean, selfAt, "")
		clean = strings.TrimSpace(clean)
		if clean == "" {
			return false
		}
	}

	normalized := NormalizeQuestion(clean)
	if normalized == "" {
		return false
	}

	if isAt {
		if lastAnswerID, err := db.GetGroupLastAnswerID(p.db, groupIDStr); err != nil {
			log.Printf(common.T("", "knowledge_base_get_last_answer_failed|获取最后回答记录失败"), err)
		} else if lastAnswerID > 0 {
			if err := db.IncrementAnswerShortIntervalUsageIfRecent(p.db, lastAnswerID); err != nil {
				log.Printf(common.T("", "knowledge_base_update_short_interval_failed|更新短间隔使用统计失败"), err)
			}
		}

		q, err := db.GetQuestionByGroupAndNormalized(p.db, groupIDStr, normalized)
		if err != nil {
			log.Printf(common.T("", "knowledge_base_get_question_stats_failed|获取问题统计失败"), err)
		} else {
			if q == nil {
				newQ := &db.Question{
					GroupID:            groupIDStr,
					QuestionRaw:        clean,
					QuestionNormalized: normalized,
					Status:             "unanswered",
					CreatedBy:          userIDStr,
					SourceGroupID:      groupIDStr,
				}
				if created, err := db.CreateQuestion(p.db, newQ); err != nil {
					log.Printf(common.T("", "knowledge_base_create_question_stats_failed|创建问题统计记录失败"), err)
				} else if created != nil {
					q = created
				}
			}
			if q != nil {
				if err := db.IncrementQuestionUsage(p.db, q.ID); err != nil {
					log.Printf(common.T("", "knowledge_base_update_question_usage_failed|更新问题使用统计失败"), err)
				}
			}
		}
	}

	mode, err := db.GetGroupQAMode(p.db, groupIDStr)
	if err != nil {
		log.Printf(common.T("", "knowledge_base_get_qa_mode_failed|获取群问答模式失败"), err)
	}
	if mode == "" {
		mode = "group"
	}

	answerText, _, answerID, ok := p.findAnswer(normalized, groupIDStr, mode, isAt)
	if !ok || answerText == "" {
		return false
	}

	if isAt {
		if err := db.IncrementAnswerUsage(p.db, answerID); err != nil {
			log.Printf(common.T("", "knowledge_base_update_answer_usage_failed|更新回答使用统计失败"), err)
		}
		if err := db.SetGroupLastAnswerID(p.db, groupIDStr, answerID); err != nil {
			log.Printf(common.T("", "knowledge_base_set_last_answer_failed|设置最后回答记录失败"), err)
		}
	}

	final := SubstituteAllVariables(answerText, event)

	p.sendMessage(robot, event, final)
	return true
}

func (p *KnowledgeBasePlugin) findAnswer(normalized string, groupID string, mode string, isAt bool) (string, int, int, bool) {
	if normalized == "" || groupID == "" {
		return "", 0, 0, false
	}

	effectiveMode := mode
	if isAt && (mode == "official" || mode == "chatty") {
		effectiveMode = "ultimate"
	}

	switch effectiveMode {
	case "silent":
		return "", 0, 0, false
	case "group":
		return p.findFromGroups(normalized, []string{groupID})
	case "official":
		return p.findFromGroups(normalized, []string{groupID})
	case "chatty":
		groups := []string{groupID}
		if p.officialGroupID != "" {
			groups = append(groups, p.officialGroupID)
		}
		return p.findFromGroups(normalized, groups)
	case "ultimate":
		return p.findUltimate(normalized, groupID)
	default:
		return p.findFromGroups(normalized, []string{groupID})
	}
}

func (p *KnowledgeBasePlugin) findFromGroups(normalized string, groups []string) (string, int, int, bool) {
	for _, gid := range groups {
		if gid == "" {
			continue
		}
		q, err := db.GetQuestionByGroupAndNormalized(p.db, gid, normalized)
		if err != nil {
			log.Printf(common.T("", "knowledge_base_query_question_failed|查询问题失败"), err)
			continue
		}
		if q == nil || q.Status != "approved" {
			continue
		}
		answer, err := db.GetRandomApprovedAnswer(p.db, q.ID)
		if err != nil {
			log.Printf(common.T("", "knowledge_base_query_answer_failed|查询回答失败"), err)
			continue
		}
		if answer == nil || answer.Status != "approved" || answer.Answer == "" {
			continue
		}
		return answer.Answer, q.ID, answer.ID, true
	}
	return "", 0, 0, false
}

func (p *KnowledgeBasePlugin) findUltimate(normalized string, currentGroupID string) (string, int, int, bool) {
	if normalized == "" {
		return "", 0, 0, false
	}

	if ans, qID, aID, ok := p.findFromGroups(normalized, []string{currentGroupID}); ok {
		return ans, qID, aID, true
	}

	if p.officialGroupID != "" {
		if ans, qID, aID, ok := p.findFromGroups(normalized, []string{p.officialGroupID}); ok {
			return ans, qID, aID, true
		}
	}

	query := `
	SELECT id, group_id
	FROM questions
	WHERE question_normalized = $1 AND status = 'approved'
	LIMIT 20
	`

	rows, err := p.db.Query(query, normalized)
	if err != nil {
		log.Printf("查询跨群问题失败: %v", err)
		return "", 0, 0, false
	}
	defer rows.Close()

	type candidate struct {
		id      int
		groupID string
	}

	var candidates []candidate
	for rows.Next() {
		var id int
		var gid string
		if err := rows.Scan(&id, &gid); err != nil {
			log.Printf("扫描跨群问题失败: %v", err)
			continue
		}
		if gid == currentGroupID {
			continue
		}
		if p.officialGroupID != "" && gid == p.officialGroupID {
			continue
		}
		candidates = append(candidates, candidate{id: id, groupID: gid})
	}

	for _, c := range candidates {
		answer, err := db.GetRandomApprovedAnswer(p.db, c.id)
		if err != nil {
			log.Printf("查询跨群答案失败: %v", err)
			continue
		}
		if answer == nil || answer.Status != "approved" || answer.Answer == "" {
			continue
		}
		return answer.Answer, c.id, answer.ID, true
	}

	return "", 0, 0, false
}

func (p *KnowledgeBasePlugin) containsSensitive(text string) bool {
	if text == "" {
		return false
	}
	words := []string{"脏话", "傻逼", "SB", "傻b", "垃圾"}
	s := strings.ToLower(text)
	for _, w := range words {
		if strings.Contains(s, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

func (p *KnowledgeBasePlugin) containsAdvertisement(text string) bool {
	if text == "" {
		return false
	}
	words := []string{"广告", "推广", "促销", "优惠", "打折", "VX", "微信", "QQ", "群号", "购买", "出售", "出货"}
	s := strings.ToLower(text)
	for _, w := range words {
		if strings.Contains(s, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

func (p *KnowledgeBasePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}
