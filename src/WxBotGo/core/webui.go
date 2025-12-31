package core

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// WebUI 提供配置界面
type WebUI struct {
	config *Config
}

// NewWebUI 创建新的 WebUI 实例
func NewWebUI(config *Config) *WebUI {
	return &WebUI{
		config: config,
	}
}

// Start 启动 WebUI
func (w *WebUI) Start(port string) error {
	// 配置路由
	http.HandleFunc("/", w.handleIndex)
	http.HandleFunc("/config", w.handleConfig)
	http.HandleFunc("/save", w.handleSave)

	fmt.Printf("WebUI 启动在端口 %s\n", port)
	return http.ListenAndServe(":"+port, nil)
}

// handleIndex 处理首页请求
func (w *WebUI) handleIndex(writer http.ResponseWriter, request *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WxBotGo 配置管理</title>
    <style>
        body {
            font-family: 'Microsoft YaHei', sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .section {
            margin-bottom: 20px;
            padding: 15px;
            border: 1px solid #eee;
            border-radius: 4px;
        }
        .section h2 {
            color: #666;
            margin-top: 0;
            font-size: 1.2em;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #333;
        }
        input[type="text"], select {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        input[type="checkbox"] {
            margin-right: 5px;
        }
        .btn {
            background-color: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        .btn:hover {
            background-color: #0056b3;
        }
        .success {
            color: #28a745;
            margin-top: 10px;
        }
        .error {
            color: #dc3545;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>WxBotGo 配置管理</h1>
        
        <form action="/save" method="post">
            <div class="section">
                <h2>网络配置</h2>
                {{if .Config.Networks}}
                {{with index .Config.Networks 0}}
                <div class="form-group">
                    <label for="manager_url">Manager URL:</label>
                    <input type="text" id="manager_url" name="manager_url" value="{{.ManagerURL}}" required>
                </div>
                <div class="form-group">
                    <label for="self_id">Self ID:</label>
                    <input type="text" id="self_id" name="self_id" value="{{.SelfID}}" required>
                </div>
                {{end}}
                {{end}}
            </div>

            <div class="section">
                <h2>HTTP 配置</h2>
                {{if .Config.HTTPs}}
                {{with index .Config.HTTPs 0}}
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="http_enabled" {{if .Enabled}}checked{{end}}>
                        启用 HTTP 服务
                    </label>
                </div>
                <div class="form-group">
                    <label for="http_host">HTTP 主机:</label>
                    <input type="text" id="http_host" name="http_host" value="{{.Host}}">
                </div>
                <div class="form-group">
                    <label for="http_port">HTTP 端口:</label>
                    <input type="text" id="http_port" name="http_port" value="{{.Port}}">
                </div>
                {{end}}
                {{end}}
            </div>

            <div class="section">
                <h2>WebSocket 配置</h2>
                {{if .Config.WebSockets}}
                {{with index .Config.WebSockets 0}}
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="ws_enabled" {{if .Enabled}}checked{{end}}>
                        启用 WebSocket 服务
                    </label>
                </div>
                <div class="form-group">
                    <label for="ws_host">WebSocket 主机:</label>
                    <input type="text" id="ws_host" name="ws_host" value="{{.Host}}">
                </div>
                <div class="form-group">
                    <label for="ws_port">WebSocket 端口:</label>
                    <input type="text" id="ws_port" name="ws_port" value="{{.Port}}">
                </div>
                <div class="form-group">
                    <label for="ws_path">WebSocket 路径:</label>
                    <input type="text" id="ws_path" name="ws_path" value="{{.Path}}">
                </div>
                {{end}}
                {{end}}
            </div>

            <div class="section">
                <h2>日志配置</h2>
                <div class="form-group">
                    <label for="log_level">日志级别:</label>
                    <select id="log_level" name="log_level">
                        <option value="debug" {{if eq .Config.Logging.Level "debug"}}selected{{end}}>Debug</option>
                        <option value="info" {{if eq .Config.Logging.Level "info"}}selected{{end}}>Info</option>
                        <option value="warn" {{if eq .Config.Logging.Level "warn"}}selected{{end}}>Warn</option>
                        <option value="error" {{if eq .Config.Logging.Level "error"}}selected{{end}}>Error</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="log_file">日志文件:</label>
                    <input type="text" id="log_file" name="log_file" value="{{.Config.Logging.File}}">
                </div>
            </div>

            <div class="section">
                <h2>功能配置</h2>
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="auto_login" {{if .Config.Features.AutoLogin}}checked{{end}}>
                        自动登录
                    </label>
                </div>
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="qr_code_save" {{if .Config.Features.QRCodeSave}}checked{{end}}>
                        保存二维码
                    </label>
                </div>
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="auto_reconnect" {{if .Config.Features.AutoReconnect}}checked{{end}}>
                        自动重连
                    </label>
                </div>
                <div class="form-group">
                    <label>
                        <input type="checkbox" name="report_self_msg" {{if .Config.Features.ReportSelfMsg}}checked{{end}}>
                        上报自身消息
                    </label>
                </div>
            </div>

            <button type="submit" class="btn">保存配置</button>
        </form>

        {{if .Success}}
            <div class="success">配置已保存成功！</div>
        {{end}}
        {{if .Error}}
            <div class="error">保存失败: {{.Error}}</div>
        {{end}}
    </div>
</body>
</html>
	`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查是否有成功或错误消息
	data := map[string]any{
		"Config":  w.config,
		"Success": request.URL.Query().Get("success") == "true",
		"Error":   request.URL.Query().Get("error"),
	}

	err = t.Execute(writer, data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// handleConfig 处理配置 JSON 请求
func (w *WebUI) handleConfig(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(w.config)
}

// handleSave 处理配置保存
func (w *WebUI) handleSave(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单数据
	request.ParseForm()

	// 更新配置
	if len(w.config.Networks) > 0 {
		w.config.Networks[0].ManagerURL = request.Form.Get("manager_url")
		w.config.Networks[0].SelfID = request.Form.Get("self_id")
	}
	if len(w.config.HTTPs) > 0 {
		w.config.HTTPs[0].Enabled = request.Form.Get("http_enabled") == "on"
		w.config.HTTPs[0].Host = request.Form.Get("http_host")
		w.config.HTTPs[0].Port = request.Form.Get("http_port")
	}
	if len(w.config.WebSockets) > 0 {
		w.config.WebSockets[0].Enabled = request.Form.Get("ws_enabled") == "on"
		w.config.WebSockets[0].Host = request.Form.Get("ws_host")
		w.config.WebSockets[0].Port = request.Form.Get("ws_port")
		w.config.WebSockets[0].Path = request.Form.Get("ws_path")
	}

	w.config.Logging.Level = request.Form.Get("log_level")
	w.config.Logging.File = request.Form.Get("log_file")
	w.config.Features.AutoLogin = request.Form.Get("auto_login") == "on"
	w.config.Features.QRCodeSave = request.Form.Get("qr_code_save") == "on"
	w.config.Features.AutoReconnect = request.Form.Get("auto_reconnect") == "on"
	w.config.Features.ReportSelfMsg = request.Form.Get("report_self_msg") == "on"

	// 保存配置文件
	err := SaveConfig("config.json", w.config)
	if err != nil {
		// 保存失败，返回错误
		http.Redirect(writer, request, "/?error="+err.Error(), http.StatusSeeOther)
		return
	}

	// 保存成功
	http.Redirect(writer, request, "/?success=true", http.StatusSeeOther)
}
