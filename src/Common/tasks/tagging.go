package tasks

import (
	"BotMatrix/common/models"
	"strings"

	"gorm.io/gorm"
)

// TaggingManager 处理标签逻辑
type TaggingManager struct {
	db *gorm.DB
}

func NewTaggingManager(db *gorm.DB) *TaggingManager {
	return &TaggingManager{db: db}
}

// AddTag 为目标添加标签
func (tm *TaggingManager) AddTag(targetType, targetID, tagName string) error {
	tag := models.Tag{
		Name:     tagName,
		Type:     targetType,
		TargetID: targetID,
	}
	return tm.db.Where(models.Tag{Name: tagName, Type: targetType, TargetID: targetID}).FirstOrCreate(&tag).Error
}

// RemoveTag 为目标移除标签
func (tm *TaggingManager) RemoveTag(targetType, targetID, tagName string) error {
	return tm.db.Where("type = ? AND target_id = ? AND name = ?", targetType, targetID, tagName).Delete(&models.Tag{}).Error
}

// GetTargetsByTags 根据标签组合查找目标
// logic: AND, OR
func (tm *TaggingManager) GetTargetsByTags(targetType string, tags []string, logic string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	var results []string
	if strings.ToUpper(logic) == "OR" {
		err := tm.db.Model(&models.Tag{}).
			Where("type = ? AND name IN ?", targetType, tags).
			Pluck("DISTINCT target_id", &results).Error
		return results, err
	}

	// AND 逻辑：必须包含所有指定的标签
	err := tm.db.Model(&models.Tag{}).
		Select("target_id").
		Where("type = ? AND name IN ?", targetType, tags).
		Group("target_id").
		Having("COUNT(DISTINCT name) = ?", len(tags)).
		Pluck("target_id", &results).Error

	return results, err
}

// GetTagsByTarget 获取目标的标签
func (tm *TaggingManager) GetTagsByTarget(targetType, targetID string) ([]string, error) {
	var tags []string
	err := tm.db.Model(&models.Tag{}).
		Where("type = ? AND target_id = ?", targetType, targetID).
		Pluck("name", &tags).Error
	return tags, err
}
