package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GeneratorResult struct {
	Manifest map[string]interface{} `json:"manifest"`
	Code     string                 `json:"code"`
	Filename string                 `json:"filename"`
	Name     string                 `json:"name"`
}

func GeneratePlugin(prompt string, lang string) (*GeneratorResult, error) {
	apiKey := os.Getenv("BM_AI_KEY")
	baseURL := os.Getenv("BM_AI_URL")
	model := os.Getenv("BM_AI_MODEL")

	if apiKey == "" {
		return nil, fmt.Errorf("BM_AI_KEY environment variable not set")
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	systemPrompt := fmt.Sprintf(`You are a BotMatrix Plugin Generator.
BotMatrix plugins communicate via JSON over stdin/stdout.
The response must be a valid JSON object with:
1. "manifest": Object representing plugin.json.
2. "code": String containing the source code.
3. "filename": String for the source file.
4. "name": String for the plugin directory name (e.g., "my_plugin").

You MUST generate the code in %s.

SECURITY RULES:
- DO NOT use dangerous system calls (e.g., os.system, subprocess, exec, eval).
- DO NOT attempt to read sensitive files or environment variables.
- DO NOT include any hardcoded credentials.
- The plugin should ONLY perform its intended task within the BotMatrix protocol.

If %s is "python":
- Use "asyncio" for the main loop.
- Use "json" and "sys" for I/O.
- Entry point in manifest should be "python <filename>".

If %s is "go":
- Use "encoding/json" and "os".
- Entry point in manifest should be "./main.exe".

Return ONLY the raw JSON object, no markdown.`, lang, lang, lang)

	type aiRequest struct {
		Model    string              `json:"model"`
		Messages []map[string]string `json:"messages"`
	}

	reqBody := aiRequest{
		Model: model,
		Messages: []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequest("POST", strings.TrimSuffix(baseURL, "/")+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("error parsing AI response: %v, body: %s", err, string(body))
	}

	if aiResp.Error.Message != "" {
		return nil, fmt.Errorf("AI Error: %s", aiResp.Error.Message)
	}

	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("AI returned no choices")
	}

	content := aiResp.Choices[0].Message.Content
	// Remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result GeneratorResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("error parsing AI content JSON: %v, content: %s", err, content)
	}

	// Security Audit
	if err := SecurityAudit(&result, lang); err != nil {
		return nil, fmt.Errorf("security audit failed: %v", err)
	}

	return &result, nil
}

func SecurityAudit(result *GeneratorResult, lang string) error {
	code := result.Code

	dangerousKeywords := []string{}
	if lang == "python" {
		dangerousKeywords = []string{
			"os.system", "subprocess", "eval(", "exec(", "open(", "__import__",
			"pickle", "shutil", "telnetlib", "ftplib",
		}
	} else if lang == "go" {
		dangerousKeywords = []string{
			"os/exec", "syscall", "unsafe", "reflect",
		}
	}

	for _, kw := range dangerousKeywords {
		if strings.Contains(code, kw) {
			return fmt.Errorf("detected dangerous keyword: %s", kw)
		}
	}

	// Check manifest permissions
	if perms, ok := result.Manifest["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if p == "*" {
				return fmt.Errorf("wildcard permission '*' is not allowed for auto-generated plugins")
			}
		}
	}

	return nil
}

func SavePlugin(result *GeneratorResult, targetDir string) (string, error) {
	dir := filepath.Join(targetDir, result.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	manifestData, _ := json.MarshalIndent(result.Manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), manifestData, 0644); err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(dir, result.Filename), []byte(result.Code), 0644); err != nil {
		return "", err
	}

	return dir, nil
}
