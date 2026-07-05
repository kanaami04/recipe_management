package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// グループ未所属のユーザーが作成した時、自分が所有者兼メンバーのグループができること。
func TestShareGroupCreate_Succeeds(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	svc := NewShareGroupService(gr)

	// Act
	group, err := svc.Create(context.Background(), "u1", "我が家")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u1", group.OwnerID)
	assert.Equal(t, "我が家", group.Name)
	require.Len(t, group.Members, 1)
	assert.Equal(t, "u1", group.Members[0].ID)
}

// 名前を空で作成した時、既定名が使われること。
func TestShareGroupCreate_DefaultsName(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo())

	// Act
	group, err := svc.Create(context.Background(), "u1", "  ")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, defaultGroupName, group.Name)
}

// 既にグループに所属しているユーザーが作成しようとした時、ErrAlreadyInGroup が返ること。
func TestShareGroupCreate_AlreadyInGroup(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "u1")
	svc := NewShareGroupService(gr)

	// Act
	_, err := svc.Create(context.Background(), "u1", "別グループ")

	// Assert
	assert.ErrorIs(t, err, ErrAlreadyInGroup)
}

// 有効な招待コードで参加した時、メンバーに加わること。
func TestShareGroupJoin_Succeeds(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	g := gr.seed("g1", "owner")
	g.InviteCode = "CODE1234"
	g.InviteCodeExpiresAt = time.Now().Add(time.Hour)
	svc := NewShareGroupService(gr)

	// Act
	group, err := svc.Join(context.Background(), "u2", "CODE1234")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "g1", group.ID)
	ids, _ := gr.MemberIDs(context.Background(), "u2")
	assert.Contains(t, ids, "u2")
}

// 既にグループに所属しているユーザーが参加しようとした時、ErrAlreadyInGroup が返ること。
func TestShareGroupJoin_AlreadyInGroup(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	other := gr.seed("g1", "owner")
	other.InviteCode = "CODE1234"
	other.InviteCodeExpiresAt = time.Now().Add(time.Hour)
	gr.seed("g2", "u2") // u2 は既に別グループ
	svc := NewShareGroupService(gr)

	// Act
	_, err := svc.Join(context.Background(), "u2", "CODE1234")

	// Assert
	assert.ErrorIs(t, err, ErrAlreadyInGroup)
}

// 存在しない招待コードで参加しようとした時、ErrInviteCodeInvalid が返ること。
func TestShareGroupJoin_InvalidCode(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "NOPE")

	// Assert
	assert.ErrorIs(t, err, ErrInviteCodeInvalid)
}

// 期限切れの招待コードで参加しようとした時、ErrInviteCodeInvalid が返ること。
func TestShareGroupJoin_ExpiredCode(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	g := gr.seed("g1", "owner")
	g.InviteCode = "OLDCODE1"
	g.InviteCodeExpiresAt = time.Now().Add(-time.Hour) // 期限切れ
	svc := NewShareGroupService(gr)

	// Act
	_, err := svc.Join(context.Background(), "u2", "OLDCODE1")

	// Assert
	assert.ErrorIs(t, err, ErrInviteCodeInvalid)
}

// メンバーが抜けた時、そのメンバーだけがグループから外れること(グループは残る)。
func TestShareGroupLeave_Member(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	svc := NewShareGroupService(gr)

	// Act
	err := svc.Leave(context.Background(), "u2")

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, gr.deletedIDs, "g1") // 解散していない
	group, _ := gr.FindByUserID(context.Background(), "u2")
	assert.Nil(t, group) // u2 は抜けた
}

// 所有者が抜けた時、グループが解散されること。
func TestShareGroupLeave_OwnerDisbands(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	svc := NewShareGroupService(gr)

	// Act
	err := svc.Leave(context.Background(), "owner")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, gr.deletedIDs, "g1")
}

// グループ未所属で抜けようとした時、ErrNotInGroup が返ること。
func TestShareGroupLeave_NotInGroup(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo())

	// Act
	err := svc.Leave(context.Background(), "u1")

	// Assert
	assert.ErrorIs(t, err, ErrNotInGroup)
}

// 所有者がメンバーを外した時、そのメンバーが外れること。
func TestShareGroupRemoveMember_OwnerRemoves(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	svc := NewShareGroupService(gr)

	// Act
	err := svc.RemoveMember(context.Background(), "owner", "u2")

	// Assert
	require.NoError(t, err)
	group, _ := gr.FindByUserID(context.Background(), "u2")
	assert.Nil(t, group)
}

// 所有者でないメンバーが他人を外そうとした時、ErrNotGroupOwner が返ること。
func TestShareGroupRemoveMember_NotOwner(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2", "u3")
	svc := NewShareGroupService(gr)

	// Act: u2(非所有者)が u3 を外そうとする
	err := svc.RemoveMember(context.Background(), "u2", "u3")

	// Assert
	assert.ErrorIs(t, err, ErrNotGroupOwner)
}

// 所有者が自分自身を外そうとした時、ErrForbidden が返ること(抜けるなら Leave)。
func TestShareGroupRemoveMember_CannotRemoveSelf(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	svc := NewShareGroupService(gr)

	// Act
	err := svc.RemoveMember(context.Background(), "owner", "owner")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 所有者が招待コードを再発行した時、コードが変わること。
func TestShareGroupRegenerateInviteCode_ChangesCode(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	g := gr.seed("g1", "owner")
	g.InviteCode = "OLDCODE1"
	svc := NewShareGroupService(gr)

	// Act
	group, err := svc.RegenerateInviteCode(context.Background(), "owner")

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, "OLDCODE1", group.InviteCode)
}

// 所有者でないメンバーが再発行しようとした時、ErrNotGroupOwner が返ること。
func TestShareGroupRegenerateInviteCode_NotOwner(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	svc := NewShareGroupService(gr)

	// Act
	_, err := svc.RegenerateInviteCode(context.Background(), "u2")

	// Assert
	assert.ErrorIs(t, err, ErrNotGroupOwner)
}
