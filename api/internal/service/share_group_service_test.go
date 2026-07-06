package service

import (
	"context"
	"testing"
	"time"

	"recipe-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// グループ未所属のユーザーが作成した時、自分が所有者兼メンバーのグループができること。
func TestShareGroupCreate_Succeeds(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(newMockShareGroupRepo(), newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	group, err := svc.Join(context.Background(), "u2", "CODE1234", true)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "g1", group.ID)
	ids, _ := gr.MemberIDs(context.Background(), "u2")
	assert.Contains(t, ids, "u2")
}

// 統合を選んで参加した時、自分の個人リストが物理削除されること。
func TestShareGroupJoin_SharesShoppingList_DeletesPersonalList(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	g := gr.seed("g1", "owner")
	g.InviteCode = "CODE1234"
	g.InviteCodeExpiresAt = time.Now().Add(time.Hour)
	lr := newMockShoppingListRepo()
	require.NoError(t, lr.Create(context.Background(), &domain.ShoppingList{OwnerID: "u2"}))
	svc := NewShareGroupService(gr, lr, newMockRecipeRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "CODE1234", true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, lr.deletedOwnerIDs, "u2")
}

// 個人運用を選んで参加した時、自分の個人リストは消されないこと。
func TestShareGroupJoin_KeepsPersonalList(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	g := gr.seed("g1", "owner")
	g.InviteCode = "CODE1234"
	g.InviteCodeExpiresAt = time.Now().Add(time.Hour)
	lr := newMockShoppingListRepo()
	require.NoError(t, lr.Create(context.Background(), &domain.ShoppingList{OwnerID: "u2"}))
	svc := NewShareGroupService(gr, lr, newMockRecipeRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "CODE1234", false)

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, lr.deletedOwnerIDs, "u2")
}

// 既にグループに所属しているユーザーが参加しようとした時、ErrAlreadyInGroup が返ること。
func TestShareGroupJoin_AlreadyInGroup(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	other := gr.seed("g1", "owner")
	other.InviteCode = "CODE1234"
	other.InviteCodeExpiresAt = time.Now().Add(time.Hour)
	gr.seed("g2", "u2") // u2 は既に別グループ
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "CODE1234", true)

	// Assert
	assert.ErrorIs(t, err, ErrAlreadyInGroup)
}

// 存在しない招待コードで参加しようとした時、ErrInviteCodeInvalid が返ること。
func TestShareGroupJoin_InvalidCode(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo(), newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "NOPE", true)

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	_, err := svc.Join(context.Background(), "u2", "OLDCODE1", true)

	// Assert
	assert.ErrorIs(t, err, ErrInviteCodeInvalid)
}

// メンバーが抜けた時、そのメンバーだけがグループから外れること(グループは残る)。
func TestShareGroupLeave_Member(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	rr := newMockRecipeRepo()
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), rr)

	// Act
	err := svc.Leave(context.Background(), "u2")

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, gr.deletedIDs, "g1") // 解散していない
	group, _ := gr.FindByUserID(context.Background(), "u2")
	assert.Nil(t, group) // u2 は抜けた
	// 抜けた本人と残る owner の双方について残置レシピ状態が掃除される。
	assert.ElementsMatch(t, []string{"u2", "owner"}, rr.prunedUserIDs)
}

// 所有者が抜けた時、グループが解散されること。
func TestShareGroupLeave_OwnerDisbands(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	rr := newMockRecipeRepo()
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), rr)

	// Act
	err := svc.Leave(context.Background(), "owner")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, gr.deletedIDs, "g1")
	// 解散で全員が互いのレシピを見られなくなるため、全メンバーの状態を掃除する。
	assert.ElementsMatch(t, []string{"owner", "u2"}, rr.prunedUserIDs)
}

// グループ未所属で抜けようとした時、ErrNotInGroup が返ること。
func TestShareGroupLeave_NotInGroup(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo(), newMockShoppingListRepo(), newMockRecipeRepo())

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
	rr := newMockRecipeRepo()
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), rr)

	// Act
	err := svc.RemoveMember(context.Background(), "owner", "u2")

	// Assert
	require.NoError(t, err)
	group, _ := gr.FindByUserID(context.Background(), "u2")
	assert.Nil(t, group)
	// 外された u2 と残る owner の双方について残置レシピ状態が掃除される。
	assert.ElementsMatch(t, []string{"u2", "owner"}, rr.prunedUserIDs)
}

// 所有者でないメンバーが他人を外そうとした時、ErrNotGroupOwner が返ること。
func TestShareGroupRemoveMember_NotOwner(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2", "u3")
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

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
	svc := NewShareGroupService(gr, newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	_, err := svc.RegenerateInviteCode(context.Background(), "u2")

	// Assert
	assert.ErrorIs(t, err, ErrNotGroupOwner)
}

// メンバーが買い物リストの統合をオンにした時、自分の個人リストが物理削除されること。
func TestShareGroupSetShoppingListSharing_EnablingDeletesPersonalList(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	gr.shareShoppingList["u2"] = false // 個人運用中
	lr := newMockShoppingListRepo()
	require.NoError(t, lr.Create(context.Background(), &domain.ShoppingList{OwnerID: "u2"}))
	svc := NewShareGroupService(gr, lr, newMockRecipeRepo())

	// Act
	err := svc.SetShoppingListSharing(context.Background(), "u2", true)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, lr.deletedOwnerIDs, "u2")
	membership, _ := gr.FindMembership(context.Background(), "u2")
	assert.True(t, membership.ShareShoppingList)
}

// メンバーが買い物リストの統合をオフにした時、個人リストは削除されず設定だけ変わること。
func TestShareGroupSetShoppingListSharing_DisablingKeepsExistingList(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2") // seed は既定で ShareShoppingList=true
	lr := newMockShoppingListRepo()
	svc := NewShareGroupService(gr, lr, newMockRecipeRepo())

	// Act
	err := svc.SetShoppingListSharing(context.Background(), "u2", false)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, lr.deletedOwnerIDs)
	membership, _ := gr.FindMembership(context.Background(), "u2")
	assert.False(t, membership.ShareShoppingList)
}

// グループ未所属で切り替えようとした時、ErrNotInGroup が返ること。
func TestShareGroupSetShoppingListSharing_NotInGroup(t *testing.T) {
	// Arrange
	svc := NewShareGroupService(newMockShareGroupRepo(), newMockShoppingListRepo(), newMockRecipeRepo())

	// Act
	err := svc.SetShoppingListSharing(context.Background(), "u1", true)

	// Assert
	assert.ErrorIs(t, err, ErrNotInGroup)
}

// 所有者が自分の統合設定を変えようとした時、ErrForbidden が返り、所有者のリスト
// (= グループの共有リストそのもの)が消されないこと。
func TestShareGroupSetShoppingListSharing_OwnerForbidden(t *testing.T) {
	// Arrange
	gr := newMockShareGroupRepo()
	gr.seed("g1", "owner", "u2")
	lr := newMockShoppingListRepo()
	require.NoError(t, lr.Create(context.Background(), &domain.ShoppingList{OwnerID: "owner"}))
	svc := NewShareGroupService(gr, lr, newMockRecipeRepo())

	// Act
	err := svc.SetShoppingListSharing(context.Background(), "owner", true)

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Empty(t, lr.deletedOwnerIDs)
}
