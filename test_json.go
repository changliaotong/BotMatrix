package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	jsonStr := `{"user_id": 123456789012345678}`
	var msg map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &msg)

	userID := msg["user_id"]
	fmt.Printf("Type: %T, Value: %v\n", userID, userID)

	uID := fmt.Sprintf("user_%v", userID)
	fmt.Printf("uID: %s\n", uID)

	// Check scientific notation
	jsonStr2 := `{"user_id": 1234567890123456789}`
	json.Unmarshal([]byte(jsonStr2), &msg)
	userID2 := msg["user_id"]
	fmt.Printf("Type2: %T, Value2: %v\n", userID2, userID2)
}
