package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// EventMessage represents an event received from the core
type EventMessage struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

// Action represents an action to be performed
type Action struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	TargetID string `json:"target_id"`
	Text     string `json:"text"`
}

// ResponseMessage represents a response to the core
type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	for {
		var msg EventMessage
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error decoding message: %v\n", err)
			continue
		}

		if msg.Type == "event" && msg.Name == "on_message" {
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				fmt.Fprintf(os.Stderr, "Invalid payload type\n")
				continue
			}

			text, textOk := payload["text"].(string)
			target, targetOk := payload["from"].(string)
			targetID, targetIDOk := payload["group_id"].(string)

			if !textOk || !targetOk || !targetIDOk {
				fmt.Fprintf(os.Stderr, "Missing required fields in payload\n")
				continue
			}

			// TODO: Add plugin logic here
			response := ResponseMessage{
				ID: msg.ID,
				OK: true,
				Actions: []Action{
					{
						Type:     "send_message",
						Target:   target,
						TargetID: targetID,
						Text:     text,
					},
				},
			}

			if err := encoder.Encode(response); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
				continue
			}
		}
	}
}
