package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/types"
	"botworker/internal/config"
	"botworker/plugins"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// WorkerAIProvider 为 BotWorker 提供的 AIServiceProvider 实现
type WorkerAIProvider struct {
	config *config.Config
	kb     rag.KnowledgeBase
}

func NewWorkerAIProvider(cfg *config.Config, kb rag.KnowledgeBase) *WorkerAIProvider {
	return &WorkerAIProvider{
		config: cfg,
		kb:     kb,
	}
}

func (p *WorkerAIProvider) SyncSkillCall(ctx context.Context, skillName string, params map[string]any) (any, error) {
	// Worker 端的 SyncSkillCall 暂时通过向上请求 BotNexus 或者在本地查找插件实现
	// 这里简单实现为：如果在本地插件中能找到，就本地执行
	// 实际生产中可能需要通过 Redis 转发给 Nexus，再由 Nexus 分发
	return nil, fmt.Errorf("SyncSkillCall not implemented on Worker yet")
}

func (p *WorkerAIProvider) GetKnowledgeBase() ai.KnowledgeBase {
	return p.kb
}

func (p *WorkerAIProvider) GetWorkers() []types.WorkerInfo {
	// Worker 自身作为唯一的节点返回，这样 SkillManager 就能获取到本地能力
	serverMutex.RLock()
	s := workerServer
	serverMutex.RUnlock()
	if s == nil {
		return nil
	}

	pm := s.GetPluginManager()
	if pm == nil {
		return nil
	}

	var capabilities []types.WorkerCapability

	// 汇总所有插件的能力
	// 1. 外部插件
	for _, versions := range pm.GetPlugins() {
		for _, p := range versions {
			// 外部插件通过 Config 提取技能
			if p.Config != nil {
				// Intents
				for _, intent := range p.Config.Intents {
					capabilities = append(capabilities, types.WorkerCapability{
						Name:        intent.Name,
						Description: p.Config.Description,
						Usage:       fmt.Sprintf("Keywords: %v", intent.Keywords),
						Regex:       intent.Regex,
					})
				}
				// Capabilities
				for _, capName := range p.Config.Capabilities {
					alreadyAdded := false
					for _, c := range capabilities {
						if c.Name == capName {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						capabilities = append(capabilities, types.WorkerCapability{
							Name:        capName,
							Description: fmt.Sprintf("Capability: %s", capName),
							Usage:       fmt.Sprintf("Directly call capability %s", capName),
						})
					}
				}
			}
		}
	}

	// 2. 内部插件
	for _, p := range pm.GetInternalPlugins() {
		if sc, ok := p.(core.SkillCapable); ok {
			for _, skill := range sc.GetSkills() {
				capabilities = append(capabilities, types.WorkerCapability{
					Name:        skill.Name,
					Description: skill.Description,
					Usage:       skill.Usage,
					Params:      skill.Params,
					Regex:       skill.Regex,
				})
			}
		}
	}

	return []types.WorkerInfo{
		{
			ID:           "local_worker",
			Capabilities: capabilities,
		},
	}
}

func (p *WorkerAIProvider) CheckPermission(ctx context.Context, botID string, userID uint, orgID uint, skillName string) (bool, error) {
	// 简单实现，默认允许，或者查询本地数据库
	return true, nil
}

func (p *WorkerAIProvider) GetGORMDB() *gorm.DB {
	return plugins.GlobalGORMDB
}

func (p *WorkerAIProvider) GetManifest() *ai.SystemManifest {
	// 返回一个基础的 Manifest
	return &ai.SystemManifest{
		KnowledgeBase: p.kb,
	}
}

func (p *WorkerAIProvider) IsDigitalEmployeeEnabled() bool {
	return true // 默认开启，或者从配置读取
}
