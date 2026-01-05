package b2b

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
)

// B2BService 企业间通信服务接口
type B2BService interface {
	Connect(sourceEntCode, targetEntCode string) error
	HandleHandshake(req HandshakeRequest) (*HandshakeResponse, error)
	SearchLocalKnowledge(query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error)
	SearchMeshKnowledge(query string, limit int, filter *types.SearchFilter) ([]types.DocChunk, error)
	SendCrossEnterpriseMessage(fromEmployeeID, toEmployeeID string, msg string) error
	CallRemoteTool(sourceEntID, targetEntID uint, toolName string, arguments map[string]any) (any, error)
	VerifyIdentity(entCode string, signature string) bool
	VerifyB2BToken(tokenString string) (*models.EnterpriseGORM, error)
	RegisterEndpoint(entID uint, name, endpointType, url string) error
	DiscoverEndpoints(query string) ([]models.MCPServerGORM, error)
	DiscoverMeshEndpoints(query string) ([]models.MCPServerGORM, error)
	CheckDispatchPermission(employeeID, targetOrgID uint, action string) (bool, error)

	// 技能共享
	RequestSkillSharing(fromEntID, targetEntID uint, skillName string) error
	ApproveSkillSharing(sharingID uint, status string) error
	ListSkillSharings(entID uint, role string) ([]models.B2BSkillSharingGORM, error)

	// 员工外派
	DispatchEmployee(employeeID, sourceEntID, targetEntID uint, permissions []string) error
	ApproveDispatch(dispatchID uint, status string) error
	ListDispatchedEmployees(entID uint, role string) ([]models.DigitalEmployeeDispatchGORM, error)
}

type HandshakeRequest struct {
	SourceEntCode string `json:"source_ent_code"`
	Challenge     string `json:"challenge"`
	Signature     string `json:"signature"`
}

type HandshakeResponse struct {
	Success    bool   `json:"success"`
	TargetCode string `json:"target_code"`
	Acceptance string `json:"acceptance"`
	Signature  string `json:"signature"`
}
