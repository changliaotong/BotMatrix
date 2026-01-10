package app

import (
	"BotMatrix/common/bot"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"BotMatrix/common/utils"
	"net/http"
	"time"
)

// HandleToolAuditActions 处理工具审计相关的通用接口 (GET: 列表, POST: 审批/拒绝)
func HandleToolAuditActions(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetToolAuditLogs(w, r)
		case http.MethodPost:
			// 根据 URL 后缀判断操作
			if r.URL.Path == "/api/admin/audit/tools/approve" {
				handleApproveToolCall(w, r)
			} else if r.URL.Path == "/api/admin/audit/tools/reject" {
				handleRejectToolCall(w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// handleGetToolAuditLogs 获取工具调用审计日志列表
func handleGetToolAuditLogs(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	query := db.Model(&models.ToolAuditLog{})

	status := r.URL.Query().Get("status")
	if status != "" {
		query = query.Where("Status = ?", status)
	}

	risk := r.URL.Query().Get("risk")
	if risk != "" {
		query = query.Where("RiskLevel = ?", risk)
	}

	var total int64
	query.Count(&total)

	page := utils.ParseInt(r.URL.Query().Get("page"), 1)
	limit := utils.ParseInt(r.URL.Query().Get("limit"), 20)
	offset := (page - 1) * limit

	var logs []models.ToolAuditLog
	query.Order("CreateTime DESC").Limit(limit).Offset(offset).Find(&logs)

	utils.WriteJSON(w, http.StatusOK, utils.JSONResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"total": total,
			"page":  page,
			"limit": limit,
			"list":  logs,
		},
	})
}

// handleApproveToolCall 审批通过工具调用
func handleApproveToolCall(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.JSONResponse{Code: 400, Message: "Missing id"})
		return
	}

	db := database.GetDB()
	var log models.ToolAuditLog
	if err := db.First(&log, "Id = ?", id).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.JSONResponse{Code: 404, Message: "Log not found"})
		return
	}

	if log.Status != "PendingApproval" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.JSONResponse{Code: 400, Message: "Only logs in PendingApproval status can be approved"})
		return
	}

	now := time.Now()
	log.Status = "Approved"
	log.ApprovedAt = &now
	log.ApprovedBy = "admin"

	if err := db.Save(&log).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.JSONResponse{Code: 500, Message: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JSONResponse{Code: 200, Message: "Approved"})
}

// handleRejectToolCall 拒绝工具调用
func handleRejectToolCall(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.JSONResponse{Code: 400, Message: "Missing id"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.JSONResponse{Code: 400, Message: "Invalid request body"})
		return
	}

	db := database.GetDB()
	var log models.ToolAuditLog
	if err := db.First(&log, "Id = ?", id).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.JSONResponse{Code: 404, Message: "Log not found"})
		return
	}

	if log.Status != "PendingApproval" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.JSONResponse{Code: 400, Message: "Only logs in PendingApproval status can be rejected"})
		return
	}

	now := time.Now()
	log.Status = "Rejected"
	log.ApprovedAt = &now
	log.ApprovedBy = "admin"
	log.RejectionReason = req.Reason

	if err := db.Save(&log).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.JSONResponse{Code: 500, Message: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JSONResponse{Code: 200, Message: "Rejected"})
}
