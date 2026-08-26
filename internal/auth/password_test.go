package auth

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid password", "Abcdef1!", true},
		{"valid with multiple specials", "Passw0rd@#$", true},
		{"empty string", "", false},
		{"too short", "Ab1!", false},
		{"exactly 7 chars", "Abcde1!", false},
		{"no lowercase", "ABCDEF1!", false},
		{"no uppercase", "abcdef1!", false},
		{"no digit", "Abcdefg!", false},
		{"no special", "Abcdefg1", false},
		{"only digits", "12345678", false},
		{"only letters lower", "abcdefgh", false},
		{"only letters upper", "ABCDEFGH", false},
		{"numbers and lowercase only", "abcd1234", false},
		{"numbers and uppercase only", "ABCD1234", false},
		{"lowercase and uppercase only", "abcdABCD", false},
		{"special characters not in set", "Abcdef1#", false},
		{"unicode special rejected", "Abcdef1£", false},
		{"whitespace counts as no category", "         ", false},
		{"valid long password", "VeryLongPassword1234@ThisIsGreat", true},
		{"special only from set", "Abcde1?*&", true},
		{"multiple category chars", "aA1@bB2#", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPassword(tt.password)
			if got != tt.want {
				t.Errorf("CheckPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestHashAndMatchPassword(t *testing.T) {
	password := "TestPass1!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty string")
	}
	if hash == password {
		t.Fatal("HashPassword() returned the plaintext password")
	}

	match, err := MatchPassword(password, hash)
	if err != nil {
		t.Fatalf("MatchPassword() returned unexpected error: %v", err)
	}
	if !match {
		t.Error("MatchPassword() = false for correct password")
	}

	match, err = MatchPassword("WrongPass1!", hash)
	if err != nil {
		t.Fatalf("MatchPassword() returned unexpected error for wrong password: %v", err)
	}
	if match {
		t.Error("MatchPassword() = true for incorrect password")
	}
}

func TestMatchPasswordInvalidHash(t *testing.T) {
	_, err := MatchPassword("password", "not-a-valid-hash")
	if err == nil {
		t.Error("MatchPassword() expected error for invalid hash, got nil")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "UniqueHash1!"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for same input (expected unique due to random salt)")
	}
}

func TestEmptyPasswordHashAndMatch(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword(\"\") returned unexpected error: %v", err)
	}

	match, err := MatchPassword("", hash)
	if err != nil {
		t.Fatalf("MatchPassword(\"\", hash) returned unexpected error: %v", err)
	}
	if !match {
		t.Error("MatchPassword() = false for empty password matching its hash")
	}
}

func TestPasswordWithSpaces(t *testing.T) {
	password := "Passw0rd @"
	got := CheckPassword(password)
	if !got {
		t.Errorf("CheckPassword(%q) = false, want true", password)
	}

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	match, err := MatchPassword(password, hash)
	if err != nil {
		t.Fatalf("MatchPassword() error: %v", err)
	}
	if !match {
		t.Error("MatchPassword() = false for password with space + special")
	}
}

func TestUnicodePassword(t *testing.T) {
	password := "Pässw0rd!"
	got := CheckPassword(password)
	if !got {
		t.Errorf("CheckPassword(%q) = false, want true (ä satisfies hasLower via unicode.IsLower)", password)
	}

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	match, err := MatchPassword(password, hash)
	if err != nil {
		t.Fatalf("MatchPassword() error: %v", err)
	}
	if !match {
		t.Error("MatchPassword() = false for unicode password matching its hash")
	}
}
