package models

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Cov represents a Column-Value pair, used for dynamic updates
type Cov struct {
	Name  string
	Value interface{}
}

// ==========================================
// Extension Methods Compatibility Layer (ExtAs / Ext)
// ==========================================

// Sz84AsLong converts any value to int64 safely
func Sz84AsLong(val interface{}, def ...int64) int64 {
	var defaultVal int64 = 0
	if len(def) > 0 {
		defaultVal = def[0]
	}

	if val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return defaultVal
		}
		if res, err := strconv.ParseInt(v, 10, 64); err == nil {
			return res
		}
	}
	// Try string conversion for other types
	str := fmt.Sprintf("%v", val)
	if res, err := strconv.ParseInt(str, 10, 64); err == nil {
		return res
	}
	return defaultVal
}

// Sz84AsInt converts any value to int safely
func Sz84AsInt(val interface{}, def ...int) int {
	var defaultVal int = 0
	if len(def) > 0 {
		defaultVal = def[0]
	}
	return int(Sz84AsLong(val, int64(defaultVal)))
}

// Sz84AsString converts any value to string safely
func Sz84AsString(val interface{}, def ...string) string {
	var defaultVal string = ""
	if len(def) > 0 {
		defaultVal = def[0]
	}

	if val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return fmt.Sprintf("%v", val)
}

// Sz84AsBool converts any value to bool safely
func Sz84AsBool(val interface{}, def ...bool) bool {
	var defaultVal bool = false
	if len(def) > 0 {
		defaultVal = def[0]
	}

	if val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case bool:
		return v
	case int, int64, int32:
		return Sz84AsLong(v) != 0
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "true" || v == "1" || v == "yes" || v == "on" {
			return true
		}
		if v == "false" || v == "0" || v == "no" || v == "off" || v == "" {
			return false
		}
	}
	return defaultVal
}

// Sz84IsNull checks if a string is null or empty/whitespace
func Sz84IsNull(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}

// Sz84In checks if a string exists in a list of strings
func Sz84In(val string, list ...string) bool {
	for _, item := range list {
		if strings.EqualFold(val, item) {
			return true
		}
	}
	return false
}

// Sz84GetRandom returns a random string from a list
func Sz84GetRandom(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[rand.Intn(len(list))]
}

// Sz84AsTime formats time in a human-readable way (Today, Yesterday, etc.)
func Sz84AsTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	days := int(diff.Hours() / 24)

	if days == 0 {
		if now.Day() == t.Day() {
			return t.Format("15:04")
		}
		return "昨天 " + t.Format("15:04")
	} else if days == 1 {
		return "昨天 " + t.Format("15:04")
	} else if days == 2 {
		return "前天 " + t.Format("15:04")
	} else if days < 7 {
		return t.Format("Monday 15:04") // Go's layout for Weekday
	}
	return t.Format("2006-01-02 15:04")
}

// ==========================================
// MetaData Compatibility Layer (Database Ops)
// ==========================================

// MetaDataHelpers provides common DB operations mimicking C# MetaData<T>
type MetaDataHelpers struct {
	DB *gorm.DB
}

// NewMetaDataHelpers creates a new helper instance
func NewMetaDataHelpers(db *gorm.DB) *MetaDataHelpers {
	return &MetaDataHelpers{DB: db}
}

// SqlPlus increments a numeric column
// C#: SqlPlus("Credit=Credit+1", id)
// Go: SqlPlus(&User{}, id, "Credit", 1)
func (h *MetaDataHelpers) SqlPlus(model interface{}, id int64, column string, val int) error {
	return h.DB.Model(model).Where("Id = ?", id).Update(column, gorm.Expr(column+" + ?", val)).Error
}

// SetValue updates a single column
// C#: SetValue("Name", "NewName", id)
// Go: SetValue(&User{}, id, "Name", "NewName")
func (h *MetaDataHelpers) SetValue(model interface{}, id int64, column string, val interface{}) error {
	return h.DB.Model(model).Where("Id = ?", id).Update(column, val).Error
}

// GetLong retrieves a single long value
// C#: GetLong("Credit", id)
// Go: GetLong(&User{}, id, "Credit")
func (h *MetaDataHelpers) GetLong(model interface{}, id int64, column string) int64 {
	var result sql.NullInt64
	err := h.DB.Model(model).Where("Id = ?", id).Select(column).Scan(&result).Error
	if err != nil || !result.Valid {
		return 0
	}
	return result.Int64
}

// GetString retrieves a single string value
func (h *MetaDataHelpers) GetString(model interface{}, id int64, column string) string {
	var result sql.NullString
	err := h.DB.Model(model).Where("Id = ?", id).Select(column).Scan(&result).Error
	if err != nil || !result.Valid {
		return ""
	}
	return result.String
}

// QueryRes executes a raw SQL query and formats the result
// C#: QueryRes("select count(*) from Table", "Count: {0}")
func (h *MetaDataHelpers) QueryRes(sqlStr string, format string, args ...interface{}) string {
	var result interface{}
	// This is a simplified version. In C#, QueryRes handles complex formatting and list iteration.
	// Here we assume scalar result for simplicity as a start.
	err := h.DB.Raw(sqlStr, args...).Scan(&result).Error
	if err != nil {
		return ""
	}
	
	// Basic formatting: replace {0} with value
	// For more complex C# string.Format behavior, we might need a regex replacer
	resStr := fmt.Sprintf("%v", result)
	if strings.Contains(format, "{0}") {
		return strings.Replace(format, "{0}", resStr, -1)
	}
	return resStr
}

// Exists checks if a record exists by ID
func (h *MetaDataHelpers) Exists(model interface{}, id int64) bool {
	var count int64
	h.DB.Model(model).Where("Id = ?", id).Count(&count)
	return count > 0
}

// Insert inserts a record with specified Cov values
// C#: Insert([new Cov("Name", "A"), ...])
func (h *MetaDataHelpers) Insert(model interface{}, covs []Cov) error {
	// Create a map from covs
	values := make(map[string]interface{})
	for _, cov := range covs {
		values[cov.Name] = cov.Value
	}
	return h.DB.Model(model).Create(values).Error
}

// Update updates a record with specified Cov values
func (h *MetaDataHelpers) Update(model interface{}, id int64, covs []Cov) error {
	values := make(map[string]interface{})
	for _, cov := range covs {
		values[cov.Name] = cov.Value
	}
	return h.DB.Model(model).Where("Id = ?", id).Updates(values).Error
}

// Helper to bridge Sz84Store with these helpers
func (s *Sz84Store) Meta() *MetaDataHelpers {
	return NewMetaDataHelpers(s.db)
}
