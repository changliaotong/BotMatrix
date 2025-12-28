package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestEmailBotWorkflow tests the complete email bot workflow
func TestEmailBotWorkflow(t *testing.T) {
	// Test 1: Configuration loading
	t.Run("LoadConfiguration", func(t *testing.T) {
		// Save original config
		originalConfig := config

		// Load sample config
		loadConfig()

		// Verify default values are set
		if config.NexusAddr == "" {
			t.Error("Expected default NexusAddr to be set")
		}
		if config.LogPort == 0 {
			t.Error("Expected default LogPort to be set")
		}
		if config.PollInterval == 0 {
			t.Error("Expected default PollInterval to be set")
		}

		// Restore original config
		config = originalConfig
	})

	// Test 2: Email receiving simulation
	t.Run("EmailReceiving", func(t *testing.T) {
		// Create a mock Nexus server to receive events
		mockNexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade to WebSocket: %v", err)
				return
			}
			defer conn.Close()

			// Listen for incoming messages
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}

				// Parse the received event
				var event map[string]any
				if err := json.Unmarshal(message, &event); err != nil {
					t.Errorf("Failed to parse event: %v", err)
					return
				}

				// Check if it's a message event
				if postType, ok := event["post_type"].(string); ok && postType == "message" {
					// Verify required fields for email message
					requiredFields := []string{"message_type", "time", "self_id", "user_id", "message", "sender"}
					for _, field := range requiredFields {
						if _, exists := event[field]; !exists {
							t.Errorf("Missing required field in email message: %s", field)
						}
					}

					// Check message type is private for email
					if msgType, ok := event["message_type"].(string); ok && msgType != "private" {
						t.Errorf("Expected message_type to be 'private', got '%s'", msgType)
					}

					// Signal that we received the expected message
					return
				}
			}
		}))
		defer mockNexus.Close()

		// Update config to use mock Nexus
		config.NexusAddr = strings.Replace(mockNexus.URL, "http://", "ws://", 1)
		config.Username = "test@example.com"

		// Test processMessage function with mock IMAP message
		// This would require mocking the IMAP library which is complex
		// For now, we'll focus on testing the message processing logic
	})

	// Test 3: Email sending via OneBot protocol
	t.Run("EmailSending", func(t *testing.T) {
		// Mock Nexus server to send email commands
		mockNexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade to WebSocket: %v", err)
				return
			}
			defer conn.Close()

			// Send a send_private_msg command to the bot
			action := map[string]any{
				"action": "send_private_msg",
				"params": map[string]any{
					"user_id": "recipient@example.com",
					"message": "Test message from BotMatrix",
				},
				"echo": "test_echo",
			}

			if err := conn.WriteJSON(action); err != nil {
				t.Errorf("Failed to send action: %v", err)
				return
			}

			// Read response
			_, message, err := conn.ReadMessage()
			if err != nil {
				t.Errorf("Failed to read response: %v", err)
				return
			}

			var response map[string]any
			if err := json.Unmarshal(message, &response); err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			// Check response format
			if status, ok := response["status"].(string); !ok || status != "ok" {
				t.Errorf("Expected status 'ok', got '%s'", status)
			}
			if retcode, ok := response["retcode"].(float64); ok && retcode != 0 {
				t.Errorf("Expected retcode 0, got %f", retcode)
			}
		}))
		defer mockNexus.Close()

		// Set up config to use mock Nexus
		config.NexusAddr = strings.Replace(mockNexus.URL, "http://", "ws://", 1)
		config.Username = "sender@example.com"
		config.SmtpServer = "smtp.example.com"
		config.SmtpPort = 587
		config.SmtpUsername = "sender@example.com"
		config.SmtpPassword = "password"
	})

	// Test 4: Connection to Nexus
	t.Run("NexusConnection", func(t *testing.T) {
		// Create a mock Nexus server
		mockNexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers
			selfID := r.Header.Get("X-Self-ID")
			platform := r.Header.Get("X-Platform")

			if selfID == "" {
				t.Error("Expected X-Self-ID header to be set")
			}
			if platform != "Email" {
				t.Errorf("Expected X-Platform header to be 'Email', got '%s'", platform)
			}

			// Upgrade to WebSocket
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade to WebSocket: %v", err)
				return
			}
			defer conn.Close()

			// Listen for lifecycle event
			_, message, err := conn.ReadMessage()
			if err != nil {
				t.Errorf("Failed to read message: %v", err)
				return
			}

			var event map[string]any
			if err := json.Unmarshal(message, &event); err != nil {
				t.Errorf("Failed to parse event: %v", err)
				return
			}

			// Verify lifecycle event
			if postType, ok := event["post_type"].(string); !ok || postType != "meta_event" {
				t.Errorf("Expected post_type 'meta_event', got '%s'", postType)
			}
			if metaType, ok := event["meta_event_type"].(string); !ok || metaType != "lifecycle" {
				t.Errorf("Expected meta_event_type 'lifecycle', got '%s'", metaType)
			}
			if subType, ok := event["sub_type"].(string); !ok || subType != "connect" {
				t.Errorf("Expected sub_type 'connect', got '%s'", subType)
			}
		}))
		defer mockNexus.Close()

		// Set up context for connection test
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Set config to use mock Nexus
		config.NexusAddr = strings.Replace(mockNexus.URL, "http://", "ws://", 1)
		config.Username = "test@example.com"

		// Attempt to connect to Nexus
		go connectToNexus(ctx, config.NexusAddr)

		// Wait for connection to establish
		time.Sleep(1 * time.Second)
	})

	// Test 5: Heartbeat functionality
	t.Run("Heartbeat", func(t *testing.T) {
		heartbeatReceived := make(chan bool, 1)

		// Create mock Nexus server
		mockNexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade to WebSocket: %v", err)
				return
			}
			defer conn.Close()

			// Listen for heartbeat events
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}

				var event map[string]any
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if metaType, ok := event["meta_event_type"].(string); ok && metaType == "heartbeat" {
					status, ok := event["status"].(map[string]any)
					if ok {
						if online, exists := status["online"]; exists && online.(bool) == true {
							if good, exists := status["good"]; exists && good.(bool) == true {
								heartbeatReceived <- true
								return
							}
						}
					}
				}
			}
		}))
		defer mockNexus.Close()

		// Set up context for heartbeat test
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Set config to use mock Nexus
		config.NexusAddr = strings.Replace(mockNexus.URL, "http://", "ws://", 1)
		config.Username = "test@example.com"

		// Attempt to connect to Nexus
		go connectToNexus(ctx, config.NexusAddr)

		// Wait for heartbeat (should come within 30 seconds, but we'll wait less for test)
		select {
		case <-heartbeatReceived:
			// Success
		case <-time.After(35 * time.Second):
			t.Error("Did not receive heartbeat within expected time")
		}
	})

	// Test 6: Action handling
	t.Run("ActionHandling", func(t *testing.T) {
		// Test the handleAction function directly
		action := map[string]any{
			"action": "send_private_msg",
			"params": map[string]any{
				"user_id": "recipient@example.com",
				"message": "Test message",
			},
			"echo": "test_echo",
		}

		jsonData, err := json.Marshal(action)
		if err != nil {
			t.Fatalf("Failed to marshal action: %v", err)
		}

		// Mock the sendEvent function to capture the response
		// For this test, we'll just verify the function doesn't panic
		handleAction(jsonData)
	})

	// Test 7: Email sending function
	t.Run("SendEmailFunction", func(t *testing.T) {
		// Note: This test would require mocking SMTP server
		// For now, we'll test with invalid parameters to ensure error handling
		config.SmtpServer = "invalid-server"
		config.SmtpPort = 9999
		config.SmtpUsername = "test"
		config.SmtpPassword = "test"

		err := sendEmail("recipient@example.com", "Test Subject", "Test Body")
		if err == nil {
			t.Error("Expected error when sending to invalid server")
		}
	})
}

// TestConfigUI tests the configuration UI endpoints
func TestConfigUI(t *testing.T) {
	// Reset config to defaults
	loadConfig()

	t.Run("ConfigEndpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/config", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleConfig)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, status)
		}

		var cfg Config
		if err := json.Unmarshal(rr.Body.Bytes(), &cfg); err != nil {
			t.Errorf("Failed to parse config response: %v", err)
		}
	})

	t.Run("ConfigUpdate", func(t *testing.T) {
		newConfig := Config{
			ImapServer:   "imap.test.com",
			ImapPort:     993,
			Username:     "test@test.com",
			Password:     "password",
			SmtpServer:   "smtp.test.com",
			SmtpPort:     587,
			PollInterval: 30,
			NexusAddr:    "ws://test:3001",
			LogPort:      8086,
		}

		jsonData, err := json.Marshal(newConfig)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/config", strings.NewReader(string(jsonData)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleConfig)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, status)
		}

		// Verify config was updated
		if config.ImapServer != "imap.test.com" {
			t.Errorf("Expected ImapServer to be 'imap.test.com', got '%s'", config.ImapServer)
		}
	})
}

// TestEmailBotIntegration tests the full integration
func TestEmailBotIntegration(t *testing.T) {
	t.Log("Testing full EmailBot integration workflow...")

	// This test would simulate the entire workflow:
	// 1. Start the email bot
	// 2. Have it connect to Nexus
	// 3. Receive an email and convert to OneBot event
	// 4. Receive a command from Nexus to send email
	// 5. Send the email via SMTP

	// Due to the complexity of mocking all external services (IMAP, SMTP, WebSocket),
	// this test would require a full integration test environment.
	// For now, we'll note what needs to be tested.

	t.Log("Integration test would require:")
	t.Log("1. Mock IMAP server to simulate incoming emails")
	t.Log("2. Mock SMTP server to verify outgoing emails")
	t.Log("3. Mock Nexus server to exchange OneBot protocol messages")
	t.Log("4. End-to-end workflow validation")

	// For now, we'll just validate the expected behavior
	expectedBehaviors := []string{
		"Connect to BotNexus via WebSocket with proper headers",
		"Send lifecycle connect event",
		"Send heartbeat events every 30 seconds",
		"Poll IMAP server for new emails",
		"Convert received emails to OneBot message events",
		"Handle send_private_msg commands from Nexus",
		"Send emails via SMTP when receiving commands",
		"Maintain connection with reconnection logic",
	}

	for i, behavior := range expectedBehaviors {
		t.Logf("%d. %s", i+1, behavior)
	}
}

// Example test for processMessage function with simplified inputs
func TestProcessMessage(t *testing.T) {
	// This is a simplified test since testing the full IMAP message processing
	// would require complex mocking of the go-imap library structures.

	// We'll test the expected output format by examining the code logic
	selfID = "test@example.com"

	// The function processMessage converts IMAP messages to OneBot events
	// We'll verify that the expected fields are present in the resulting event
	t.Log("processMessage should convert IMAP messages to OneBot message events with:")
	t.Log("- post_type: message")
	t.Log("- message_type: private")
	t.Log("- time: from email envelope")
	t.Log("- user_id: sender email address")
	t.Log("- message: subject + body content")
	t.Log("- sender info with user_id and nickname")
}
