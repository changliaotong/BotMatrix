package ai

import (
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HandleKnowledgeUpload 处理知识文件上传
// @Summary 上传知识文件
// @Description 上传文件并将其索引到知识库中
// @Tags Knowledge
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "知识文件"
// @Param type formData string false "文档类型 (doc, code, manual, etc.)"
// @Param target_type formData string false "授权对象类型 (user, group, system)"
// @Param target_id formData string false "授权对象 ID"
// @Success 200 {object} utils.JSONResponse "上传成功"
// @Router /api/knowledge/upload [post]
func HandleKnowledgeUpload(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取知识库
		kbBase := m.GetKnowledgeBase()
		if kbBase == nil {
			utils.SendJSONResponse(w, false, "知识库未初始化", nil)
			return
		}
		kb, ok := kbBase.(*rag.PostgresKnowledgeBase)
		if !ok {
			utils.SendJSONResponse(w, false, "知识库类型不匹配", nil)
			return
		}

		// 2. 解析表单
		err := r.ParseMultipartForm(50 << 20) // 50MB max
		if err != nil {
			utils.SendJSONResponse(w, false, "解析表单失败: "+err.Error(), nil)
			return
		}

		docType := r.FormValue("type")
		if docType == "" {
			docType = "doc"
		}
		targetType := r.FormValue("target_type")
		targetID := r.FormValue("target_id")

		// --- 优化：支持显式的 bot_id 和 group_id ---
		botID := r.FormValue("bot_id")
		groupID := r.FormValue("group_id")

		if targetType == "" {
			if botID != "" {
				targetType = "bot"
				targetID = botID
			} else if groupID != "" {
				targetType = "group"
				targetID = groupID
			}
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			utils.SendJSONResponse(w, false, "读取文件失败: "+err.Error(), nil)
			return
		}
		defer file.Close()

		// 获取用户信息
		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		uploaderID := "system"
		if claims != nil {
			uploaderID = fmt.Sprintf("%d", claims.UserID)
		}

		// 如果仍然没有指定 target，默认授权给上传者
		if targetType == "" {
			targetType = "user"
			targetID = uploaderID
		}

		// 3. 读取内容并索引
		content, err := io.ReadAll(file)
		if err != nil {
			utils.SendJSONResponse(w, false, "读取文件内容失败: "+err.Error(), nil)
			return
		}

		filename := header.Filename
		indexer := rag.NewIndexer(kb, m.GetAIService(), 0)

		// 异步执行索引任务，避免前端超时
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// source 使用 uuid 确保唯一性，title 使用原始文件名
			source := "upload://" + uuid.New().String() + "/" + filename
			err := indexer.IndexContent(ctx, filename, source, content, docType, uploaderID, targetType, targetID)
			if err != nil {
				fmt.Printf("[Knowledge] Index failed: %v\n", err)
				return
			}

			// --- 优化：如果同时提供了多个 ID，则都进行绑定 ---
			// 先找到刚刚创建的文档 ID (IndexContent 内部没有返回，我们需要查询)
			var doc rag.KnowledgeDoc
			db := m.GetGORMDB()
			if err := db.Where("source = ?", source).First(&doc).Error; err == nil {
				if botID != "" && (targetType != "bot" || targetID != botID) {
					kb.AddDocAccess(ctx, doc.ID, "bot", botID)
				}
				if groupID != "" && (targetType != "group" || targetID != groupID) {
					kb.AddDocAccess(ctx, doc.ID, "group", groupID)
				}
			}

			fmt.Printf("[Knowledge] Indexed file: %s\n", filename)
		}()

		utils.SendJSONResponse(w, true, "文件已接收，后台正在进行解析与向量化处理", nil)
	}
}

// HandleKnowledgeList 获取知识文档列表
// @Summary 获取知识文档列表
// @Description 获取当前用户有权访问的所有知识文档
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param q query string false "搜索关键词"
// @Param type query string false "文档类型"
// @Success 200 {object} utils.JSONResponse "文档列表"
// @Router /api/knowledge/list [get]
func HandleKnowledgeList(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		docType := r.URL.Query().Get("type")

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := ""
		isAdmin := false
		if claims != nil {
			userID = fmt.Sprintf("%d", claims.UserID)
			isAdmin = claims.IsAdmin
		}

		var docs []rag.KnowledgeDoc
		db := m.GetGORMDB().Model(&rag.KnowledgeDoc{})

		// 权限控制：管理员看所有，普通用户只能看授权给自己的或系统的
		if !isAdmin {
			// 关联权限表查询
			db = db.Joins("JOIN knowledge_doc_access ON knowledge_docs.id = knowledge_doc_access.doc_id").
				Where("(knowledge_doc_access.owner_type = 'user' AND knowledge_doc_access.owner_id = ?) OR knowledge_doc_access.owner_type = 'system'", userID).
				Distinct()
		}

		if query != "" {
			db = db.Where("title LIKE ?", "%"+query+"%")
		}
		if docType != "" {
			db = db.Where("type = ?", docType)
		}

		if err := db.Order("updated_at DESC").Find(&docs).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取列表失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", docs)
	}
}

// HandleKnowledgeDelete 删除知识文档
// @Summary 删除知识文档
// @Description 删除指定的知识文档及其所有切片
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param id path int true "文档 ID"
// @Success 200 {object} utils.JSONResponse "删除成功"
// @Router /api/knowledge/delete/{id} [delete]
func HandleKnowledgeDelete(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/knowledge/delete/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的 ID", nil)
			return
		}

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := ""
		isAdmin := false
		if claims != nil {
			userID = fmt.Sprintf("%d", claims.UserID)
			isAdmin = claims.IsAdmin
		}

		// 检查权限：只有上传者或管理员可以删除
		var doc rag.KnowledgeDoc
		db := m.GetGORMDB()
		if err := db.First(&doc, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "未找到文档", nil)
			return
		}

		if !isAdmin && doc.UploaderID != userID {
			utils.SendJSONResponse(w, false, "没有权限删除此文档", nil)
			return
		}

		// 开启事务删除
		err := db.Transaction(func(tx *gorm.DB) error {
			// 1. 删除切片
			if err := tx.Where("doc_id = ?", id).Delete(&rag.KnowledgeChunk{}).Error; err != nil {
				return err
			}
			// 2. 删除权限记录
			if err := tx.Where("doc_id = ?", id).Delete(&rag.KnowledgeDocAccess{}).Error; err != nil {
				return err
			}
			// 3. 删除图谱关系 (如果存在)
			tx.Exec("DELETE FROM knowledge_relations WHERE doc_id = ?", id)
			// 4. 删除文档本身
			if err := tx.Delete(&doc).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// HandleKnowledgeDetail 获取知识文档详情 (包括切片)
// @Summary 获取文档详情
// @Description 获取知识文档的元数据和所有切片内容
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param id path int true "文档 ID"
// @Success 200 {object} utils.JSONResponse "详情数据"
// @Router /api/knowledge/detail/{id} [get]
func HandleKnowledgeDetail(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/knowledge/detail/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的 ID", nil)
			return
		}

		var doc rag.KnowledgeDoc
		db := m.GetGORMDB()
		if err := db.Preload("Accesses").First(&doc, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "未找到文档", nil)
			return
		}

		var chunks []rag.KnowledgeChunk
		db.Where("doc_id = ?", id).Order("id ASC").Find(&chunks)

		utils.SendJSONResponse(w, true, "", struct {
			Doc    rag.KnowledgeDoc     `json:"doc"`
			Chunks []rag.KnowledgeChunk `json:"chunks"`
		}{
			Doc:    doc,
			Chunks: chunks,
		})
	}
}
