package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 1. Login to get token
	loginURL := "http://127.0.0.1:8080/api/login"
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123", // Correct password from config.json
	}
	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var loginResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResult)

	if !loginResult.Success {
		fmt.Printf("Login failed: %s\n", loginResult.Message)
		return
	}
	token := loginResult.Data.Token
	fmt.Printf("Login successful, token: %s...\n", token[:10])

	// 2. Call chat stream API
	chatURL := "http://127.0.0.1:8080/api/ai/chat/stream"
	chatData := map[string]interface{}{
		"agent_id": 86, // 早喵
		"messages": []map[string]string{
			{"role": "user", "content": "你好"},
		},
	}
	jsonData, _ = json.Marshal(chatData)
	req, _ := http.NewRequest("POST", chatURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Chat request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Read streaming response
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("\nError reading stream: %v\n", err)
			break
		}
	}
}
