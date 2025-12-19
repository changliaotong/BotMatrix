package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
)

// Config holds the configuration
type Config struct {
	AppID      uint64 `json:"app_id"` // AppID is uint64 in SDK
	Token      string `json:"token"`
	Secret     string `json:"secret"`
	Sandbox    bool   `json:"sandbox"`
	SelfID     string `json:"self_id"` // Optional: manually set SelfID
	NexusAddr  string `json:"nexus_addr"`
	LogPort    int    `json:"log_port"`    // Port for HTTP Log Viewer
	FileHost   string `json:"file_host"`   // Public base URL for serving files (e.g. http://1.2.3.4:8080)
	MediaRoute string `json:"media_route"` // Internal route path for media (default: /media/)
}

var (
	config         Config
	nexusConn      *websocket.Conn
	nexusMu        sync.Mutex
	api            openapi.OpenAPI
	ctx            context.Context
	selfID         string
	logManager     *LogManager
	msgSeq         int64
	accessToken    string
	tokenExpiresAt int64
	tokenMu        sync.Mutex
)

// getAppAccessToken fetches or returns a valid access token
func getAppAccessToken() (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	// Return cached token if valid (buffer 60s)
	if accessToken != "" && time.Now().Unix() < tokenExpiresAt-60 {
		return accessToken, nil
	}

	url := "https://bots.qq.com/app/getAppAccessToken"
	data := map[string]string{
		"appId":        fmt.Sprintf("%d", config.AppID),
		"clientSecret": config.Secret,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	accessToken = result.AccessToken
	exp, _ := strconv.Atoi(result.ExpiresIn)
	tokenExpiresAt = time.Now().Unix() + int64(exp)

	log.Printf("Refreshed Access Token, expires in %d seconds", exp)
	return accessToken, nil
}

// LogManager handles in-memory log buffering
type LogManager struct {
	buffer []string
	size   int
	mu     sync.RWMutex
}

func NewLogManager(size int) *LogManager {
	return &LogManager{
		buffer: make([]string, 0, size),
		size:   size,
	}
}

func (l *LogManager) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := string(p)
	// Simple rotation
	if len(l.buffer) >= l.size {
		l.buffer = l.buffer[1:]
	}
	l.buffer = append(l.buffer, strings.TrimRight(msg, "\n"))

	// Stream to Nexus
	go func(m string) {
		sendToNexus(map[string]interface{}{
			"post_type": "log",
			"level":     "INFO",
			"message":   strings.TrimSpace(m),
			"time":      time.Now().Format("15:04:05"),
			"self_id":   selfID,
		})
	}(msg)

	return os.Stdout.Write(p)
}

func (l *LogManager) GetLogs(lines int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if lines <= 0 || lines >= len(l.buffer) {
		// Return a copy
		result := make([]string, len(l.buffer))
		copy(result, l.buffer)
		return result
	}

	return l.buffer[len(l.buffer)-lines:]
}

// SessionCache to store last message ID for replying
type SessionCache struct {
	sync.RWMutex     `json:"-"`
	UserLastMsgID    map[string]string                   `json:"user_last_msg_id"`
	GroupLastMsgID   map[string]string                   `json:"group_last_msg_id"`
	ChannelLastMsgID map[string]string                   `json:"channel_last_msg_id"`
	LastMsgTime      map[string]int64                    `json:"last_msg_time"`
	PendingActions   map[string][]map[string]interface{} `json:"pending_actions"`
}

var sessionCache = &SessionCache{
	UserLastMsgID:    make(map[string]string),
	GroupLastMsgID:   make(map[string]string),
	ChannelLastMsgID: make(map[string]string),
	LastMsgTime:      make(map[string]int64),
	PendingActions:   make(map[string][]map[string]interface{}),
}

func (s *SessionCache) SaveDisk() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal session cache: %v", err)
		return
	}
	if err := ioutil.WriteFile("session_cache.json", data, 0644); err != nil {
		log.Printf("Failed to save session cache to disk: %v", err)
	}
}

func (s *SessionCache) LoadDisk() {
	data, err := ioutil.ReadFile("session_cache.json")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read session cache from disk: %v", err)
		}
		return
	}
	s.Lock()
	defer s.Unlock()
	if err := json.Unmarshal(data, s); err != nil {
		log.Printf("Failed to unmarshal session cache: %v", err)
	} else {
		log.Printf("Loaded session cache from disk: %d users, %d groups", len(s.UserLastMsgID), len(s.GroupLastMsgID))
	}
	// Initialize maps if nil
	if s.UserLastMsgID == nil {
		s.UserLastMsgID = make(map[string]string)
	}
	if s.GroupLastMsgID == nil {
		s.GroupLastMsgID = make(map[string]string)
	}
	if s.ChannelLastMsgID == nil {
		s.ChannelLastMsgID = make(map[string]string)
	}
	if s.LastMsgTime == nil {
		s.LastMsgTime = make(map[string]int64)
	}
	if s.PendingActions == nil {
		s.PendingActions = make(map[string][]map[string]interface{})
	}
}

func (s *SessionCache) AddPending(keyType, key string, action map[string]interface{}) {
	s.Lock()
	defer s.Unlock()
	compositeKey := keyType + ":" + key
	if s.PendingActions == nil {
		s.PendingActions = make(map[string][]map[string]interface{})
	}
	s.PendingActions[compositeKey] = append(s.PendingActions[compositeKey], action)
	log.Printf("[SessionCache] Queued pending action for %s %s", keyType, key)
	go s.SaveDisk()
}

func (s *SessionCache) Save(keyType, key, msgID string) []map[string]interface{} {
	s.Lock()
	defer s.Unlock()
	switch keyType {
	case "user":
		s.UserLastMsgID[key] = msgID
	case "group":
		s.GroupLastMsgID[key] = msgID
	case "channel":
		s.ChannelLastMsgID[key] = msgID
	}
	s.LastMsgTime[msgID] = time.Now().Unix()
	log.Printf("[SessionCache] Saved %s session for %s: %s", keyType, key, msgID)

	// Check pending
	compositeKey := keyType + ":" + key
	var pending []map[string]interface{}
	if actions, ok := s.PendingActions[compositeKey]; ok && len(actions) > 0 {
		pending = actions
		delete(s.PendingActions, compositeKey)
		log.Printf("[SessionCache] Found %d pending actions for %s %s", len(pending), keyType, key)
	}

	// Save to disk asynchronously
	go s.SaveDisk()

	return pending
}

func (s *SessionCache) Get(keyType, key string) string {
	s.RLock()
	defer s.RUnlock()

	var msgID string
	switch keyType {
	case "user":
		msgID = s.UserLastMsgID[key]
	case "group":
		msgID = s.GroupLastMsgID[key]
	case "channel":
		msgID = s.ChannelLastMsgID[key]
	}

	if msgID == "" {
		return ""
	}

	// Check 5-minute limit (300 seconds)
	// We use a slightly shorter limit (290s) to be safe
	if ts, ok := s.LastMsgTime[msgID]; ok {
		if time.Now().Unix()-ts > 290 {
			log.Printf("[SessionCache] Session expired for %s %s (MsgID: %s)", keyType, key, msgID)
			return "" // Expired
		}
		return msgID
	}
	return ""
}

func loadConfig() {
	// Try to load from file first
	file, err := os.Open("config.json")
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			log.Println("Error decoding config.json, falling back to environment variables:", err)
		}
	} else {
		log.Println("config.json not found, using environment variables.")
	}

	// Override with environment variables if present
	if envAppID := os.Getenv("TENCENT_APP_ID"); envAppID != "" {
		fmt.Sscanf(envAppID, "%d", &config.AppID)
	}
	if envToken := os.Getenv("TENCENT_TOKEN"); envToken != "" {
		config.Token = envToken
	}
	if envSecret := os.Getenv("TENCENT_SECRET"); envSecret != "" {
		config.Secret = envSecret
	}
	if envSandbox := os.Getenv("TENCENT_SANDBOX"); envSandbox != "" {
		config.Sandbox = (envSandbox == "true" || envSandbox == "1")
	}
	if envSelfID := os.Getenv("TENCENT_SELF_ID"); envSelfID != "" {
		config.SelfID = envSelfID
	}
	if envNexusAddr := os.Getenv("NEXUS_ADDR"); envNexusAddr != "" {
		config.NexusAddr = envNexusAddr
	}

	if envLogPort := os.Getenv("LOG_PORT"); envLogPort != "" {
		fmt.Sscanf(envLogPort, "%d", &config.LogPort)
	}
	if envFileHost := os.Getenv("FILE_HOST"); envFileHost != "" {
		config.FileHost = envFileHost
	}
	if envMediaRoute := os.Getenv("MEDIA_ROUTE"); envMediaRoute != "" {
		config.MediaRoute = envMediaRoute
	}

	// Defaults
	if config.MediaRoute == "" {
		config.MediaRoute = "/media/"
	}
	// Ensure MediaRoute starts and ends with /
	if !strings.HasPrefix(config.MediaRoute, "/") {
		config.MediaRoute = "/" + config.MediaRoute
	}
	if !strings.HasSuffix(config.MediaRoute, "/") {
		config.MediaRoute = config.MediaRoute + "/"
	}

	// Validation
	if config.AppID == 0 || config.Token == "" || config.Secret == "" {
		log.Fatal("Missing configuration. Please check config.json or environment variables (TENCENT_APP_ID, TENCENT_TOKEN, TENCENT_SECRET).")
	}
	if config.NexusAddr == "" {
		config.NexusAddr = "ws://192.168.0.167:3005/ws/bots"
	}
}

// NexusConnect connects to BotNexus
func NexusConnect() {
	headers := http.Header{}
	// Wait for selfID to be populated
	for selfID == "" {
		time.Sleep(100 * time.Millisecond)
	}
	headers.Add("X-Self-ID", selfID)
	headers.Add("X-Platform", "Guild")

	for {
		log.Printf("Connecting to BotNexus at %s...", config.NexusAddr)
		conn, _, err := websocket.DefaultDialer.Dial(config.NexusAddr, headers)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		nexusConn = conn
		log.Println("Connected to BotNexus!")

		// Handle incoming messages from BotNexus (Actions)
		go handleNexusMessages()

		return
	}
}

func handleNexusMessages() {
	defer nexusConn.Close()
	for {
		_, message, err := nexusConn.ReadMessage()
		if err != nil {
			log.Println("BotNexus connection lost:", err)
			// Reconnect logic could be here, but for now we just exit/restart
			os.Exit(1)
			return
		}

		var actionMap map[string]interface{}
		if err := json.Unmarshal(message, &actionMap); err != nil {
			log.Println("Error parsing action:", err)
			continue
		}

		// Handle Actions (e.g. send_msg)
		// This is where we translate OneBot actions to Tencent SDK calls
		handleAction(actionMap)
	}
}

func uploadGroupFile(groupID string, filePath string, fileType int) (string, error) {
	// Helper to parse response
	parseResponse := func(resp *http.Response) (string, error) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("upload failed: %s - %s", resp.Status, string(body))
		}
		log.Printf("[DEBUG] Upload Response: %s", string(body))
		var result struct {
			FileUUID string `json:"file_uuid"`
			FileInfo string `json:"file_info"`
			TTL      int    `json:"ttl"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}
		return result.FileInfo, nil
	}

	// 1. If URL, Use JSON Payload (Direct URL Upload)
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		log.Printf("[DEBUG] Using Direct URL Upload for Group: %s", filePath)
		payload := map[string]interface{}{
			"file_type":    fileType,
			"url":          filePath,
			"srv_send_msg": false,
		}
		jsonBody, err := json.Marshal(payload)
		if err == nil {
			url := fmt.Sprintf("https://api.sgroup.qq.com/v2/groups/%s/files", groupID)
			if config.Sandbox {
				url = fmt.Sprintf("https://sandbox.api.sgroup.qq.com/v2/groups/%s/files", groupID)
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
			if err == nil {
				token, err := getAppAccessToken()
				if err == nil {
					req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", token))
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{Timeout: 30 * time.Second}
					resp, err := client.Do(req)
					if err == nil {
						fileInfo, err := parseResponse(resp)
						resp.Body.Close()
						if err == nil {
							return fileInfo, nil
						}
						log.Printf("[WARN] Direct URL upload failed (API error): %v. Falling back to Multipart Upload.", err)
					} else {
						log.Printf("[WARN] Direct URL upload request failed: %v. Falling back to Multipart Upload.", err)
					}
				} else {
					log.Printf("[WARN] Failed to get token for Direct URL upload: %v", err)
				}
			} else {
				log.Printf("[WARN] Failed to create request for Direct URL upload: %v", err)
			}
		}
	}

	// 2. Multipart Upload (Local File or Downloaded URL fallback)
	localPath := filePath
	var cleanUp func()

	// If it's a URL (fallback case), we need to download it to a temp file
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		log.Printf("[DEBUG] Downloading file for Multipart Upload: %s", filePath)
		resp, err := http.Get(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to download file for fallback upload: %v", err)
		}
		defer resp.Body.Close()

		tmpFile, err := ioutil.TempFile("", "upload_fallback_*.jpg")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %v", err)
		}

		_, err = io.Copy(tmpFile, resp.Body)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to save downloaded file: %v", err)
		}

		localPath = tmpFile.Name()
		cleanUp = func() {
			os.Remove(localPath)
		}
	}

	if cleanUp != nil {
		defer cleanUp()
	}

	// 2. Prepare Multipart Request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// File Field
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(localPath)))
	h.Set("Content-Type", "image/jpeg") // Force JPEG for now as we save as .jpg
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	// Other Fields
	writer.WriteField("file_type", fmt.Sprintf("%d", fileType))
	writer.WriteField("srv_send_msg", "false")

	err = writer.Close()
	if err != nil {
		return "", err
	}

	// 3. Send Request
	url := fmt.Sprintf("https://api.sgroup.qq.com/v2/groups/%s/files", groupID)
	if config.Sandbox {
		url = fmt.Sprintf("https://sandbox.api.sgroup.qq.com/v2/groups/%s/files", groupID)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	token, err := getAppAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token for upload: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return parseResponse(resp)
}

func uploadC2CFile(userID string, filePath string, fileType int) (string, error) {
	// Helper to parse response
	parseResponse := func(resp *http.Response) (string, error) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("upload failed: %s - %s", resp.Status, string(body))
		}
		log.Printf("[DEBUG] C2C Upload Response: %s", string(body))
		var result struct {
			FileUUID string `json:"file_uuid"`
			FileInfo string `json:"file_info"`
			TTL      int    `json:"ttl"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}
		return result.FileInfo, nil
	}

	// 1. If URL, Use JSON Payload (Direct URL Upload)
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		log.Printf("[DEBUG] Using Direct URL Upload for C2C: %s", filePath)
		payload := map[string]interface{}{
			"file_type":    fileType,
			"url":          filePath,
			"srv_send_msg": false,
		}
		jsonBody, err := json.Marshal(payload)
		if err == nil {
			url := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/files", userID)
			if config.Sandbox {
				url = fmt.Sprintf("https://sandbox.api.sgroup.qq.com/v2/users/%s/files", userID)
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
			if err == nil {
				token, err := getAppAccessToken()
				if err == nil {
					req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", token))
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{Timeout: 30 * time.Second}
					resp, err := client.Do(req)
					if err == nil {
						fileInfo, err := parseResponse(resp)
						resp.Body.Close()
						if err == nil {
							return fileInfo, nil
						}
						log.Printf("[WARN] Direct URL upload failed (API error): %v. Falling back to Multipart Upload.", err)
					} else {
						log.Printf("[WARN] Direct URL upload request failed: %v. Falling back to Multipart Upload.", err)
					}
				} else {
					log.Printf("[WARN] Failed to get token for Direct URL upload: %v", err)
				}
			} else {
				log.Printf("[WARN] Failed to create request for Direct URL upload: %v", err)
			}
		}
	}

	// 2. Multipart Upload (Local File or Downloaded URL fallback)
	localPath := filePath
	var cleanUp func()

	// If it's a URL (fallback case), we need to download it to a temp file
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		log.Printf("[DEBUG] Downloading file for Multipart Upload: %s", filePath)
		resp, err := http.Get(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to download file for fallback upload: %v", err)
		}
		defer resp.Body.Close()

		tmpFile, err := ioutil.TempFile("", "upload_fallback_*.jpg")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %v", err)
		}

		_, err = io.Copy(tmpFile, resp.Body)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to save downloaded file: %v", err)
		}

		localPath = tmpFile.Name()
		cleanUp = func() {
			os.Remove(localPath)
		}
	}

	if cleanUp != nil {
		defer cleanUp()
	}

	// 2. Prepare Multipart Request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// File Field
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(localPath)))
	h.Set("Content-Type", "image/jpeg")
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	// Other Fields
	writer.WriteField("file_type", fmt.Sprintf("%d", fileType))
	writer.WriteField("srv_send_msg", "false")

	err = writer.Close()
	if err != nil {
		return "", err
	}

	// 3. Send Request
	url := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/files", userID)
	if config.Sandbox {
		url = fmt.Sprintf("https://sandbox.api.sgroup.qq.com/v2/users/%s/files", userID)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	token, err := getAppAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get access token for upload: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("QQBot %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return parseResponse(resp)
}

func cleanContent(content string) (string, string, int) {
	// Return: cleanedContent, filePath, fileType (1:Image, 2:Video, 3:Audio, 4:File)

	// Helper to process base64 or path
	process := func(fullMatch, fileVal string, fType int) (string, string, int) {
		if strings.HasPrefix(fileVal, "base64://") {
			b64Data := strings.TrimPrefix(fileVal, "base64://")
			data, err := base64.StdEncoding.DecodeString(b64Data)
			if err != nil {
				log.Printf("Error decoding base64: %v", err)
				return strings.ReplaceAll(content, fullMatch, "[Media Error]"), "", 0
			}

			prefix := "tencent_media_"
			ext := ".dat"
			if fType == 1 {
				ext = ".png"
			}
			if fType == 2 {
				ext = ".mp4"
			}
			if fType == 3 {
				ext = ".amr"
			}

			tmpFile, err := ioutil.TempFile("", prefix+"*"+ext)
			if err != nil {
				log.Printf("Error creating temp file: %v", err)
				return strings.ReplaceAll(content, fullMatch, "[Media Save Error]"), "", 0
			}
			defer tmpFile.Close()

			if _, err := tmpFile.Write(data); err != nil {
				log.Printf("Error writing to temp file: %v", err)
				return strings.ReplaceAll(content, fullMatch, "[Media Write Error]"), "", 0
			}
			cleanMsg := strings.ReplaceAll(content, fullMatch, "")
			return strings.TrimSpace(cleanMsg), tmpFile.Name(), fType
		}
		// Local file
		cleanMsg := strings.ReplaceAll(content, fullMatch, "")
		return strings.TrimSpace(cleanMsg), fileVal, fType
	}

	// Image
	reImg := regexp.MustCompile(`\[CQ:image,[^\]]*\]`)
	if match := reImg.FindString(content); match != "" {
		reFile := regexp.MustCompile(`file=([^,\]]+)`)
		if fileMatches := reFile.FindStringSubmatch(match); len(fileMatches) > 1 {
			return process(match, fileMatches[1], 1)
		}
	}

	// Video
	reVid := regexp.MustCompile(`\[CQ:video,[^\]]*\]`)
	if match := reVid.FindString(content); match != "" {
		reFile := regexp.MustCompile(`file=([^,\]]+)`)
		if fileMatches := reFile.FindStringSubmatch(match); len(fileMatches) > 1 {
			return process(match, fileMatches[1], 2)
		}
	}

	// Audio
	reAud := regexp.MustCompile(`\[CQ:record,[^\]]*\]`)
	if match := reAud.FindString(content); match != "" {
		reFile := regexp.MustCompile(`file=([^,\]]+)`)
		if fileMatches := reFile.FindStringSubmatch(match); len(fileMatches) > 1 {
			return process(match, fileMatches[1], 3)
		}
	}

	return content, "", 0
}

func handleAction(action map[string]interface{}) {
	act, ok := action["action"].(string)
	if !ok {
		return
	}

	params, _ := action["params"].(map[string]interface{})
	log.Printf("[NEXUS-MSG] Received action: %s | Params: %+v", act, params)

	switch act {
	case "send_msg":
		// Generic send_msg
		params, _ := action["params"].(map[string]interface{})
		messageType, _ := params["message_type"].(string)
		content, _ := params["message"].(string)

		if messageType == "private" {
			// C2C
			userID := getString(params, "user_id")
			safeContent, imagePath, fileType := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("user", userID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for User %s", msgID, userID)
				} else {
					log.Printf("[NEXUS-MSG] No active session (MsgID) for User %s. Caching action pending user reply.", userID)
					sessionCache.AddPending("user", userID, action)
					return
				}
			}

			log.Printf("[NEXUS-MSG] Sending Private Message to %s: %s (Img: %s)", userID, safeContent, imagePath)

			var media *dto.MediaInfo
			var publicURL string

			if imagePath != "" {
				// Calculate Public URL for fallback
				if config.FileHost != "" {
					fileName := filepath.Base(imagePath)
					host := strings.TrimRight(config.FileHost, "/")
					route := strings.Trim(config.MediaRoute, "/")
					if route != "" {
						route = "/" + route
					}
					publicURL = fmt.Sprintf("%s%s/%s", host, route, fileName)
				}

				fileInfo, errUpload := uploadC2CFile(userID, imagePath, fileType)
				if errUpload != nil {
					log.Printf("[NEXUS-MSG] Failed to upload C2C file: %v", errUpload)
					// Fallback to URL
					if publicURL != "" {
						// Remove scheme to avoid 40054010
						safeUrl := strings.Replace(publicURL, "http://", "", 1)
						safeUrl = strings.Replace(safeUrl, "https://", "", 1)
						safeContent += fmt.Sprintf("\n[图片]: %s", safeUrl)
					} else {
						safeContent += "\n[Image Upload Failed]"
					}
				} else {
					media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
				}

				// Do not remove immediately if using URL fallback or successful upload (needed for serving)
				if !strings.HasPrefix(imagePath, "http") {
					go func(path string) {
						time.Sleep(10 * time.Minute)
						os.Remove(path)
						log.Printf("Cleaned up temp file: %s", path)
					}(imagePath)
				}
			}

			if safeContent == "" {
				safeContent = " "
			}

			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
				Media:   media,
			}
			if media != nil {
				msgData.MsgType = 7 // 7: Media
			}

			_, err := api.PostC2CMessage(ctx, userID, msgData)
			if err != nil {
				log.Printf("[NEXUS-MSG] Failed to send private message: %v", err)
				// If Media failed (e.g. 40034004), retry with Text Only URL
				if media != nil && publicURL != "" {
					log.Printf("[NEXUS-MSG] Retrying with Text URL fallback...")
					msgData.Media = nil
					msgData.MsgType = 0
					msgData.Content = safeContent + fmt.Sprintf("\n[图片]: %s", publicURL)
					_, errRetry := api.PostC2CMessage(ctx, userID, msgData)
					if errRetry != nil {
						log.Printf("[NEXUS-MSG] Retry failed: %v", errRetry)
					} else {
						log.Printf("[NEXUS-MSG] Retry with URL successful")
						err = nil // Treat as success
					}
				}
			} else {
				log.Printf("[NEXUS-MSG] Private message sent successfully")
			}
			handleSendResponse(err, nil, action)
		} else if messageType == "group" {
			// QQ Group
			groupID := getString(params, "group_id")
			safeContent, imagePath, fileType := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("group", groupID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Group %s", msgID, groupID)
				} else {
					log.Printf("[NEXUS-MSG] No active session (MsgID) for Group %s. Caching action pending group reply.", groupID)
					sessionCache.AddPending("group", groupID, action)
					return
				}
			}

			log.Printf("[NEXUS-MSG] Sending Group Message to %s: %s (Img: %s)", groupID, safeContent, imagePath)

			var err error
			var media *dto.MediaInfo
			var publicURL string

			// Upload file if exists
			if imagePath != "" {
				// Calculate Public URL for fallback
				if config.FileHost != "" {
					fileName := filepath.Base(imagePath)
					host := strings.TrimRight(config.FileHost, "/")
					route := strings.Trim(config.MediaRoute, "/")
					if route != "" {
						route = "/" + route
					}
					publicURL = fmt.Sprintf("%s%s/%s", host, route, fileName)
				}

				fileInfo, errUpload := uploadGroupFile(groupID, imagePath, fileType)
				if errUpload != nil {
					log.Printf("[NEXUS-MSG] Failed to upload group file: %v", errUpload)
					err = errUpload
					if publicURL != "" {
						safeUrl := strings.Replace(publicURL, "http://", "", 1)
						safeUrl = strings.Replace(safeUrl, "https://", "", 1)
						safeContent += fmt.Sprintf("\n[图片]: %s", safeUrl)
					} else {
						safeContent += "\n[Image Upload Failed]"
					}
				} else {
					log.Printf("[NEXUS-MSG] Group file uploaded successfully")
					media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
				}
				if !strings.HasPrefix(imagePath, "http") {
					os.Remove(imagePath)
				}
			}

			// Send message if content is not empty or media is present
			if safeContent != "" || media != nil {
				if safeContent == "" {
					safeContent = " "
				}
				msgData := &dto.MessageToCreate{
					Content: safeContent,
					MsgID:   msgID,
					MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
					Media:   media,
				}
				if media != nil {
					msgData.MsgType = 7
				}
				_, errPost := api.PostGroupMessage(ctx, groupID, msgData)
				if errPost != nil {
					log.Printf("[NEXUS-MSG] Failed to send group message: %v", errPost)
					if err == nil {
						err = errPost
					}
					// Retry with URL fallback
					if media != nil && publicURL != "" {
						log.Printf("[NEXUS-MSG] Retrying Group Msg with Text URL fallback...")
						msgData.Media = nil
						msgData.MsgType = 0
						// Remove scheme
						safeUrl := strings.Replace(publicURL, "http://", "", 1)
						safeUrl = strings.Replace(safeUrl, "https://", "", 1)
						msgData.Content = safeContent + fmt.Sprintf("\n[图片]: %s", safeUrl)
						_, errRetry := api.PostGroupMessage(ctx, groupID, msgData)
						if errRetry != nil {
							log.Printf("[NEXUS-MSG] Retry failed: %v", errRetry)
						} else {
							log.Printf("[NEXUS-MSG] Retry with URL successful")
							err = nil
						}
					}
				} else {
					log.Printf("[NEXUS-MSG] Group message sent successfully")
				}
			}

			handleSendResponse(err, nil, action)
		} else if messageType == "guild" {
			// Guild Channel
			channelID := getString(params, "channel_id")
			// Also support group_id as alias if strictly needed, but prefer channel_id
			if channelID == "" {
				channelID = getString(params, "group_id")
			}
			safeContent, imagePath, _ := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("channel", channelID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Channel %s", msgID, channelID)
				} else {
					log.Printf("[NEXUS-MSG] No active session (MsgID) for Channel %s. Caching action pending channel reply.", channelID)
					sessionCache.AddPending("channel", channelID, action)
					return
				}
			}

			log.Printf("[NEXUS-MSG] Sending Guild Message to %s: %s (Img: %s)", channelID, safeContent, imagePath)
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			}
			if imagePath != "" {
				// msgData.FileImage = imagePath // Not supported
				safeContent += "\n[Image not supported in Guild Channel]"
				msgData.Content = safeContent
				if !strings.HasPrefix(imagePath, "http") {
					os.Remove(imagePath)
				}
			}

			msg, err := api.PostMessage(ctx, channelID, msgData)
			if err != nil {
				log.Printf("[NEXUS-MSG] Failed to send guild message: %v", err)
			} else {
				log.Printf("[NEXUS-MSG] Guild message sent successfully")
			}
			handleSendResponse(err, msg, action)
		}

	case "send_group_msg":
		// Strictly for QQ Groups
		params, _ := action["params"].(map[string]interface{})
		groupID := getString(params, "group_id")
		content, _ := params["message"].(string)
		safeContent, imagePath, fileType := cleanContent(content)

		// Try to find session if message_id is missing
		msgID := getString(params, "message_id")
		if msgID == "" {
			msgID = sessionCache.Get("group", groupID)
			if msgID != "" {
				log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Group %s", msgID, groupID)
			}
		}

		log.Printf("[NEXUS-MSG] Sending Group Message (send_group_msg) to %s: %s", groupID, safeContent)

		var err error
		var media *dto.MediaInfo

		if imagePath != "" {
			fileInfo, errUpload := uploadGroupFile(groupID, imagePath, fileType)
			if errUpload != nil {
				log.Printf("[NEXUS-MSG] Failed to upload group file: %v", errUpload)
				err = errUpload
				safeContent += "\n[Image Upload Failed: Local file upload not supported, use URL]"
			} else {
				log.Printf("[NEXUS-MSG] Group file uploaded successfully")
				media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
			}
			if !strings.HasPrefix(imagePath, "http") {
				os.Remove(imagePath)
			}
		}

		// Send message if content is not empty or media is present
		if safeContent != "" || media != nil {
			if safeContent == "" {
				safeContent = " "
			}
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
				Media:   media,
			}
			_, errPost := api.PostGroupMessage(ctx, groupID, msgData)
			if errPost != nil {
				log.Printf("[NEXUS-MSG] Failed to send group message: %v", errPost)
				if err == nil {
					err = errPost
				}
			} else {
				log.Printf("[NEXUS-MSG] Group message sent successfully")
			}
		}

		handleSendResponse(err, nil, action)

	case "send_private_msg":
		// Strictly for C2C
		params, _ := action["params"].(map[string]interface{})
		userID := getString(params, "user_id")
		content, _ := params["message"].(string)
		safeContent, imagePath, fileType := cleanContent(content)

		// Try to find session if message_id is missing
		msgID := getString(params, "message_id")
		if msgID == "" {
			msgID = sessionCache.Get("user", userID)
			if msgID != "" {
				log.Printf("[NEXUS-MSG] Using cached session MsgID %s for User %s", msgID, userID)
			}
		}

		log.Printf("[NEXUS-MSG] Sending Private Message (send_private_msg) to %s: %s (Img: %s)", userID, safeContent, imagePath)

		var media *dto.MediaInfo
		if imagePath != "" {
			fileInfo, errUpload := uploadC2CFile(userID, imagePath, fileType)
			if errUpload != nil {
				log.Printf("[NEXUS-MSG] Failed to upload C2C file: %v", errUpload)
				safeContent += "\n[Image Upload Failed: Local file upload not supported, use URL]"
			} else {
				media = &dto.MediaInfo{FileInfo: []byte(fileInfo)}
			}
			if !strings.HasPrefix(imagePath, "http") {
				os.Remove(imagePath)
			}
		}

		if safeContent == "" {
			safeContent = " "
		}

		msgData := &dto.MessageToCreate{
			Content: safeContent,
			MsgID:   msgID,
			MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			Media:   media,
		}
		_, err := api.PostC2CMessage(ctx, userID, msgData)
		if err != nil {
			log.Printf("[NEXUS-MSG] Failed to send private message: %v", err)
		} else {
			log.Printf("[NEXUS-MSG] Private message sent successfully")
		}
		handleSendResponse(err, nil, action)

	case "send_guild_channel_msg":
		// Strictly for Guild Channels
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		channelID := getString(params, "channel_id")
		content, _ := params["message"].(string)

		// Note: api.PostMessage only needs channelID. GuildID is extra context.
		// However, if we only have GuildID, we can't send.
		if channelID == "" {
			log.Println("send_guild_channel_msg requires channel_id")
			sendToNexus(map[string]interface{}{"status": "failed", "message": "missing channel_id", "echo": action["echo"]})
			return
		}

		// Try to find session if message_id is missing
		msgID := getString(params, "message_id")
		if msgID == "" {
			msgID = sessionCache.Get("channel", channelID)
			if msgID != "" {
				log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Channel %s", msgID, channelID)
			}
		}

		log.Printf("Sending to Guild %s Channel %s: %s", guildID, channelID, content)
		msg, err := api.PostMessage(ctx, channelID, &dto.MessageToCreate{
			Content: content,
			MsgID:   msgID,
			MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
		})
		handleSendResponse(err, msg, action)

	case "delete_msg":
		params, _ := action["params"].(map[string]interface{})
		messageID := getString(params, "message_id")
		// We need channel_id to retract. Try to find it in params (non-standard)
		channelID := getString(params, "group_id") // Reuse group_id as channel_id

		if messageID != "" && channelID != "" {
			err := api.RetractMessage(ctx, channelID, messageID)
			if err != nil {
				log.Println("Error retracting message:", err)
			}
			sendToNexus(map[string]interface{}{
				"status": "ok",
				"echo":   action["echo"],
			})
		} else {
			log.Println("delete_msg requires message_id and group_id (channel_id)")
		}

	case "get_login_info":
		// Return bot info
		me, err := api.Me(ctx)
		if err == nil {
			resp := map[string]interface{}{
				"status": "ok",
				"data": map[string]interface{}{
					"user_id":  me.ID,
					"nickname": me.Username,
				},
				"echo": action["echo"],
			}
			sendToNexus(resp)
		}

	case "get_group_list":
		// Return empty list as we can't easily fetch joined groups yet
		// User explicitly requested separation of Groups and Guilds
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   []interface{}{},
			"echo":   action["echo"],
		})

	case "get_group_count":
		// For official bots, map Guilds to Groups for compatibility/visibility
		// Fetch Guilds to count
		guilds, err := api.MeGuilds(ctx, &dto.GuildPager{Limit: "100"})
		count := 0
		if err == nil {
			count = len(guilds)
		}
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"count": count,
			},
			"echo": action["echo"],
		})

	case "get_friend_count":
		// Return count of friends (0 for now, as official bots don't have friends)
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"count": 0,
			},
			"echo": action["echo"],
		})

	case "get_guild_count":
		// Fetch Guilds to count
		// Note: This is still somewhat expensive if we have many guilds, but saves bandwidth to Nexus
		guilds, err := api.MeGuilds(ctx, &dto.GuildPager{Limit: "100"})
		count := 0
		if err == nil {
			count = len(guilds)
		}
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"count": count,
			},
			"echo": action["echo"],
		})

	case "get_guild_list":
		// Fetch Guilds
		guilds, err := api.MeGuilds(ctx, &dto.GuildPager{Limit: "100"})
		if err != nil {
			log.Println("Error getting guilds:", err)
			sendToNexus(map[string]interface{}{
				"status":  "failed",
				"retcode": 100,
				"echo":    action["echo"],
			})
			return
		}

		var guildList []map[string]interface{}
		for _, guild := range guilds {
			guildList = append(guildList, map[string]interface{}{
				"guild_id":         guild.ID,
				"guild_name":       guild.Name,
				"member_count":     guild.MemberCount,
				"max_member_count": guild.MaxMembers,
			})
		}

		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   guildList,
			"echo":   action["echo"],
		})

	case "get_guild_channel_list":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		if guildID == "" {
			sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})
			return
		}

		channels, err := api.Channels(ctx, guildID)
		if err != nil {
			sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})
			return
		}

		var channelList []map[string]interface{}
		for _, channel := range channels {
			channelList = append(channelList, map[string]interface{}{
				"guild_id":     guildID,
				"channel_id":   channel.ID,
				"channel_name": channel.Name,
				"channel_type": channel.Type,
			})
		}

		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   channelList,
			"echo":   action["echo"],
		})

	case "get_version_info":
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"app_name":         "TencentBot",
				"app_version":      "1.0.0",
				"protocol_version": "v11",
			},
			"echo": action["echo"],
		})

	case "get_group_info":
		// As requested, Groups and Guilds are separate.
		// Since we don't have full Group API access yet, return mock/empty or specific error
		sendToNexus(map[string]interface{}{
			"status":  "failed",
			"retcode": 100,
			"message": "get_group_info not fully supported for QQ Groups yet",
			"echo":    action["echo"],
		})

	default:

	case "get_friend_list":
		// Official bots don't have friends in the traditional sense
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   []interface{}{},
			"echo":   action["echo"],
		})

	case "get_group_member_list":
		// Strict QQ Group implementation: Currently not supported by official API for bots in this manner
		// Return empty list or specific error to indicate separation from Guilds
		sendToNexus(map[string]interface{}{
			"status":  "failed",
			"retcode": 100,
			"message": "get_group_member_list not supported for QQ Groups yet",
			"data":    []interface{}{},
			"echo":    action["echo"],
		})

	case "get_group_member_info":
		// Strict QQ Group implementation
		sendToNexus(map[string]interface{}{
			"status":  "failed",
			"retcode": 100,
			"message": "get_group_member_info not supported for QQ Groups yet",
			"echo":    action["echo"],
		})

	case "get_guild_member_list":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		// Optional: next_token/limit for pagination, but OneBot usually expects full list or handled differently
		// Tencent API uses limit/after
		limit := "400" // Max limit
		if l, ok := params["limit"].(string); ok {
			limit = l
		}

		var members []map[string]interface{}

		if guildID != "" {
			guildMembers, err := api.GuildMembers(ctx, guildID, &dto.GuildMembersPager{Limit: limit})
			if err == nil {
				for _, m := range guildMembers {
					members = append(members, map[string]interface{}{
						"guild_id":  guildID,
						"user_id":   m.User.ID,
						"nickname":  m.User.Username,
						"card":      m.Nick, // Guild Nickname
						"role":      getRoleName(m.Roles),
						"join_time": parseTimestamp(m.JoinedAt),
						"title":     "", // Not supported
					})
				}
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data":   members,
					"echo":   action["echo"],
				})
				return
			} else {
				log.Println("Error getting guild members:", err)
			}
		}
		sendToNexus(map[string]interface{}{
			"status": "failed",
			"echo":   action["echo"],
		})

	case "get_guild_member_profile":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		userID := getString(params, "user_id")

		if guildID != "" && userID != "" {
			m, err := api.GuildMember(ctx, guildID, userID)
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data": map[string]interface{}{
						"guild_id":  guildID,
						"user_id":   m.User.ID,
						"nickname":  m.User.Username,
						"card":      m.Nick,
						"role":      getRoleName(m.Roles),
						"join_time": parseTimestamp(m.JoinedAt),
					},
					"echo": action["echo"],
				})
				return
			} else {
				log.Println("Error getting guild member profile:", err)
			}
		}
		sendToNexus(map[string]interface{}{
			"status": "failed",
			"echo":   action["echo"],
		})

	case "set_group_kick":
		// Strict QQ Group implementation
		sendToNexus(map[string]interface{}{
			"status":  "failed",
			"retcode": 100,
			"message": "set_group_kick not supported for QQ Groups yet",
			"echo":    action["echo"],
		})

	case "set_guild_kick": // or delete_guild_member
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		userID := getString(params, "user_id")
		// Some implementations might use delete_guild_member
		if guildID != "" && userID != "" {
			err := api.DeleteGuildMember(ctx, guildID, userID)
			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
				return
			} else {
				log.Println("Error kicking guild member:", err)
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "get_guild_meta":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		if guildID != "" {
			guild, err := api.Guild(ctx, guildID)
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data": map[string]interface{}{
						"guild_id":         guild.ID,
						"guild_name":       guild.Name,
						"member_count":     guild.MemberCount,
						"max_member_count": guild.MaxMembers,
						// "description":      guild.Description, // Not supported in v0.2.1
						"joined_at": parseTimestamp(guild.JoinedAt),
					},
					"echo": action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "create_guild_channel":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		name := getString(params, "name")
		cTypeVal, _ := params["type"].(float64)
		parentID := getString(params, "parent_id")

		if guildID != "" && name != "" {
			channel, err := api.PostChannel(ctx, guildID, &dto.ChannelValueObject{
				Name:     name,
				Type:     dto.ChannelType(cTypeVal),
				ParentID: parentID,
			})
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data":   map[string]interface{}{"channel_id": channel.ID},
					"echo":   action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "update_guild_channel":
		params, _ := action["params"].(map[string]interface{})
		channelID := getString(params, "channel_id")
		name := getString(params, "name")
		cTypeVal, _ := params["type"].(float64)

		if channelID != "" && name != "" {
			// Note: Type might not be updatable in some contexts, but SDK allows it in struct
			channel, err := api.PatchChannel(ctx, channelID, &dto.ChannelValueObject{
				Name: name,
				Type: dto.ChannelType(cTypeVal),
			})
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data":   map[string]interface{}{"channel_id": channel.ID},
					"echo":   action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "delete_guild_channel":
		params, _ := action["params"].(map[string]interface{})
		channelID := getString(params, "channel_id")
		if channelID != "" {
			err := api.DeleteChannel(ctx, channelID)
			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "get_guild_roles":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		if guildID != "" {
			roles, err := api.Roles(ctx, guildID)
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data":   roles,
					"echo":   action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "create_guild_role":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		name := getString(params, "name")
		colorVal, _ := params["color"].(float64)
		hoistVal, _ := params["hoist"].(float64) // 0 or 1

		if guildID != "" {
			_, err := api.PostRole(ctx, guildID, &dto.Role{
				Name:  name,
				Color: uint32(colorVal),
				Hoist: uint32(hoistVal),
			})
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					// "data":   map[string]interface{}{"role_id": role.ID}, // role.ID undefined in v0.2.1 UpdateResult
					"echo": action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "update_guild_role":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		roleID := getString(params, "role_id")
		name := getString(params, "name")
		colorVal, _ := params["color"].(float64)
		hoistVal, _ := params["hoist"].(float64)

		if guildID != "" && roleID != "" {
			_, err := api.PatchRole(ctx, guildID, dto.RoleID(roleID), &dto.Role{
				Name:  name,
				Color: uint32(colorVal),
				Hoist: uint32(hoistVal),
			})
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					// "data":   map[string]interface{}{"role_id": role.ID},
					"echo": action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "delete_guild_role":
		params, _ := action["params"].(map[string]interface{})
		guildID := getString(params, "guild_id")
		roleID := getString(params, "role_id")
		if guildID != "" && roleID != "" {
			err := api.DeleteRole(ctx, guildID, dto.RoleID(roleID))
			if err == nil {
				sendToNexus(map[string]interface{}{"status": "ok", "echo": action["echo"]})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "get_message":
		params, _ := action["params"].(map[string]interface{})
		channelID := getString(params, "channel_id")
		messageID := getString(params, "message_id")

		if channelID != "" && messageID != "" {
			msg, err := api.Message(ctx, channelID, messageID)
			if err == nil {
				sendToNexus(map[string]interface{}{
					"status": "ok",
					"data":   msg,
					"echo":   action["echo"],
				})
				return
			}
		}
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "set_group_ban":
		// Mute member not fully supported in this SDK version or requires different API
		log.Println("set_group_ban not implemented yet")
		sendToNexus(map[string]interface{}{"status": "failed", "echo": action["echo"]})

	case "get_logs":
		// Fetch logs from memory buffer
		lines := 100 // Default
		if params, ok := action["params"].(map[string]interface{}); ok {
			if lStr := getString(params, "lines"); lStr != "" {
				if l, err := strconv.Atoi(lStr); err == nil {
					lines = l
				}
			}
		}

		logs := []string{}
		if logManager != nil {
			logs = logManager.GetLogs(lines)
		}

		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   logs,
			"echo":   action["echo"],
		})
	}
}

// Helper to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatFloat(v, 'f', 0, 64)
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func getRoleName(roles []string) string {
	if len(roles) > 0 {
		return "member" // Simplify for now, roles are IDs usually
	}
	return "member"
}

func parseTimestamp(t dto.Timestamp) int64 {
	ts, err := time.Parse(time.RFC3339, string(t))
	if err != nil {
		return 0
	}
	return ts.Unix()
}

func handleSendResponse(err error, msg *dto.Message, action map[string]interface{}) {
	if err != nil {
		log.Println("Error sending message:", err)
		sendToNexus(map[string]interface{}{
			"status":  "failed",
			"retcode": 100,
			"data":    nil,
			"message": err.Error(),
			"echo":    action["echo"],
		})
	} else {
		respData := map[string]interface{}{}
		if msg != nil {
			respData["message_id"] = msg.ID
		}
		sendToNexus(map[string]interface{}{
			"status": "ok",
			"data":   respData,
			"echo":   action["echo"],
		})
	}
}

func sendToNexus(data interface{}) {
	nexusMu.Lock()
	defer nexusMu.Unlock()
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(data); err != nil {
		log.Println("Error sending to Nexus:", err)
	}
}

// Event Handlers

func atMessageEventHandler(event *dto.WSPayload, data *dto.WSATMessageData) error {
	// 打印收到的完整消息信息
	log.Printf("=== 收到频道消息 ===")
	log.Printf("消息ID: %s", data.ID)
	log.Printf("频道ID: %s", data.ChannelID)
	log.Printf("频道名称: %s", data.ChannelID)
	log.Printf("用户ID: %s", data.Author.ID)
	log.Printf("用户名: %s", data.Author.Username)
	log.Printf("消息内容: %s", data.Content)
	log.Printf("消息时间: %d", data.Timestamp)
	log.Printf("完整数据结构: %+v", data)

	// Save Session for Reply
	pending := sessionCache.Save("channel", data.ChannelID, data.ID)
	for _, action := range pending {
		go handleAction(action)
	}

	// Test: Reply with Avatar
	if strings.Contains(data.Content, "头像") || strings.EqualFold(strings.TrimSpace(data.Content), "avatar") {
		avatar := data.Author.Avatar
		if avatar == "" {
			avatar = "Avatar URL not found"
		}
		log.Printf("Replying with Avatar: %s", avatar)
		_, err := api.PostMessage(context.Background(), data.ChannelID, &dto.MessageToCreate{
			Content: fmt.Sprintf("Your Avatar: %s", avatar),
			MsgID:   data.ID,
			MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
		})
		if err != nil {
			log.Printf("Failed to reply avatar: %v", err)
		}
	}

	// Translate to OneBot v11 Message Event
	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "guild", // Guild messages are distinct from group
		"sub_type":     "normal",
		"message_id":   data.ID,
		"user_id":      data.Author.ID, // String ID
		"guild_id":     data.GuildID,
		"channel_id":   data.ChannelID,
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": data.Author.Username,
		},
		"time":     time.Now().Unix(),
		"self_id":  selfID,
		"platform": "qqguild", // 添加平台标识
	}

	log.Printf("转换为OneBot v11格式: %+v", obEvent)
	sendToNexus(obEvent)

	return nil
}

func directMessageEventHandler(event *dto.WSPayload, data *dto.WSDirectMessageData) error {
	// 打印收到的完整私信信息
	log.Printf("=== 收到私信消息 ===")
	log.Printf("消息ID: %s", data.ID)
	log.Printf("频道ID: %s", data.ChannelID)
	log.Printf("用户ID: %s", data.Author.ID)
	log.Printf("用户名: %s", data.Author.Username)
	log.Printf("消息内容: %s", data.Content)
	log.Printf("消息时间: %d", data.Timestamp)
	log.Printf("完整数据结构: %+v", data)

	// Save Session for Reply (DM uses ChannelID)
	pending := sessionCache.Save("channel", data.ChannelID, data.ID)
	for _, action := range pending {
		go handleAction(action)
	}

	// Translate to OneBot v11 Message Event
	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "private",
		"sub_type":     "friend",
		"message_id":   data.ID,
		"user_id":      data.Author.ID,
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": data.Author.Username,
		},
		"time":     time.Now().Unix(),
		"self_id":  selfID,
		"platform": "qqguild", // 添加平台标识
	}

	log.Printf("转换为OneBot v11格式: %+v", obEvent)
	sendToNexus(obEvent)

	return nil
}

func guildEventHandler(event *dto.WSPayload, data *dto.WSGuildData) error {
	log.Printf("Guild Event: %s, Guild: %s(%s)", event.Type, data.Name, data.ID)
	sendToNexus(map[string]interface{}{
		"post_type":   "notice",
		"notice_type": "guild_event",
		"sub_type":    event.Type,
		"guild_id":    data.ID,
		"guild_name":  data.Name,
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"platform":    "qqguild", // 添加平台标识
	})
	return nil
}

func guildMemberEventHandler(event *dto.WSPayload, data *dto.WSGuildMemberData) error {
	log.Printf("Member Event: %s, User: %s", event.Type, data.User.Username)
	noticeType := "group_decrease"
	if event.Type == "GUILD_MEMBER_ADD" {
		noticeType = "group_increase"
	}
	sendToNexus(map[string]interface{}{
		"post_type":   "notice",
		"notice_type": noticeType,
		"group_id":    data.GuildID,
		"user_id":     data.User.ID,
		"operator_id": data.OpUserID,
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"platform":    "qqguild", // 添加平台标识
	})
	return nil
}

func channelEventHandler(event *dto.WSPayload, data *dto.WSChannelData) error {
	log.Printf("Channel Event: %s, Channel: %s(%s)", event.Type, data.Name, data.ID)
	sendToNexus(map[string]interface{}{
		"post_type":    "notice",
		"notice_type":  "channel_event",
		"sub_type":     event.Type,
		"group_id":     data.ID,
		"guild_id":     data.GuildID,
		"channel_name": data.Name,
		"time":         time.Now().Unix(),
		"self_id":      selfID,
		"platform":     "qqguild", // 添加平台标识
	})
	return nil
}

func messageReactionEventHandler(event *dto.WSPayload, data *dto.WSMessageReactionData) error {
	log.Printf("Reaction Event: %s", event.Type)
	// data.Target.ID is usually where the message ID is
	msgID := ""
	if data.Target.ID != "" {
		msgID = data.Target.ID
	}
	sendToNexus(map[string]interface{}{
		"post_type":   "notice",
		"notice_type": "group_card", // Using group_card as placeholder for reaction
		"sub_type":    "reaction",
		"group_id":    data.ChannelID,
		"user_id":     data.UserID,
		"message_id":  msgID,
		"emoji":       data.Emoji,
		"time":        time.Now().Unix(),
		"self_id":     selfID,
		"platform":    "qqguild", // 添加平台标识
	})
	return nil
}

func groupATMessageEventHandler(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	// 打印收到的完整群消息信息
	log.Printf("=== 收到群消息 ===")
	log.Printf("消息ID: %s", data.ID)
	log.Printf("群ID: %s", data.GroupID)
	log.Printf("用户ID: %s", data.Author.ID)
	log.Printf("用户名: %s", data.Author.Username)
	log.Printf("消息内容: %s", data.Content)
	log.Printf("消息时间: %d", data.Timestamp)
	log.Printf("完整数据结构: %+v", data)

	// Save Session for Reply
	pending := sessionCache.Save("group", data.GroupID, data.ID)
	for _, action := range pending {
		go handleAction(action)
	}

	// Test: Reply with Avatar
	if strings.Contains(data.Content, "头像") || strings.EqualFold(strings.TrimSpace(data.Content), "avatar") {
		avatar := data.Author.Avatar
		if avatar == "" {
			avatar = "https://q1.qlogo.cn/g?b=qq&nk=1653346663&s=100"
		}
		log.Printf("Replying with Avatar (Uploading...): %s", avatar)

		// Upload file to get FileInfo
		fileInfo, err := uploadGroupFile(data.GroupID, avatar, 1) // 1 = Image
		if err != nil {
			log.Printf("Failed to upload avatar via URL: %v. Trying proxy...", err)
			// Fallback: Download and Proxy
			resp, errDl := http.Get(avatar)
			if errDl == nil {
				defer resp.Body.Close()
				tmpFile, errTmp := ioutil.TempFile("", "avatar_*.jpg")
				if errTmp == nil {
					defer tmpFile.Close()
					_, _ = io.Copy(tmpFile, resp.Body)
					tmpPath := tmpFile.Name()
					// Upload local file
					fileInfo, err = uploadGroupFile(data.GroupID, tmpPath, 1)
					os.Remove(tmpPath)
				}
			}
		}

		if err != nil {
			log.Printf("Failed to upload avatar: %v", err)
			api.PostGroupMessage(context.Background(), data.GroupID, &dto.MessageToCreate{
				Content: fmt.Sprintf("Avatar Upload Failed: %v", err),
				MsgID:   data.ID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			})
		} else {
			_, err := api.PostGroupMessage(context.Background(), data.GroupID, &dto.MessageToCreate{
				Content: " ",
				MsgType: 7,
				Media:   &dto.MediaInfo{FileInfo: []byte(fileInfo)},
				MsgID:   data.ID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			})
			if err != nil {
				log.Printf("Failed to reply avatar media: %v", err)
			}
		}
	}

	// 处理附件
	for _, attachment := range data.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image") {
			data.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		} else if strings.HasPrefix(attachment.ContentType, "video") {
			data.Content += fmt.Sprintf("[CQ:video,file=%s]", attachment.URL)
		} else if strings.HasPrefix(attachment.ContentType, "audio") {
			data.Content += fmt.Sprintf("[CQ:record,file=%s]", attachment.URL)
		} else {
			// Generic File
			data.Content += fmt.Sprintf("[CQ:file,file=%s,name=%s]", attachment.URL, filepath.Base(attachment.URL))
		}
	}

	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "group",
		"sub_type":     "normal",
		"message_id":   data.ID,
		"user_id":      data.Author.ID,
		"group_id":     data.GroupID, // Changed from GroupOpenID to GroupID
		"anonymous":    nil,          // 标准OneBot v11要求，非匿名消息为null
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": "Group Member",
		},
		"time":     time.Now().Unix(),
		"self_id":  selfID,
		"platform": "qqguild", // 修正为统一的平台标识
	}

	log.Printf("转换为OneBot v11格式: %+v", obEvent)
	sendToNexus(obEvent)

	return nil
}

func c2cMessageEventHandler(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
	// 打印收到的完整私聊消息信息
	log.Printf("=== 收到私聊消息 ===")
	log.Printf("消息ID: %s", data.ID)
	log.Printf("用户ID: %s", data.Author.ID)
	log.Printf("用户名: %s", data.Author.Username)
	log.Printf("消息内容: %s", data.Content)
	log.Printf("消息时间: %d", data.Timestamp)
	log.Printf("完整数据结构: %+v", data)

	// Save Session for Reply
	pending := sessionCache.Save("user", data.Author.ID, data.ID)
	for _, action := range pending {
		go handleAction(action)
	}

	// Test: Reply with Avatar
	if strings.Contains(data.Content, "头像") || strings.EqualFold(strings.TrimSpace(data.Content), "avatar") {
		avatar := data.Author.Avatar
		if avatar == "" {
			// Fallback to a default avatar
			avatar = "https://q1.qlogo.cn/g?b=qq&nk=1653346663&s=100"
		}
		log.Printf("Replying with Avatar (Uploading...): %s", avatar)

		// Upload file to get FileInfo
		fileInfo, err := uploadC2CFile(data.Author.ID, avatar, 1) // 1 = Image
		if err != nil {
			log.Printf("Failed to upload avatar via URL: %v. Trying proxy...", err)
			// Fallback: Download and Proxy
			resp, errDl := http.Get(avatar)
			if errDl == nil {
				defer resp.Body.Close()
				tmpFile, errTmp := ioutil.TempFile("", "avatar_*.jpg")
				if errTmp == nil {
					defer tmpFile.Close()
					_, _ = io.Copy(tmpFile, resp.Body)
					tmpPath := tmpFile.Name()
					// Upload local file
					fileInfo, err = uploadC2CFile(data.Author.ID, tmpPath, 1)
					os.Remove(tmpPath)
				}
			}
		}

		if err != nil {
			log.Printf("Failed to upload avatar: %v", err)
			api.PostC2CMessage(context.Background(), data.Author.ID, &dto.MessageToCreate{
				Content: fmt.Sprintf("Avatar Upload Failed: %v", err),
				MsgID:   data.ID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			})
		} else {
			_, err := api.PostC2CMessage(context.Background(), data.Author.ID, &dto.MessageToCreate{
				Content: " ",
				MsgType: 7,
				Media:   &dto.MediaInfo{FileInfo: []byte(fileInfo)},
				MsgID:   data.ID,
				MsgSeq:  uint32(atomic.AddInt64(&msgSeq, 1)),
			})
			if err != nil {
				log.Printf("Failed to reply avatar media: %v", err)
			}
		}
	}

	// Handle Attachments (Images, Video, Audio, Files)
	for _, attachment := range data.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image") {
			data.Content += fmt.Sprintf("[CQ:image,file=%s]", attachment.URL)
		} else if strings.HasPrefix(attachment.ContentType, "video") {
			data.Content += fmt.Sprintf("[CQ:video,file=%s]", attachment.URL)
		} else if strings.HasPrefix(attachment.ContentType, "audio") {
			data.Content += fmt.Sprintf("[CQ:record,file=%s]", attachment.URL)
		} else {
			// Generic File
			data.Content += fmt.Sprintf("[CQ:file,file=%s,name=%s]", attachment.URL, filepath.Base(attachment.URL))
		}
	}

	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "private",
		"sub_type":     "friend",
		"message_id":   data.ID,
		"user_id":      data.Author.ID,
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": "Friend",
		},
		"time":     time.Now().Unix(),
		"self_id":  selfID,
		"platform": "qqguild", // 添加平台标识
	}

	log.Printf("转换为OneBot v11格式: %+v", obEvent)
	sendToNexus(obEvent)

	return nil
}

// Type assertions to verify handler signatures
var _ event.ATMessageEventHandler = atMessageEventHandler
var _ event.DirectMessageEventHandler = directMessageEventHandler
var _ event.GuildEventHandler = guildEventHandler
var _ event.GuildMemberEventHandler = guildMemberEventHandler
var _ event.ChannelEventHandler = channelEventHandler
var _ event.MessageReactionEventHandler = messageReactionEventHandler

// var _ event.GroupATMessageEventHandler = groupATMessageEventHandler
// var _ event.C2CMessageEventHandler = c2cMessageEventHandler

func main() {
	// Initialize Logger
	logManager = NewLogManager(2000) // Keep last 2000 lines
	log.SetOutput(logManager)

	loadConfig()
	ctx = context.Background()

	// Load Session Cache
	sessionCache.LoadDisk()

	// Start HTTP Log Viewer and File Server if configured
	if config.LogPort > 0 {
		go func() {
			http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
				lines := 100
				if l := r.URL.Query().Get("lines"); l != "" {
					if v, err := strconv.Atoi(l); err == nil {
						lines = v
					}
				}

				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				logs := logManager.GetLogs(lines)
				for _, line := range logs {
					fmt.Fprintln(w, line)
				}
			})

			// Serve temporary media files
			// Maps config.MediaRoute + filename -> os.TempDir()/filename
			http.HandleFunc(config.MediaRoute, func(w http.ResponseWriter, r *http.Request) {
				fileName := strings.TrimPrefix(r.URL.Path, config.MediaRoute)
				if fileName == "" || strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
					http.Error(w, "Invalid file name", http.StatusBadRequest)
					return
				}
				// Look in temp dir
				tmpPath := filepath.Join(os.TempDir(), fileName)
				// Check existence
				if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
					http.Error(w, "File not found", http.StatusNotFound)
					return
				}

				// Log access
				log.Printf("Serving media: %s to %s (UA: %s)", fileName, r.RemoteAddr, r.UserAgent())

				// Explicitly set Content-Type based on extension
				ext := strings.ToLower(filepath.Ext(fileName))
				switch ext {
				case ".png":
					w.Header().Set("Content-Type", "image/png")
				case ".jpg", ".jpeg":
					w.Header().Set("Content-Type", "image/jpeg")
				case ".gif":
					w.Header().Set("Content-Type", "image/gif")
				case ".mp4":
					w.Header().Set("Content-Type", "video/mp4")
				case ".amr":
					w.Header().Set("Content-Type", "audio/amr")
				}

				http.ServeFile(w, r, tmpPath)
			})

			addr := fmt.Sprintf(":%d", config.LogPort)
			log.Printf("Starting HTTP Server at http://localhost%s (Logs: /logs, Media: %s)", addr, config.MediaRoute)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("Failed to start HTTP Server: %v", err)
			}
		}()
	}

	// Initialize Bot Token
	botToken := token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     fmt.Sprintf("%d", config.AppID),
			AppSecret: config.Secret,
		},
	)

	// Initialize API
	if config.Sandbox {
		log.Println("Initializing Tencent Bot API in SANDBOX mode...")
		api = botgo.NewSandboxOpenAPI(fmt.Sprintf("%d", config.AppID), botToken).WithTimeout(3 * time.Second)
	} else {
		log.Println("Initializing Tencent Bot API in PRODUCTION mode...")
		api = botgo.NewOpenAPI(fmt.Sprintf("%d", config.AppID), botToken).WithTimeout(3 * time.Second)
	}

	// Get Bot Info (SelfID)
	// Always try to get nickname from API for better UX
	me, err := api.Me(ctx)
	if err != nil {
		log.Printf("Error getting bot info: %v", err)
	}

	if config.SelfID != "" {
		selfID = config.SelfID
		nickname := "Unknown"
		if err == nil {
			nickname = me.Username
		}
		log.Printf("Using configured Bot SelfID: %s, Nickname: %s", selfID, nickname)
	} else {
		if err == nil {
			selfID = me.ID
			log.Printf("Using Bot SelfID from API: %s, Nickname: %s", selfID, me.Username)
		} else {
			log.Printf("Error getting bot info and no SelfID configured. Using AppID as fallback (Not recommended).")
			selfID = fmt.Sprintf("%d", config.AppID)
		}
	}

	// Connect to BotNexus
	go NexusConnect()

	// Connect to Tencent WebSocket
	go func() {
		wsInfo, err := api.WS(ctx, nil, "")
		if err != nil {
			log.Fatal("Error getting WS info:", err)
		}

		// Register handlers using event package and explicit casting
		intent := event.RegisterHandlers(
			event.ATMessageEventHandler(atMessageEventHandler),
			event.DirectMessageEventHandler(directMessageEventHandler),
			event.GuildEventHandler(guildEventHandler),
			event.GuildMemberEventHandler(guildMemberEventHandler),
			event.ChannelEventHandler(channelEventHandler),
			event.MessageReactionEventHandler(messageReactionEventHandler),
			event.GroupATMessageEventHandler(groupATMessageEventHandler),
			event.C2CMessageEventHandler(c2cMessageEventHandler),
		)
		log.Printf("Calculated Intent from Handlers: %d", intent)

		// Explicitly enable intents to ensure they are active
		// 1<<30 (Public/At Messages) | 1<<12 (Direct Messages) | 1<<0 (Guilds)
		// 1<<1 (Guild Members) | 1<<10 (Guild Message Reactions)
		// 1<<25 (Group & C2C)
		// 1<<9 (Guild Messages - for private bots ONLY, causes 4014 in sandbox/public)
		// Forum/Audio/Interaction removed as not fully supported in v0.2.1
		intent = intent | (1 << 30) | (1 << 12) | (1 << 0) | (1 << 1) | (1 << 10) | (1 << 25)

		log.Printf("Final Intent after manual override: %d", intent)

		log.Printf("Starting Tencent Bot Session Manager with Intent: %d...", intent)
		if err := botgo.NewSessionManager().Start(wsInfo, botToken, &intent); err != nil {
			log.Fatal(err)
		}
	}()

	// Keep alive
	select {}
}
