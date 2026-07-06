package service

import (
	"context"
	"strings"
	"time"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/invite"
)

// inviteCodeTTL は招待コードの有効期間。
const inviteCodeTTL = 7 * 24 * time.Hour

// defaultGroupName は名前未指定時のグループ名。
const defaultGroupName = "マイグループ"

type ShareGroupService interface {
	// GetMine は userID が所属するグループを返す(未所属なら nil)。
	GetMine(ctx context.Context, userID string) (*domain.ShareGroup, error)
	// Create はグループを作成し、userID を所有者兼メンバーにする。既に所属していれば ErrAlreadyInGroup。
	Create(ctx context.Context, userID, name string) (*domain.ShareGroup, error)
	// Join は招待コードでグループに参加する。既に所属していれば ErrAlreadyInGroup、
	// コードが無効/期限切れなら ErrInviteCodeInvalid。shareShoppingList が true のときは
	// 参加と同時に買い物リストを統合し、参加者自身の個人リストは物理削除する。
	Join(ctx context.Context, userID, code string, shareShoppingList bool) (*domain.ShareGroup, error)
	// Leave はグループを抜ける。所有者が抜ける場合はグループを解散する。未所属なら ErrNotInGroup。
	Leave(ctx context.Context, userID string) error
	// RemoveMember は所有者が他メンバーを外す。所有者以外の操作は ErrNotGroupOwner。
	RemoveMember(ctx context.Context, ownerID, targetUserID string) error
	// RegenerateInviteCode は所有者が招待コードを再発行する(旧コードを失効)。
	RegenerateInviteCode(ctx context.Context, ownerID string) (*domain.ShareGroup, error)
	// SetShoppingListSharing は userID 自身の買い物リスト統合設定を切り替える。未所属なら
	// ErrNotInGroup。true にする(統合する)ときは、自分の個人リストを物理削除する。
	// false にする(個人運用に戻す)ときは何も削除しない。次回アクセス時に新規の空リストができる。
	SetShoppingListSharing(ctx context.Context, userID string, share bool) error
	// ShoppingListSharing は userID 自身の買い物リスト統合設定を返す(どのグループにも
	// 属さないときは true。個人リストのみなのでこの値自体は意味を持たない)。
	ShoppingListSharing(ctx context.Context, userID string) (bool, error)
}

type shareGroupService struct {
	groups  domain.ShareGroupRepository
	lists   domain.ShoppingListRepository
	recipes domain.RecipeRepository
}

func NewShareGroupService(groups domain.ShareGroupRepository, lists domain.ShoppingListRepository, recipes domain.RecipeRepository) ShareGroupService {
	return &shareGroupService{groups: groups, lists: lists, recipes: recipes}
}

func (s *shareGroupService) GetMine(ctx context.Context, userID string) (*domain.ShareGroup, error) {
	return s.groups.FindByUserID(ctx, userID)
}

func (s *shareGroupService) Create(ctx context.Context, userID, name string) (*domain.ShareGroup, error) {
	existing, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGroup
	}
	code, err := invite.Code()
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultGroupName
	}
	group := &domain.ShareGroup{
		Name:                name,
		OwnerID:             userID,
		InviteCode:          code,
		InviteCodeExpiresAt: time.Now().Add(inviteCodeTTL),
	}
	if err := s.groups.Create(ctx, group); err != nil {
		// 別リクエストが同時に参加/作成した場合は user_id の一意制約で弾かれる。
		// 事前チェックを抜けた競合を ErrAlreadyInGroup に正規化する(生 DB エラーで 500 にしない)。
		return nil, s.mapMembershipConflict(ctx, userID, err)
	}
	return s.groups.FindByID(ctx, group.ID)
}

// mapMembershipConflict は、書き込み後に userID が既にどこかのグループへ所属していれば
// ErrAlreadyInGroup を返す(check-then-insert の競合を一意制約で受けたケース)。そうでなければ
// 元のエラーをそのまま返す。
func (s *shareGroupService) mapMembershipConflict(ctx context.Context, userID string, err error) error {
	if existing, ferr := s.groups.FindByUserID(ctx, userID); ferr == nil && existing != nil {
		return ErrAlreadyInGroup
	}
	return err
}

func (s *shareGroupService) Join(ctx context.Context, userID, code string, shareShoppingList bool) (*domain.ShareGroup, error) {
	existing, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGroup
	}
	group, err := s.groups.FindByInviteCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if group == nil || time.Now().After(group.InviteCodeExpiresAt) {
		return nil, ErrInviteCodeInvalid
	}
	if err := s.groups.AddMember(ctx, group.ID, userID, shareShoppingList); err != nil {
		return nil, s.mapMembershipConflict(ctx, userID, err)
	}
	if shareShoppingList {
		// 買い物リストをグループへ統合する: 自分の個人リストは物理削除する。
		if err := s.lists.DeleteByOwnerID(ctx, userID); err != nil {
			return nil, err
		}
	}
	return s.groups.FindByID(ctx, group.ID)
}

func (s *shareGroupService) SetShoppingListSharing(ctx context.Context, userID string, share bool) error {
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrNotInGroup
	}
	// 所有者の買い物リストはそのものが「グループの共有リスト」なので、この設定は
	// 所有者には意味を持たない。ここで弾かないと share=true 指定時に統合先である
	// 共有リスト自体を物理削除してしまう。
	if group.OwnerID == userID {
		return ErrForbidden
	}
	// 先にフラグを更新し、その後で個人リストを消す。逆順だと、削除は成功したのに
	// フラグ更新が失敗した場合にデータだけ失われ設定は変わっていない状態になる。
	if err := s.groups.UpdateShareShoppingList(ctx, userID, share); err != nil {
		return err
	}
	if share {
		// 統合する: 自分の個人リストを物理削除する(統合をやめて個人運用に戻した後に
		// 新規で作られる分には影響しない)。
		return s.lists.DeleteByOwnerID(ctx, userID)
	}
	return nil
}

func (s *shareGroupService) ShoppingListSharing(ctx context.Context, userID string) (bool, error) {
	membership, err := s.groups.FindMembership(ctx, userID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return true, nil
	}
	return membership.ShareShoppingList, nil
}

func (s *shareGroupService) Leave(ctx context.Context, userID string) error {
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrNotInGroup
	}
	// 所有者が抜ける場合はグループを解散する(メンバー行は CASCADE で消える)。
	if group.OwnerID == userID {
		members := userIDsOf(group.Members)
		if err := s.groups.Delete(ctx, group.ID); err != nil {
			return err
		}
		// 解散で全員が互いのレシピを見られなくなる。全メンバーの残置状態を掃除する。
		return s.pruneRecipeState(ctx, members)
	}
	// 抜けた本人と残るメンバーが互いのレシピを見られなくなる。両者の残置状態を掃除する。
	affected := append(userIDsOf(membersExcept(group.Members, userID)), userID)
	if err := s.groups.RemoveMember(ctx, group.ID, userID); err != nil {
		return err
	}
	return s.pruneRecipeState(ctx, affected)
}

func (s *shareGroupService) RemoveMember(ctx context.Context, ownerID, targetUserID string) error {
	group, err := s.groups.FindByUserID(ctx, ownerID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrNotInGroup
	}
	if group.OwnerID != ownerID {
		return ErrNotGroupOwner
	}
	// 所有者自身は外せない(抜けるなら Leave = 解散)。
	if targetUserID == ownerID {
		return ErrForbidden
	}
	// 外された本人と残るメンバーが互いのレシピを見られなくなる。両者の残置状態を掃除する。
	affected := append(userIDsOf(membersExcept(group.Members, targetUserID)), targetUserID)
	if err := s.groups.RemoveMember(ctx, group.ID, targetUserID); err != nil {
		return err
	}
	return s.pruneRecipeState(ctx, affected)
}

// pruneRecipeState は userIDs それぞれについて、見えなくなったレシピに残る
// recipe_archives / recipe_orders を掃除する(メンバー行の更新後に呼ぶ)。
func (s *shareGroupService) pruneRecipeState(ctx context.Context, userIDs []string) error {
	for _, uid := range userIDs {
		if err := s.recipes.PruneRecipeState(ctx, uid); err != nil {
			return err
		}
	}
	return nil
}

func (s *shareGroupService) RegenerateInviteCode(ctx context.Context, ownerID string) (*domain.ShareGroup, error) {
	group, err := s.groups.FindByUserID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrNotInGroup
	}
	if group.OwnerID != ownerID {
		return nil, ErrNotGroupOwner
	}
	code, err := invite.Code()
	if err != nil {
		return nil, err
	}
	if err := s.groups.UpdateInviteCode(ctx, group.ID, code, time.Now().Add(inviteCodeTTL)); err != nil {
		return nil, err
	}
	return s.groups.FindByID(ctx, group.ID)
}

// userIDsOf は users の ID を並び順のまま取り出す。新しいスライスを返すため、
// 呼び出し側で append しても元の Members を汚さない。
func userIDsOf(users []domain.User) []string {
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}
