package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const testSecret = "unit-test-secret"

func signHS256(t *testing.T, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return signed
}

func accessClaims(subject string, expiresIn time.Duration) jwt.RegisteredClaims {
	now := time.Now().UTC()

	return jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
}

func TestMakeAndValidateJWTRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		secret    string
		expiresIn time.Duration
	}{
		{
			name:      "standard uuid, secret, and expiry",
			userID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			secret:    testSecret,
			expiresIn: time.Hour,
		},
		{
			name:      "zero uuid",
			userID:    uuid.Nil,
			secret:    testSecret,
			expiresIn: time.Hour,
		},
		{
			name:      "long expiry",
			userID:    uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			secret:    testSecret,
			expiresIn: 24 * 365 * time.Hour,
		},
		{
			name:      "empty secret",
			userID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			secret:    "",
			expiresIn: time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userID, tt.secret, tt.expiresIn)
			if err != nil {
				t.Fatalf("MakeJWT failed: %v", err)
			}
			if token == "" {
				t.Fatal("expected non-empty token string")
			}

			got, err := ValidateJWT(token, tt.secret)
			if err != nil {
				t.Fatalf("ValidateJWT failed: %v", err)
			}
			if got != tt.userID {
				t.Errorf("got user ID %s, want %s", got, tt.userID)
			}
		})
	}
}

func TestMakeJWTClaims(t *testing.T) {
	userID := uuid.MustParse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	const expiresIn = time.Hour

	token, err := MakeJWT(userID, testSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse freshly minted token: %v", err)
	}

	if parsed.Header["alg"] != jwt.SigningMethodHS256.Alg() {
		t.Errorf("alg header is %v, want %q", parsed.Header["alg"], jwt.SigningMethodHS256.Alg())
	}

	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("claims are not *jwt.RegisteredClaims")
	}

	if claims.Issuer != string(TokenTypeAccess) {
		t.Errorf("issuer is %q, want %q", claims.Issuer, TokenTypeAccess)
	}
	if claims.Subject != userID.String() {
		t.Errorf("subject is %q, want %q", claims.Subject, userID)
	}

	expDiff := time.Until(claims.ExpiresAt.Time) - expiresIn
	if expDiff > 10*time.Second || expDiff < -10*time.Second {
		t.Errorf("expiry deviates from now+%s by %s", expiresIn, expDiff)
	}

	iatSkew := time.Since(claims.IssuedAt.Time)
	if iatSkew > 10*time.Second || iatSkew < -10*time.Second {
		t.Errorf("issued-at deviates from now by %s", iatSkew)
	}
}

func TestValidateJWTErrors(t *testing.T) {
	userID := uuid.MustParse("99999999-8888-7777-6666-555555555555")

	validToken, err := MakeJWT(userID, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	tests := []struct {
		name            string
		token           func(t *testing.T) string
		wantErrContains string
	}{
		{
			name:  "garbage string",
			token: func(t *testing.T) string { return "not-a-jwt" },
		},
		{
			name:  "empty string",
			token: func(t *testing.T) string { return "" },
		},
		{
			name:            "too few segments",
			token:           func(t *testing.T) string { return "aaa.bbb" },
			wantErrContains: "segments",
		},
		{
			name:            "too many segments",
			token:           func(t *testing.T) string { return validToken + ".extra" },
			wantErrContains: "segments",
		},
		{
			name: "tampered signature",
			token: func(t *testing.T) string {
				parts := strings.Split(validToken, ".")
				sig := []byte(parts[2])
				i := len(sig) / 2
				if sig[i] == 'z' {
					sig[i] = 'y'
				} else {
					sig[i] = 'z'
				}
				parts[2] = string(sig)
				return strings.Join(parts, ".")
			},
			wantErrContains: "signature",
		},
		{
			name: "wrong secret",
			token: func(t *testing.T) string {
				signed, err := MakeJWT(userID, "another-secret", time.Hour)
				if err != nil {
					t.Fatalf("MakeJWT failed: %v", err)
				}
				return signed
			},
			wantErrContains: "signature",
		},
		{
			name: "expired token",
			token: func(t *testing.T) string {
				signed, err := MakeJWT(userID, testSecret, -time.Minute)
				if err != nil {
					t.Fatalf("MakeJWT failed: %v", err)
				}
				return signed
			},
			wantErrContains: "expired",
		},
		{
			name: "wrong issuer",
			token: func(t *testing.T) string {
				claims := accessClaims(userID.String(), time.Hour)
				claims.Issuer = "other-issuer"
				return signHS256(t, testSecret, claims)
			},
			wantErrContains: "invalid issuer",
		},
		{
			name: "missing issuer",
			token: func(t *testing.T) string {
				claims := accessClaims(userID.String(), time.Hour)
				claims.Issuer = ""
				return signHS256(t, testSecret, claims)
			},
			wantErrContains: "invalid issuer",
		},
		{
			name: "non-UUID subject",
			token: func(t *testing.T) string {
				return signHS256(t, testSecret, accessClaims("not-a-uuid", time.Hour))
			},
		},
		{
			name: "empty subject",
			token: func(t *testing.T) string {
				return signHS256(t, testSecret, accessClaims("", time.Hour))
			},
		},
		{
			name: "none algorithm rejected",
			token: func(t *testing.T) string {
				signed, err := jwt.NewWithClaims(jwt.SigningMethodNone, accessClaims(userID.String(), time.Hour)).
					SignedString(jwt.UnsafeAllowNoneSignatureType)
				if err != nil {
					t.Fatalf("failed to build none-token: %v", err)
				}
				return signed
			},
			wantErrContains: "unexpected signing method",
		},
		{
			name: "non-HMAC algorithm rejected",
			token: func(t *testing.T) string {
				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					t.Fatalf("failed to generate ECDSA key: %v", err)
				}
				signed, err := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims(userID.String(), time.Hour)).
					SignedString(key)
				if err != nil {
					t.Fatalf("failed to sign ES256 token: %v", err)
				}
				return signed
			},
			wantErrContains: "unexpected signing method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateJWT(tt.token(t), testSecret)

			if err == nil {
				t.Fatalf("expected error, got user ID %s", got)
			}
			if got != uuid.Nil {
				t.Errorf("expected zero UUID on failure, got %s", got)
			}
			if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestValidateJWTAcceptsOtherHMACAlgorithms(t *testing.T) {
	userID := uuid.MustParse("12121212-3434-5656-7878-909090909090")

	claims := accessClaims(userID.String(), time.Hour)
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign HS512 token: %v", err)
	}

	got, err := ValidateJWT(signed, testSecret)
	if err != nil {
		t.Fatalf("expected HS512 token to validate, got error: %v", err)
	}
	if got != userID {
		t.Errorf("got user ID %s, want %s", got, userID)
	}
}

func TestMakeRefreshToken(t *testing.T) {
	t.Run("returns 64 lowercase hex characters", func(t *testing.T) {
		token := MakeRefreshToken()

		if len(token) != 64 {
			t.Fatalf("got token of length %d, want 64", len(token))
		}

		for _, c := range token {
			if !strings.ContainsRune("0123456789abcdef", c) {
				t.Fatalf("non-hex character %q in token %q", c, token)
			}
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		seen := make(map[string]struct{})

		for i := 0; i < 100; i++ {
			seen[MakeRefreshToken()] = struct{}{}
		}

		if len(seen) != 100 {
			t.Errorf("generated %d unique tokens out of 100", len(seen))
		}
	})
}
