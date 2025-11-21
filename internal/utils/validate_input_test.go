package utils

import "testing"

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ok", " user_01 ", false},
		{"empty", " ", true},
		{"invalid chars", "user-name", true},
		{"too short", "a", true},
		{"too long", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", true},
	}

	for _, tt := range tests {
		_, err := IsValidUsername(tt.input)
		if tt.wantErr && err == nil {
			t.Fatalf("expected error for %s", tt.name)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("unexpected error for %s: %v", tt.name, err)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	valid, err := IsValidEmail(" user@example.com ")
	if err != nil || valid != "user@example.com" {
		t.Fatalf("expected trimmed valid email, got %q, err %v", valid, err)
	}

	if _, err := IsValidEmail("invalid"); err == nil {
		t.Fatal("expected invalid email to error")
	}
}

func TestIsValidPassword(t *testing.T) {
	if _, err := IsValidPassword(" Valid123! "); err != nil {
		t.Fatalf("expected strong password, got %v", err)
	}

	weakPasswords := []string{
		"short1!",
		"alllowercase1!",
		"ALLUPPERCASE1!",
		"NoDigits!!",
		"NoSpecial123",
	}

	for _, pwd := range weakPasswords {
		if _, err := IsValidPassword(pwd); err == nil {
			t.Fatalf("expected password %q to be rejected", pwd)
		}
	}
}

func TestIsValidTitleAndDescription(t *testing.T) {
	if _, err := IsValidTitle("  Article title  "); err != nil {
		t.Fatalf("expected valid title, got %v", err)
	}
	if _, err := IsValidDescription("  Summary  "); err != nil {
		t.Fatalf("expected valid description, got %v", err)
	}

	if _, err := IsValidTitle(" "); err == nil {
		t.Fatalf("expected empty title to fail")
	}

	long := make([]byte, 260)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := IsValidDescription(string(long)); err == nil {
		t.Fatalf("expected too long description to fail")
	}
}

func TestIsValidTagList(t *testing.T) {
	tags, err := IsValidTagList([]string{" go ", " web "})
	if err != nil {
		t.Fatalf("expected valid tags, got %v", err)
	}
	if tags[0] != "go" || tags[1] != "web" {
		t.Fatalf("expected trimmed tags, got %v", tags)
	}

	if _, err := IsValidTagList([]string{"", "go"}); err == nil {
		t.Fatalf("expected empty tag to fail")
	}
}
