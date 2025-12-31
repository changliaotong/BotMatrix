package utils

import (
	"encoding/json"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "my_secret_password"
	
	// Test HashPassword
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	if hash == password {
		t.Error("HashPassword returned the original password, it should be hashed")
	}
	
	// Test CheckPassword with correct password
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword failed with correct password")
	}
	
	// Test CheckPassword with wrong password
	if CheckPassword("wrong_password", hash) {
		t.Error("CheckPassword should have failed with wrong password")
	}
}

func TestGenerateRandomToken(t *testing.T) {
	length := 16
	token1 := GenerateRandomToken(length)
	token2 := GenerateRandomToken(length)
	
	if len(token1) != length*2 { // hex encoding doubles the length
		t.Errorf("Expected token length %d, got %d", length*2, len(token1))
	}
	
	if token1 == token2 {
		t.Error("GenerateRandomToken produced identical tokens")
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"int64", int64(1234567890), "1234567890"},
		{"float64 small", 12.34, "12.34"},
		{"float64 large", 2958935140.0, "2958935140"},
		{"json.Number int", json.Number("2958935140"), "2958935140"},
		{"json.Number float", json.Number("12.34"), "12.34"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.expected {
				t.Errorf("ToString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
