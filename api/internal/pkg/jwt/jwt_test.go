package jwt

import "testing"

func TestGenerateAndParseAccess(t *testing.T) {
	m := NewManager("secret")

	token, err := m.GenerateAccess(42)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	uid, err := m.Parse(token, TypeAccess)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if uid != 42 {
		t.Errorf("uid = %d, want 42", uid)
	}
}

func TestParse_WrongTokenType(t *testing.T) {
	m := NewManager("secret")

	// access トークンを refresh として検証 → 失敗するべき
	access, _ := m.GenerateAccess(1)
	if _, err := m.Parse(access, TypeRefresh); err == nil {
		t.Error("expected error for wrong token type, got nil")
	}

	// refresh を access として検証 → 失敗
	refresh, _ := m.GenerateRefresh(1)
	if _, err := m.Parse(refresh, TypeAccess); err == nil {
		t.Error("expected error for wrong token type, got nil")
	}
}

func TestParse_WrongSecret(t *testing.T) {
	signer := NewManager("secret-a")
	verifier := NewManager("secret-b")

	token, _ := signer.GenerateAccess(1)
	if _, err := verifier.Parse(token, TypeAccess); err == nil {
		t.Error("expected error for mismatched secret, got nil")
	}
}

func TestParse_Garbage(t *testing.T) {
	m := NewManager("secret")
	if _, err := m.Parse("not-a-jwt", TypeAccess); err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}
