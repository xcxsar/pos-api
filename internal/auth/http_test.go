package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headerSet bool
		value     string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid bearer token",
			headerSet: true,
			value:     "Bearer abc.def.ghi",
			want:      "abc.def.ghi",
		},
		{
			name:      "missing authorization header",
			headerSet: false,
			wantErr:   true,
		},
		{
			name:      "empty authorization header",
			headerSet: true,
			value:     "",
			wantErr:   true,
		},
		{
			name:      "token without bearer prefix",
			headerSet: true,
			value:     "abc.def.ghi",
			wantErr:   true,
		},
		{
			name:      "wrong scheme",
			headerSet: true,
			value:     "Basic dXNlcjpwYXNz",
			wantErr:   true,
		},
		{
			name:      "lowercase bearer prefix rejected",
			headerSet: true,
			value:     "bearer abc.def.ghi",
			wantErr:   true,
		},
		{
			name:      "prefix without space rejected",
			headerSet: true,
			value:     "Bearerabc.def.ghi",
			wantErr:   true,
		},
		{
			name:      "empty token after prefix",
			headerSet: true,
			value:     "Bearer ",
			want:      "",
		},
		{
			name:      "leading space preserved in token",
			headerSet: true,
			value:     "Bearer  padded",
			want:      " padded",
		},
		{
			name:      "trailing whitespace preserved in token",
			headerSet: true,
			value:     "Bearer padded ",
			want:      "padded ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.headerSet {
				headers.Set("Authorization", tt.value)
			}

			got, err := GetBearerToken(headers)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got token %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got token %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	t.Run("returns user ID when present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, id)

		got, ok := GetUserIDFromContext(ctx)

		if !ok {
			t.Fatal("expected ok=true")
		}
		if got != id {
			t.Errorf("got %s, want %s", got, id)
		}
	})

	t.Run("returns false when absent", func(t *testing.T) {
		_, ok := GetUserIDFromContext(context.Background())

		if ok {
			t.Error("expected ok=false for empty context")
		}
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, id.String())

		got, ok := GetUserIDFromContext(ctx)

		if ok {
			t.Error("expected ok=false for string value under UserIDKey")
		}
		if got != uuid.Nil {
			t.Errorf("expected zero UUID, got %s", got)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	const secret = "test-secret"

	userID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	newRequest := func(authHeader string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if authHeader != "" {
			r.Header.Set("Authorization", authHeader)
		}
		return r
	}

	run := func(t *testing.T, r *http.Request) (*httptest.ResponseRecorder, *int, uuid.UUID) {
		var calls int
		var gotID uuid.UUID

		next := func(w http.ResponseWriter, req *http.Request) {
			calls++
			gotID, _ = GetUserIDFromContext(req.Context())
			w.WriteHeader(http.StatusOK)
		}

		handler := AuthMiddleware(secret, next)
		rec := httptest.NewRecorder()
		handler(rec, r)

		return rec, &calls, gotID
	}

	assertUnauthorized := func(t *testing.T, rec *httptest.ResponseRecorder, calls *int, wantMsg string) {
		t.Helper()

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != wantMsg {
			t.Errorf("got body %q, want %q", got, wantMsg)
		}
		if *calls != 0 {
			t.Errorf("next called %d times, want 0", *calls)
		}
	}

	t.Run("passes valid token through with user ID in context", func(t *testing.T) {
		token, err := MakeJWT(userID, secret, time.Minute)
		if err != nil {
			t.Fatalf("MakeJWT failed: %v", err)
		}

		rec, calls, gotID := run(t, newRequest("Bearer "+token))

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if *calls != 1 {
			t.Errorf("next called %d times, want 1", *calls)
		}
		if gotID != userID {
			t.Errorf("user ID in context is %s, want %s", gotID, userID)
		}
	})

	t.Run("rejects missing authorization header", func(t *testing.T) {
		rec, calls, _ := run(t, newRequest(""))

		assertUnauthorized(t, rec, calls, `{"error": "missing or invalid token"}`)
	})

	t.Run("rejects malformed authorization header", func(t *testing.T) {
		rec, calls, _ := run(t, newRequest("NotBearer "+userID.String()))

		assertUnauthorized(t, rec, calls, `{"error": "missing or invalid token"}`)
	})

	t.Run("rejects token signed with wrong secret", func(t *testing.T) {
		token, err := MakeJWT(userID, "other-secret", time.Minute)
		if err != nil {
			t.Fatalf("MakeJWT failed: %v", err)
		}

		rec, calls, _ := run(t, newRequest("Bearer "+token))

		assertUnauthorized(t, rec, calls, `{"error": "invalid or expired token"}`)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		token, err := MakeJWT(userID, secret, -time.Minute)
		if err != nil {
			t.Fatalf("MakeJWT failed: %v", err)
		}

		rec, calls, _ := run(t, newRequest("Bearer "+token))

		assertUnauthorized(t, rec, calls, `{"error": "invalid or expired token"}`)
	})

	t.Run("rejects token with unexpected subject", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Issuer:    string(TokenTypeAccess),
			Subject:   "not-a-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("signing token failed: %v", err)
		}

		rec, calls, _ := run(t, newRequest("Bearer "+signed))

		assertUnauthorized(t, rec, calls, `{"error": "invalid or expired token"}`)
	})

	t.Run("rejects token from untrusted issuer", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Issuer:    "other-issuer",
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("signing token failed: %v", err)
		}

		rec, calls, _ := run(t, newRequest("Bearer "+signed))

		assertUnauthorized(t, rec, calls, `{"error": "invalid or expired token"}`)
	})
}
