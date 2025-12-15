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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	AppID     uint64 `json:"app_id"` // AppID is uint64 in SDK
	Token     string `json:"token"`
	Secret    string `json:"secret"`
	Sandbox   bool   `json:"sandbox"`
	SelfID    string `json:"self_id"` // Optional: manually set SelfID
	NexusAddr string `json:"nexus_addr"`
}

var (
	config    Config
	nexusConn *websocket.Conn
	api       openapi.OpenAPI
	ctx       context.Context
	selfID    string
)

// SessionCache to store last message ID for replying
type SessionCache struct {
	sync.RWMutex
	UserLastMsgID    map[string]string // UserID -> MsgID (C2C)
	GroupLastMsgID   map[string]string // GroupID -> MsgID (Group)
	ChannelLastMsgID map[string]string // ChannelID -> MsgID (Guild)
	LastMsgTime      map[string]int64  // MsgID -> Timestamp (Unix)
}

var sessionCache = &SessionCache{
	UserLastMsgID:    make(map[string]string),
	GroupLastMsgID:   make(map[string]string),
	ChannelLastMsgID: make(map[string]string),
	LastMsgTime:      make(map[string]int64),
}

func (s *SessionCache) Save(keyType, key, msgID string) {
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

	// Validation
	if config.AppID == 0 || config.Token == "" || config.Secret == "" {
		log.Fatal("Missing configuration. Please check config.json or environment variables (TENCENT_APP_ID, TENCENT_TOKEN, TENCENT_SECRET).")
	}
	if config.NexusAddr == "" {
		config.NexusAddr = "ws://192.168.0.167:3005"
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

func uploadGroupFile(groupID string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	io.Copy(part, file)

	_ = writer.WriteField("file_type", "1")
	_ = writer.WriteField("srv_send_msg", "true")

	err = writer.Close()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.sgroup.qq.com/v2/groups/%s/files", groupID)
	if config.Sandbox {
		url = fmt.Sprintf("https://sandbox.api.sgroup.qq.com/v2/groups/%s/files", groupID)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %d.%s", config.AppID, config.Token))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

func cleanContent(content string) (string, string) {
	// Regex for CQ Image with Base64
	re := regexp.MustCompile(`\[CQ:image,file=base64://([^\]]+)\]`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		b64Data := matches[1]
		data, err := base64.StdEncoding.DecodeString(b64Data)
		if err != nil {
			log.Printf("Error decoding base64 image: %v", err)
			return strings.ReplaceAll(content, matches[0], "[Image Error]"), ""
		}

		tmpFile, err := ioutil.TempFile("", "tencent_img_*.png")
		if err != nil {
			log.Printf("Error creating temp file: %v", err)
			return strings.ReplaceAll(content, matches[0], "[Image Save Error]"), ""
		}
		defer tmpFile.Close()

		if _, err := tmpFile.Write(data); err != nil {
			log.Printf("Error writing to temp file: %v", err)
			return strings.ReplaceAll(content, matches[0], "[Image Write Error]"), ""
		}

		cleanMsg := strings.ReplaceAll(content, matches[0], "")
		return strings.TrimSpace(cleanMsg), tmpFile.Name()
	}

	return content, ""
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
			safeContent, imagePath := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("user", userID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for User %s", msgID, userID)
				}
			}

			log.Printf("[NEXUS-MSG] Sending Private Message to %s: %s (Img: %s)", userID, safeContent, imagePath)
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
			}
			if imagePath != "" {
				// msgData.FileImage = imagePath // FileImage might not be supported in C2C
				safeContent += "\n[Image not supported in Private Chat]"
				msgData.Content = safeContent
				os.Remove(imagePath)
			}

			_, err := api.PostC2CMessage(ctx, userID, msgData)
			if err != nil {
				log.Printf("[NEXUS-MSG] Failed to send private message: %v", err)
			} else {
				log.Printf("[NEXUS-MSG] Private message sent successfully")
			}
			handleSendResponse(err, nil, action)
		} else if messageType == "group" {
			// QQ Group
			groupID := getString(params, "group_id")
			safeContent, imagePath := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("group", groupID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Group %s", msgID, groupID)
				}
			}

			log.Printf("[NEXUS-MSG] Sending Group Message to %s: %s (Img: %s)", groupID, safeContent, imagePath)

			var err error

			// Upload file if exists
			if imagePath != "" {
				err = uploadGroupFile(groupID, imagePath)
				if err != nil {
					log.Printf("[NEXUS-MSG] Failed to upload group file: %v", err)
					safeContent += "\n[Image Upload Failed]"
				} else {
					log.Printf("[NEXUS-MSG] Group file uploaded successfully")
				}
				os.Remove(imagePath)
			}

			// Send text message if content is not empty
			if safeContent != "" {
				msgData := &dto.MessageToCreate{
					Content: safeContent,
					MsgID:   msgID,
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
		} else if messageType == "guild" {
			// Guild Channel
			channelID := getString(params, "channel_id")
			// Also support group_id as alias if strictly needed, but prefer channel_id
			if channelID == "" {
				channelID = getString(params, "group_id")
			}
			safeContent, imagePath := cleanContent(content)

			// Try to find session if message_id is missing
			msgID := getString(params, "message_id")
			if msgID == "" {
				msgID = sessionCache.Get("channel", channelID)
				if msgID != "" {
					log.Printf("[NEXUS-MSG] Using cached session MsgID %s for Channel %s", msgID, channelID)
				}
			}

			log.Printf("[NEXUS-MSG] Sending Guild Message to %s: %s (Img: %s)", channelID, safeContent, imagePath)
			msgData := &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
			}
			if imagePath != "" {
				// msgData.FileImage = imagePath // Not supported
				safeContent += "\n[Image not supported in Guild Channel]"
				msgData.Content = safeContent
				os.Remove(imagePath)
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
		safeContent, imagePath := cleanContent(content)

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
		if imagePath != "" {
			err = uploadGroupFile(groupID, imagePath)
			if err != nil {
				log.Printf("[NEXUS-MSG] Failed to upload group file: %v", err)
				safeContent += "\n[Image Upload Failed]"
			}
			os.Remove(imagePath)
		}

		if safeContent != "" {
			_, errPost := api.PostGroupMessage(ctx, groupID, &dto.MessageToCreate{
				Content: safeContent,
				MsgID:   msgID,
			})
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
		safeContent, imagePath := cleanContent(content)

		if imagePath != "" {
			safeContent += "\n[Image not supported in Private Chat]"
			os.Remove(imagePath)
		}

		// Try to find session if message_id is missing
		msgID := getString(params, "message_id")
		if msgID == "" {
			msgID = sessionCache.Get("user", userID)
			if msgID != "" {
				log.Printf("[NEXUS-MSG] Using cached session MsgID %s for User %s", msgID, userID)
			}
		}

		log.Printf("[NEXUS-MSG] Sending Private Message (send_private_msg) to %s: %s", userID, safeContent)
		_, err := api.PostC2CMessage(ctx, userID, &dto.MessageToCreate{
			Content: safeContent,
			MsgID:   msgID,
		})
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
	}
}

// Helper to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
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
	if nexusConn == nil {
		return
	}
	if err := nexusConn.WriteJSON(data); err != nil {
		log.Println("Error sending to Nexus:", err)
	}
}

// Event Handlers

func atMessageEventHandler(event *dto.WSPayload, data *dto.WSATMessageData) error {
	log.Printf("Received AT Message from %s: %s", data.Author.Username, data.Content)

	// Save Session for Reply
	sessionCache.Save("channel", data.ChannelID, data.ID)

	// Translate to OneBot Message Event
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
		"time":    time.Now().Unix(),
		"self_id": selfID,
	}

	sendToNexus(obEvent)

	return nil
}

func directMessageEventHandler(event *dto.WSPayload, data *dto.WSDirectMessageData) error {
	log.Printf("Received Direct Message from %s: %s", data.Author.Username, data.Content)

	// Save Session for Reply (DM uses ChannelID)
	sessionCache.Save("channel", data.ChannelID, data.ID)

	// Translate to OneBot Message Event
	obEvent := map[string]interface{}{
		"post_type":    "message",
		"message_type": "private",
		"sub_type":     "friend",
		"message_id":   data.ID,
		"user_id":      data.Author.ID,
		"group_id":     "", // Private message
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": data.Author.Username,
		},
		"time":    time.Now().Unix(),
		"self_id": selfID,
	}

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
	})
	return nil
}

func groupATMessageEventHandler(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	log.Printf("Received Group AT Message from %s: %s", data.Author.ID, data.Content)

	// Save Session for Reply
	sessionCache.Save("group", data.GroupID, data.ID)

	sendToNexus(map[string]interface{}{
		"post_type":    "message",
		"message_type": "group",
		"sub_type":     "normal",
		"message_id":   data.ID,
		"user_id":      data.Author.ID,
		"group_id":     data.GroupID, // Changed from GroupOpenID to GroupID
		"message":      data.Content,
		"raw_message":  data.Content,
		"font":         0,
		"sender": map[string]interface{}{
			"user_id":  data.Author.ID,
			"nickname": "Group Member",
		},
		"time":    time.Now().Unix(),
		"self_id": selfID,
	})

	// TEST: Reply with random string
	randomStr := fmt.Sprintf("Group Reply: %d", time.Now().UnixNano())
	log.Printf("Sending Group test reply: %s", randomStr)

	_, err := api.PostGroupMessage(ctx, data.GroupID, &dto.MessageToCreate{
		Content: randomStr,
		MsgID:   data.ID,
	})
	if err != nil {
		log.Printf("Error sending Group test reply: %v", err)
	}

	return nil
}

func c2cMessageEventHandler(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
	log.Printf("Received C2C Message from %s: %s", data.Author.ID, data.Content)

	// Save Session for Reply
	sessionCache.Save("user", data.Author.ID, data.ID)

	sendToNexus(map[string]interface{}{
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
		"time":    time.Now().Unix(),
		"self_id": selfID,
	})

	// TEST: Reply with random string
	randomStr := fmt.Sprintf("C2C Reply: %d", time.Now().UnixNano())
	log.Printf("Sending C2C test reply: %s", randomStr)

	_, err := api.PostC2CMessage(ctx, data.Author.ID, &dto.MessageToCreate{
		Content: randomStr,
		MsgID:   data.ID,
	})
	if err != nil {
		log.Printf("Error sending C2C test reply: %v", err)
	}

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
	loadConfig()
	ctx = context.Background()

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
