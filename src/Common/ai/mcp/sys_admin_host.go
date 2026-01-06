package mcp

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SysAdminMCPHost provides system administration capabilities for the "Factory Manager"
type SysAdminMCPHost struct {
	db *gorm.DB
}

func NewSysAdminMCPHost(db *gorm.DB) *SysAdminMCPHost {
	return &SysAdminMCPHost{db: db}
}

// Interface Implementation
func (h *SysAdminMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "create_role_template",
			Description: "Create a new Digital Role Template (SOP)",
			InputSchema: jsonSchema(`{
				"type": "object",
				"properties": {
					"name": { "type": "string", "description": "Name of the role (e.g., 'QA Engineer')" },
					"description": { "type": "string", "description": "Description of the role" },
					"base_prompt": { "type": "string", "description": "The system prompt/SOP for this role" },
					"default_skills": { "type": "string", "description": "Comma-separated list of skills (e.g., 'local_dev,git_ops')" }
				},
				"required": ["name", "base_prompt"]
			}`),
		},
		{
			Name:        "create_digital_employee",
			Description: "Manufacture (recruit) a new Digital Employee based on a template",
			InputSchema: jsonSchema(`{
				"type": "object",
				"properties": {
					"role_template_id": { "type": "integer", "description": "ID of the role template" },
					"name_prefix": { "type": "string", "description": "Optional prefix for the bot name" },
					"org_id": { "type": "integer", "description": "Organization ID (default: 1)" }
				},
				"required": ["role_template_id"]
			}`),
		},
		{
			Name:        "list_role_templates",
			Description: "List available role templates",
			InputSchema: jsonSchema(`{
				"type": "object",
				"properties": {}
			}`),
		},
	}, nil
}

func (h *SysAdminMCPHost) CallTool(ctx context.Context, serverID string, toolName string, args map[string]any) (any, error) {
	switch toolName {
	case "create_role_template":
		return h.createRoleTemplate(args)
	case "create_digital_employee":
		return h.createDigitalEmployee(args)
	case "list_role_templates":
		return h.listRoleTemplates()
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (h *SysAdminMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *SysAdminMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *SysAdminMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *SysAdminMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// Helper methods

func (h *SysAdminMCPHost) createRoleTemplate(args map[string]any) (any, error) {
	name, _ := args["name"].(string)
	basePrompt, _ := args["base_prompt"].(string)
	description, _ := args["description"].(string)
	defaultSkills, _ := args["default_skills"].(string)

	template := models.DigitalRoleTemplate{
		Name:          name,
		Description:   description,
		BasePrompt:    basePrompt,
		DefaultSkills: defaultSkills,
	}

	if err := h.db.Create(&template).Error; err != nil {
		return nil, fmt.Errorf("failed to create template: %v", err)
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: fmt.Sprintf("Role Template '%s' created with ID: %d", name, template.ID)},
		},
	}, nil
}

func (h *SysAdminMCPHost) createDigitalEmployee(args map[string]any) (any, error) {
	roleTemplateIDFloat, ok := args["role_template_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("role_template_id is required")
	}
	roleTemplateID := uint(roleTemplateIDFloat)

	orgIDFloat, _ := args["org_id"].(float64)
	orgID := uint(orgIDFloat)
	if orgID == 0 {
		orgID = 1
	}

	namePrefix, _ := args["name_prefix"].(string)

	// Fetch template to get default name or logic
	var template models.DigitalRoleTemplate
	if err := h.db.First(&template, roleTemplateID).Error; err != nil {
		return nil, fmt.Errorf("template not found: %v", err)
	}

	empName := template.Name
	if namePrefix != "" {
		empName = namePrefix + " - " + empName
	}

	emp := models.DigitalEmployee{
		EnterpriseID:   orgID,
		RoleTemplateID: roleTemplateID,
		Name:           empName,
		BotID:          fmt.Sprintf("bot_%d_%d", roleTemplateID, time.Now().UnixNano()),
		EmployeeID:     fmt.Sprintf("EMP%d%d", orgID, time.Now().UnixNano()%100000),
		Status:         "active",
		Bio:            template.DefaultBio,
		Skills:         template.DefaultSkills,
		OnboardingAt:   time.Now(),
	}

	if err := h.db.Create(&emp).Error; err != nil {
		return nil, fmt.Errorf("failed to recruit employee: %v", err)
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: fmt.Sprintf("Manufactured Digital Employee '%s' (BotID: %s, EmployeeID: %s)", emp.Name, emp.BotID, emp.EmployeeID)},
		},
	}, nil
}

func (h *SysAdminMCPHost) listRoleTemplates() (any, error) {
	var templates []models.DigitalRoleTemplate
	if err := h.db.Find(&templates).Error; err != nil {
		return nil, err
	}

	var result string
	for _, t := range templates {
		result += fmt.Sprintf("ID: %d | Name: %s | Skills: %s\n", t.ID, t.Name, t.DefaultSkills)
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{Type: "text", Text: result},
		},
	}, nil
}

func jsonSchema(s string) map[string]any {
	var m map[string]any
	json.Unmarshal([]byte(s), &m)
	return m
}
