package rag

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Vector 向量类型，支持 GORM 序列化
type Vector []float32

func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	// PostgreSQL pgvector 期望格式为 [1.1,2.2,3.3]
	res := "["
	for i, f := range v {
		if i > 0 {
			res += ","
		}
		res += fmt.Sprintf("%g", f)
	}
	res += "]"
	return res, nil
}

func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var data []byte
	switch s := value.(type) {
	case []byte:
		data = s
	case string:
		data = []byte(s)
	default:
		return fmt.Errorf("invalid type for Vector: %T", value)
	}
	return json.Unmarshal(data, v)
}

// KnowledgeDoc 知识文档
type KnowledgeDoc struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Source     string         `gorm:"size:512;uniqueIndex" json:"source"` // 文件路径或 URL
	Type       string         `gorm:"size:50" json:"type"`                // doc, code, faq, skill
	Content    string         `gorm:"type:text" json:"content"`
	Hash       string         `gorm:"size:64" json:"hash"`                          // 内容哈希，用于检测更新
	UploaderID string         `gorm:"size:64;index" json:"uploader_id"`             // 上传者 ID (拥有管理权)
	Status     string         `gorm:"size:20;index;default:'active'" json:"status"` // active, paused
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联权限控制
	Accesses []KnowledgeDocAccess `gorm:"foreignKey:DocID" json:"accesses,omitempty"`
}

// KnowledgeDocAccess 知识文档访问权限 (实现多群/多人共享)
type KnowledgeDocAccess struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DocID     uint      `gorm:"index" json:"doc_id"`
	OwnerType string    `gorm:"size:20;index" json:"owner_type"` // system, user, group
	OwnerID   string    `gorm:"size:64;index" json:"owner_id"`   // 用户 ID 或群组 ID
	CreatedAt time.Time `json:"created_at"`
}

func (KnowledgeDocAccess) TableName() string {
	return "knowledge_doc_access"
}

// KnowledgeChunk 知识切片 (用于向量搜索)
type KnowledgeChunk struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	DocID     uint          `gorm:"index" json:"doc_id"`
	Doc       *KnowledgeDoc `gorm:"foreignKey:DocID" json:"doc,omitempty"`
	Content   string        `gorm:"type:text" json:"content"`
	Embedding Vector        `gorm:"type:vector(2048)" json:"-"` // Doubao-embedding-vision (Default 2048)
	Metadata  string        `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time     `json:"created_at"`
}

// TableName 指定表名
func (KnowledgeChunk) TableName() string {
	return "knowledge_chunks"
}

func (KnowledgeDoc) TableName() string {
	return "knowledge_docs"
}

// KnowledgeEntity 知识实体 (GraphRAG)
type KnowledgeEntity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;index" json:"name"`
	Type        string    `gorm:"size:100;index" json:"type"`
	Description string    `gorm:"type:text" json:"description"`
	Embedding   Vector    `gorm:"type:vector(2048)" json:"-"`
	Metadata    string    `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (KnowledgeEntity) TableName() string {
	return "knowledge_entities"
}

// KnowledgeRelation 知识关系 (GraphRAG)
type KnowledgeRelation struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	SubjectID   uint             `gorm:"index" json:"subject_id"`
	Subject     *KnowledgeEntity `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Predicate   string           `gorm:"size:255;index" json:"predicate"`
	ObjectID    uint             `gorm:"index" json:"object_id"`
	Object      *KnowledgeEntity `gorm:"foreignKey:ObjectID" json:"object,omitempty"`
	Description string           `gorm:"type:text" json:"description"`
	DocID       uint             `gorm:"index" json:"doc_id"`
	Doc         *KnowledgeDoc    `gorm:"foreignKey:DocID" json:"doc,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

func (KnowledgeRelation) TableName() string {
	return "knowledge_relations"
}
