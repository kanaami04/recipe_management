package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// アクセストークンを発行して検証した時、元のユーザーIDが復元されること。
func TestGenerateAndParseAccess(t *testing.T) {
	// Arrange
	m := NewManager("secret")

	// Act
	token, errGen := m.GenerateAccess(42)
	uid, errParse := m.Parse(token, TypeAccess)

	// Assert
	require.NoError(t, errGen)
	require.NoError(t, errParse)
	assert.Equal(t, uint(42), uid)
}

// access トークンを refresh として検証した時、エラーになること。
func TestParse_AccessTokenAsRefresh(t *testing.T) {
	// Arrange
	m := NewManager("secret")
	access, _ := m.GenerateAccess(1)

	// Act
	_, err := m.Parse(access, TypeRefresh)

	// Assert
	assert.Error(t, err)
}

// refresh トークンを access として検証した時、エラーになること。
func TestParse_RefreshTokenAsAccess(t *testing.T) {
	// Arrange
	m := NewManager("secret")
	refresh, _ := m.GenerateRefresh(1)

	// Act
	_, err := m.Parse(refresh, TypeAccess)

	// Assert
	assert.Error(t, err)
}

// 署名鍵が異なる Manager で検証した時、エラーになること。
func TestParse_WrongSecret(t *testing.T) {
	// Arrange
	signer := NewManager("secret-a")
	verifier := NewManager("secret-b")
	token, _ := signer.GenerateAccess(1)

	// Act
	_, err := verifier.Parse(token, TypeAccess)

	// Assert
	assert.Error(t, err)
}

// JWT として不正な文字列を検証した時、エラーになること。
func TestParse_Garbage(t *testing.T) {
	// Arrange
	m := NewManager("secret")

	// Act
	_, err := m.Parse("not-a-jwt", TypeAccess)

	// Assert
	assert.Error(t, err)
}
