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
	token, errGen := m.GenerateAccess("user-42")
	uid, errParse := m.Parse(token, TypeAccess)

	// Assert
	require.NoError(t, errGen)
	require.NoError(t, errParse)
	assert.Equal(t, "user-42", uid)
}

// メール確認トークンを発行して検証した時、元のユーザーIDと email が復元されること。
func TestGenerateAndParseEmailVerify(t *testing.T) {
	// Arrange
	m := NewManager("secret")

	// Act
	token, errGen := m.GenerateEmailVerify("user-7", "u7@example.com")
	uid, email, errParse := m.ParseEmailVerify(token)

	// Assert
	require.NoError(t, errGen)
	require.NoError(t, errParse)
	assert.Equal(t, "user-7", uid)
	assert.Equal(t, "u7@example.com", email)
}

// パスワードリセットトークンを発行して検証した時、元のユーザーIDが復元されること。
func TestGenerateAndParsePasswordReset(t *testing.T) {
	// Arrange
	m := NewManager("secret")

	// Act
	token, errGen := m.GeneratePasswordReset("user-9")
	uid, errParse := m.Parse(token, TypePasswordReset)

	// Assert
	require.NoError(t, errGen)
	require.NoError(t, errParse)
	assert.Equal(t, "user-9", uid)
}

// 確認トークンをリセットとして検証した時、token_type 不一致でエラーになること。
func TestParse_EmailVerifyAsPasswordReset(t *testing.T) {
	// Arrange
	m := NewManager("secret")
	verify, _ := m.GenerateEmailVerify("u1", "u1@example.com")

	// Act
	_, err := m.Parse(verify, TypePasswordReset)

	// Assert
	assert.Error(t, err)
}

// access トークンを refresh として検証した時、エラーになること。
func TestParse_AccessTokenAsRefresh(t *testing.T) {
	// Arrange
	m := NewManager("secret")
	access, _ := m.GenerateAccess("u1")

	// Act
	_, err := m.Parse(access, TypeRefresh)

	// Assert
	assert.Error(t, err)
}

// refresh トークンを access として検証した時、エラーになること。
func TestParse_RefreshTokenAsAccess(t *testing.T) {
	// Arrange
	m := NewManager("secret")
	refresh, _ := m.GenerateRefresh("u1")

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
	token, _ := signer.GenerateAccess("u1")

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
