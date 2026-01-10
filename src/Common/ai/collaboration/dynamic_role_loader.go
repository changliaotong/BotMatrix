package collaboration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// DynamicRoleLoader 动态角色加载器
// 实现角色动态扩展机制，支持在运行时加载新的角色类型

type DynamicRoleLoader struct {
	mu          sync.RWMutex
	messageBus  MessageBus
	roleFactories map[string]RoleFactory
	loadedRoles  map[string]Role
}

// RoleFactory 角色工厂接口
type RoleFactory interface {
	// 创建角色实例
	CreateRole(config RoleConfig) (Role, error)
	// 获取支持的角色类型
	GetSupportedRoleType() string
}

// RoleConfig 角色配置
type RoleConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Skills      []string          `json:"skills"`
	Properties  map[string]interface{} `json:"properties"`
}

// NewDynamicRoleLoader 创建新的动态角色加载器
func NewDynamicRoleLoader(messageBus MessageBus) *DynamicRoleLoader {
	return &DynamicRoleLoader{
		messageBus:     messageBus,
		roleFactories: make(map[string]RoleFactory),
		loadedRoles:   make(map[string]Role),
	}
}

// RegisterRoleFactory 注册角色工厂
func (dr *DynamicRoleLoader) RegisterRoleFactory(factory RoleFactory) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	roleType := factory.GetSupportedRoleType()
	if _, exists := dr.roleFactories[roleType]; exists {
		return fmt.Errorf("role factory for type %s already registered", roleType)
	}

	dr.roleFactories[roleType] = factory
	log.Printf("Registered role factory for type: %s", roleType)
	return nil
}

// LoadRoleFromConfig 从配置加载角色
func (dr *DynamicRoleLoader) LoadRoleFromConfig(config RoleConfig) (Role, error) {
	dr.mu.RLock()
	factory, exists := dr.roleFactories[config.Type]
	dr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no role factory registered for type: %s", config.Type)
	}

	role, err := factory.CreateRole(config)
	if err != nil {
		return nil, err
	}

	dr.mu.Lock()
	dr.loadedRoles[config.ID] = role
	dr.mu.Unlock()

	log.Printf("Loaded role: %s (%s) with ID: %s", config.Name, config.Type, config.ID)
	return role, nil
}

// LoadRolesFromDirectory 从目录加载所有角色配置
func (dr *DynamicRoleLoader) LoadRolesFromDirectory(dirPath string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := dr.LoadRoleFromFile(file); err != nil {
			log.Printf("Failed to load role from file %s: %v", file, err)
		}
	}

	return nil
}

// LoadRoleFromFile 从文件加载角色配置
func (dr *DynamicRoleLoader) LoadRoleFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config RoleConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return err
	}

	_, err = dr.LoadRoleFromConfig(config)
	return err
}

// UnloadRole 卸载角色
func (dr *DynamicRoleLoader) UnloadRole(roleID string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if _, exists := dr.loadedRoles[roleID]; !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	delete(dr.loadedRoles, roleID)
	log.Printf("Unloaded role with ID: %s", roleID)
	return nil
}

// GetRole 获取角色
func (dr *DynamicRoleLoader) GetRole(roleID string) (Role, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	role, exists := dr.loadedRoles[roleID]
	if !exists {
		return nil, fmt.Errorf("role with ID %s not found", roleID)
	}

	return role, nil
}

// GetAllRoles 获取所有已加载的角色
func (dr *DynamicRoleLoader) GetAllRoles() []Role {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	roles := make([]Role, 0, len(dr.loadedRoles))
	for _, role := range dr.loadedRoles {
		roles = append(roles, role)
	}

	return roles
}

// GetRoleTypes 获取所有支持的角色类型
func (dr *DynamicRoleLoader) GetRoleTypes() []string {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	types := make([]string, 0, len(dr.roleFactories))
	for roleType := range dr.roleFactories {
		types = append(types, roleType)
	}

	return types
}