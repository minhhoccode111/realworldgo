package utils

import "testing"

func TestHashedPasswordAndValidate(t *testing.T) {
	raw := "S3curePwd!"
	hashed, err := HashedPassword(raw)
	if err != nil {
		t.Fatalf("hashing password failed: %v", err)
	}

	if hashed == raw {
		t.Fatalf("hashed password should differ from raw input")
	}

	if !ValidatePassword(hashed, raw) {
		t.Fatalf("expected password validation to pass")
	}

	if ValidatePassword(hashed, "wrong") {
		t.Fatalf("expected password validation to fail for wrong password")
	}
}
