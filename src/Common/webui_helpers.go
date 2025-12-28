package common

import (
	"html/template"
	"net/http"
)

// ConnectionWebUI 连接配置 WebUI 帮助函数
type ConnectionWebUI struct {
	manager *ConnectionManager
}

// NewConnectionWebUI 创建新的连接 WebUI 管理器
func NewConnectionWebUI(manager *ConnectionManager) *ConnectionWebUI {
	return &ConnectionWebUI{
		manager: manager,
	}
}

// RenderConfigPage 渲染连接配置页面
func (w *ConnectionWebUI) RenderConfigPage(writer http.ResponseWriter, request *http.Request) {
	connections := w.manager.GetConnections()
	data := map[string]any{
		"Connections": connections,
		"ClientTypes": []string{"wx", "qq", "wecom", "telegram"},
	}

	tmpl := template.Must(template.ParseFiles("templates/connections.html"))
	err := tmpl.Execute(writer, data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// HandleSaveConnections 处理连接配置保存
func (w *ConnectionWebUI) HandleSaveConnections(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := request.ParseForm()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// 解析表单数据并更新连接配置
	// 这里需要根据实际表单结构实现
	// ...

	http.Redirect(writer, request, "/connections", http.StatusSeeOther)
}
