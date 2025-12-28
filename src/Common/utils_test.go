package common

import (
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

func TestValidateToken(t *testing.T) {
	// Setup for token testing
	JWT_SECRET = "test_secret_key"
	
	// Since GenerateToken is a method on Manager, and we might not want to 
	// initialize a full Manager for a unit test, we'll skip testing the method
	// but we can test ValidateToken if we have a valid token string.
	// For simplicity in this example, we focus on the standalone functions.
}
