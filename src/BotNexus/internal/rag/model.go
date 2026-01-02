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
	// 对于 PostgreSQL pgvector，它期望格式如 [1,2,3]
	// 对于 SQLite，我们可以存为 JSON 字符串
	return json.Marshal(v)
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
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Source    string         `gorm:"size:512" json:"source"` // 文件路径或 URL
	Type      string         `gorm:"size:50" json:"type"`    // doc, code, faq
	Content   string         `gorm:"type:text" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// KnowledgeChunk 知识切片 (用于向量搜索)
type KnowledgeChunk struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	DocID     uint          `gorm:"index" json:"doc_id"`
	Doc       *KnowledgeDoc `gorm:"foreignKey:DocID" json:"doc,omitempty"`
	Content   string        `gorm:"type:text" json:"content"`
	Embedding Vector        `gorm:"type:vector(1536)" json:"-"` // OpenAI embedding size is 1536
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
