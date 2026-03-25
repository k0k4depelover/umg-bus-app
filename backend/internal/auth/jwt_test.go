package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-jwt"

func newTestJWT() *JWTService {
	return NewJWTService(testSecret)
}

func TestGenerateAccess_Success(t *testing.T) {
	svc := newTestJWT()
	token, err := svc.GenerateAccess("user-123", "campus-456", "pilot")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID=user-123, got %s", claims.UserID)
	}
	if claims.CampusID != "campus-456" {
		t.Errorf("expected CampusID=campus-456, got %s", claims.CampusID)
	}
	if claims.Role != "pilot" {
		t.Errorf("expected Role=pilot, got %s", claims.Role)
	}
}

func TestGenerateAccess_Expiry(t *testing.T) {
	svc := newTestJWT()
	token, err := svc.GenerateAccess("user-1", "campus-1", "student")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expiry := claims.ExpiresAt.Time
	expected := time.Now().Add(15 * time.Minute)
	diff := expiry.Sub(expected)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry should be ~15 min from now, got diff=%v", diff)
	}
}

func TestGenerateRefresh_Success(t *testing.T) {
	svc := newTestJWT()
	token, err := svc.GenerateRefresh("user-123", "student")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID=user-123, got %s", claims.UserID)
	}
	if claims.CampusID != "" {
		t.Errorf("expected empty CampusID for refresh token, got %s", claims.CampusID)
	}
	if claims.Role != "student" {
		t.Errorf("expected Role=student, got %s", claims.Role)
	}
}

func TestGenerateRefresh_Expiry(t *testing.T) {
	svc := newTestJWT()
	token, err := svc.GenerateRefresh("user-1", "pilot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expiry := claims.ExpiresAt.Time
	expected := time.Now().Add(7 * 24 * time.Hour)
	diff := expiry.Sub(expected)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry should be ~7 days from now, got diff=%v", diff)
	}
}

func TestVerify_ValidToken(t *testing.T) {
	svc := newTestJWT()
	token, _ := svc.GenerateAccess("u1", "c1", "pilot")
	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.UserID != "u1" || claims.CampusID != "c1" || claims.Role != "pilot" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	svc := newTestJWT()

	claims := Claims{
		UserID:   "user-1",
		CampusID: "campus-1",
		Role:     "pilot",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = svc.Verify(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	svcA := NewJWTService("secret-A")
	svcB := NewJWTService("secret-B")

	token, _ := svcA.GenerateAccess("u1", "c1", "pilot")
	_, err := svcB.Verify(token)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret")
	}
}

func TestVerify_MalformedToken(t *testing.T) {
	svc := newTestJWT()
	_, err := svc.Verify("this.is.not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestVerify_EmptyToken(t *testing.T) {
	svc := newTestJWT()
	_, err := svc.Verify("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
