package repository

import (
	"context"
	"testing"
	"time"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/invite"
	"recipe-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arrangeShareGroupRepo は結合テスト用の共通セットアップを行う。
func arrangeShareGroupRepo(t *testing.T) (context.Context, domain.ShareGroupRepository) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	return context.Background(), NewShareGroupRepository(testDB)
}

// makeGroup は owner 所有のグループを作る(招待コード付き)。
func makeGroup(t *testing.T, ctx context.Context, repo domain.ShareGroupRepository, owner *domain.User) *domain.ShareGroup {
	t.Helper()
	code, err := invite.Code()
	require.NoError(t, err)
	g := &domain.ShareGroup{Name: "我が家", OwnerID: owner.ID, InviteCode: code, InviteCodeExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, repo.Create(ctx, g))
	return g
}

// グループを作成した時、所有者がメンバーに含まれること。
func TestShareGroupRepo_Create_AddsOwnerAsMember(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner := seedUser(t, "owner")

	// Act
	g := makeGroup(t, ctx, repo, owner)

	// Assert
	got, err := repo.FindByID(ctx, g.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Len(t, got.Members, 1)
	assert.Equal(t, owner.ID, got.Members[0].ID)
}

// メンバーを追加した時、FindByUserID でそのメンバーからグループが引けること。
func TestShareGroupRepo_AddMember_ResolvesByUser(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner := seedUser(t, "owner")
	member := seedUser(t, "member")
	g := makeGroup(t, ctx, repo, owner)

	// Act
	require.NoError(t, repo.AddMember(ctx, g.ID, member.ID))

	// Assert
	got, err := repo.FindByUserID(ctx, member.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, g.ID, got.ID)
}

// 招待コードでグループを引けること。
func TestShareGroupRepo_FindByInviteCode(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner := seedUser(t, "owner")
	g := makeGroup(t, ctx, repo, owner)

	// Act
	got, err := repo.FindByInviteCode(ctx, g.InviteCode)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, g.ID, got.ID)
}

// MemberIDs が同じグループの全メンバー(自分を含む)を返すこと。
func TestShareGroupRepo_MemberIDs_ReturnsGroupmates(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner := seedUser(t, "owner")
	member := seedUser(t, "member")
	g := makeGroup(t, ctx, repo, owner)
	require.NoError(t, repo.AddMember(ctx, g.ID, member.ID))

	// Act
	ids, err := repo.MemberIDs(ctx, member.ID)

	// Assert
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{owner.ID, member.ID}, ids)
}

// どのグループにも属さないユーザーの MemberIDs は空であること。
func TestShareGroupRepo_MemberIDs_EmptyForNonMember(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	stranger := seedUser(t, "stranger")

	// Act
	ids, err := repo.MemberIDs(ctx, stranger.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, ids)
}

// 1 ユーザーは 2 つのグループに同時に入れないこと(user_id の一意制約)。
func TestShareGroupRepo_AddMember_RejectsSecondGroup(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner1 := seedUser(t, "owner1")
	owner2 := seedUser(t, "owner2")
	member := seedUser(t, "member")
	g1 := makeGroup(t, ctx, repo, owner1)
	g2 := makeGroup(t, ctx, repo, owner2)
	require.NoError(t, repo.AddMember(ctx, g1.ID, member.ID))

	// Act: 2 つ目のグループへ追加しようとする
	err := repo.AddMember(ctx, g2.ID, member.ID)

	// Assert: 一意制約で弾かれる
	assert.Error(t, err)
}

// グループを解散した時、メンバー行も CASCADE で消えること。
func TestShareGroupRepo_Delete_CascadesMembers(t *testing.T) {
	// Arrange
	ctx, repo := arrangeShareGroupRepo(t)
	owner := seedUser(t, "owner")
	member := seedUser(t, "member")
	g := makeGroup(t, ctx, repo, owner)
	require.NoError(t, repo.AddMember(ctx, g.ID, member.ID))

	// Act
	require.NoError(t, repo.Delete(ctx, g.ID))

	// Assert: メンバーからグループが引けなくなる
	got, err := repo.FindByUserID(ctx, member.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}
