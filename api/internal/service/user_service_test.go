package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 該当ユーザーがいる時、GetByID でそのユーザーが構造体ごと返ること。
func TestUserGetByID_Found(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID(7), factory.WithUsername("alice"))
	ur := &mockUserRepo{byID: map[uint]*domain.ApplicationUser{7: user}}
	svc := NewUserService(ur)

	// Act
	got, err := svc.GetByID(context.Background(), 7)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

// 該当ユーザーがいない時、GetByID で nil が返ること。
func TestUserGetByID_NotFound(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{byID: map[uint]*domain.ApplicationUser{}}
	svc := NewUserService(ur)

	// Act
	got, err := svc.GetByID(context.Background(), 999)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got)
}

// ユーザーが登録されている時、List で全件が返ること。
func TestUserList_ReturnsAll(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{all: []domain.ApplicationUser{
		*factory.NewUser(factory.WithID(1), factory.WithUsername("alice")),
		*factory.NewUser(factory.WithID(2), factory.WithUsername("bob")),
	}}
	svc := NewUserService(ur)

	// Act
	users, err := svc.List(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// ユーザーが1人もいない時、List で空が返ること。
func TestUserList_Empty(t *testing.T) {
	// Arrange
	svc := NewUserService(&mockUserRepo{})

	// Act
	users, err := svc.List(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Empty(t, users)
}
