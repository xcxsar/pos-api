package password

import (
	"testing"
)

func TestCheck(t *testing.T) {
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
			got := Check(tt.password)
			if got != tt.want {
				t.Errorf("Check(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestHashAndMatch(t *testing.T) {
	pwd := "TestPass1!"

	hash, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() returned unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty string")
	}
	if hash == pwd {
		t.Fatal("Hash() returned the plaintext password")
	}

	match, err := Match(pwd, hash)
	if err != nil {
		t.Fatalf("Match() returned unexpected error: %v", err)
	}
	if !match {
		t.Error("Match() = false for correct password")
	}

	match, err = Match("WrongPass1!", hash)
	if err != nil {
		t.Fatalf("Match() returned unexpected error for wrong password: %v", err)
	}
	if match {
		t.Error("Match() = true for incorrect password")
	}
}

func TestMatchInvalidHash(t *testing.T) {
	_, err := Match("password", "not-a-valid-hash")
	if err == nil {
		t.Error("Match() expected error for invalid hash, got nil")
	}
}

func TestHashUniqueness(t *testing.T) {
	pwd := "UniqueHash1!"

	hash1, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	hash2, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash() produced identical hashes for same input (expected unique due to random salt)")
	}
}

func TestEmptyPasswordHashAndMatch(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("Hash(\"\") returned unexpected error: %v", err)
	}

	match, err := Match("", hash)
	if err != nil {
		t.Fatalf("Match(\"\", hash) returned unexpected error: %v", err)
	}
	if !match {
		t.Error("Match() = false for empty password matching its hash")
	}
}

func TestPasswordWithSpaces(t *testing.T) {
	pwd := "Passw0rd @"
	got := Check(pwd)
	if !got {
		t.Errorf("Check(%q) = false, want true", pwd)
	}

	hash, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	match, err := Match(pwd, hash)
	if err != nil {
		t.Fatalf("Match() error: %v", err)
	}
	if !match {
		t.Error("Match() = false for password with space + special")
	}
}

func TestUnicodePassword(t *testing.T) {
	pwd := "Pässw0rd!"
	got := Check(pwd)
	if !got {
		t.Errorf("Check(%q) = false, want true (ä satisfies hasLower via unicode.IsLower)", pwd)
	}

	hash, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	match, err := Match(pwd, hash)
	if err != nil {
		t.Fatalf("Match() error: %v", err)
	}
	if !match {
		t.Error("Match() = false for unicode password matching its hash")
	}
}
