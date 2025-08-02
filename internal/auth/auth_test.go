package auth

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTokenValidity(t *testing.T) {
	// Basic case
	tokenSecret := "testsecret"
	expirationDuration := time.Duration(3600 * math.Pow(1000, 3))
	userUuid := uuid.New()

	token, err := MakeJWT(userUuid, tokenSecret, expirationDuration)

	if err != nil {
		t.Errorf("Failed to create token: %v", err)
		return
	}

	returnUUID, err := ValidateJWT(token, tokenSecret)

	if err != nil {
		t.Errorf("Failed to validate JWT: %v", err)
		return
	}
	
	if userUuid != returnUUID {
		t.Errorf("User %v does not correspond to user %v", userUuid, returnUUID)
		return
	}

	// Test validation if wrong secret provided
	_, err = ValidateJWT(token, "wrong secret")

	if err == nil {
		t.Error("Expired token are not rejected")
		return
	}

	// Test case if token is expired
	expirationDateInZero := time.Duration(-1000)

	token, err = MakeJWT(userUuid, tokenSecret, expirationDateInZero)

	if err != nil {
		t.Errorf("Failed to create jwt token: %v", err)
		return
	}

	_, err = ValidateJWT(token, tokenSecret)	

	if err == nil {
		t.Errorf("Failed to catch expired tokens")
	}
}


