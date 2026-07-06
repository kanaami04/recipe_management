package service

import (
	"context"

	"recipe-backend/internal/domain"
)

type ShoppingListService interface {
	// Get は userID が見るべき買い物リストを返す。無ければ空のリストを作成して返す。
	// グループ所属時はグループの 1 リスト(= グループ所有者のリスト)を全員で共同編集する。
	Get(ctx context.Context, userID string) (*domain.ShoppingList, error)
	AddItem(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error)
	SetItemChecked(ctx context.Context, userID, listID, itemID string, checked bool) (*domain.ShoppingList, error)
	DeleteItem(ctx context.Context, userID, listID, itemID string) (*domain.ShoppingList, error)
	ClearChecked(ctx context.Context, userID, listID string) (*domain.ShoppingList, error)
	Reorder(ctx context.Context, userID, listID string, itemIDs []string) (*domain.ShoppingList, error)
}

type shoppingListService struct {
	lists  domain.ShoppingListRepository
	groups domain.ShareGroupRepository
}

func NewShoppingListService(lists domain.ShoppingListRepository, groups domain.ShareGroupRepository) ShoppingListService {
	return &shoppingListService{lists: lists, groups: groups}
}

func (s *shoppingListService) Get(ctx context.Context, userID string) (*domain.ShoppingList, error) {
	// グループはここで一度だけ解決し、リスト所有者の決定と SharedUsers 詰めの両方に使う。
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	ownerID, err := s.resolveOwnerID(ctx, userID, group)
	if err != nil {
		return nil, err
	}
	existing, err := s.lists.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.applyShared(ctx, group, existing)
	}
	// リストがまだ無ければ、対象所有者の空のリストを作って返す。
	// (統合をやめて個人運用に戻した直後は、統合時に消した個人リストがまだ無いのでここで新規に作られる)
	list := &domain.ShoppingList{OwnerID: ownerID}
	if err := s.lists.Create(ctx, list); err != nil {
		// 別リクエストが同時に作成した場合は owner_id の一意制約で弾かれる。
		// その勝者を読み直して返す(二重作成を避ける)。
		if again, ferr := s.lists.FindByOwnerID(ctx, ownerID); ferr == nil && again != nil {
			return s.applyShared(ctx, group, again)
		}
		return nil, err
	}
	created, err := s.lists.FindByID(ctx, list.ID)
	if err != nil {
		return nil, err
	}
	return s.applyShared(ctx, group, created)
}

// resolveOwnerID は userID が見るべきリストの所有者 ID を返す。グループ所属かつ自分の
// ShareShoppingList が true のときだけグループ所有者のリストへ倒す。それ以外(未所属、
// または統合をやめて個人運用を選んでいる)は自分自身のリストを見る。
func (s *shoppingListService) resolveOwnerID(ctx context.Context, userID string, group *domain.ShareGroup) (string, error) {
	if group == nil {
		return userID, nil
	}
	membership, err := s.groups.FindMembership(ctx, userID)
	if err != nil {
		return "", err
	}
	if membership != nil && membership.ShareShoppingList {
		return group.OwnerID, nil
	}
	return userID, nil
}

func (s *shoppingListService) AddItem(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error) {
	if _, err := s.authorize(ctx, userID, listID); err != nil {
		return nil, err
	}
	item := &domain.ShoppingListItem{ShoppingListID: listID, Name: name}
	if err := s.lists.AddItem(ctx, item); err != nil {
		return nil, err
	}
	return s.reload(ctx, userID, listID)
}

func (s *shoppingListService) SetItemChecked(ctx context.Context, userID, listID, itemID string, checked bool) (*domain.ShoppingList, error) {
	list, err := s.authorize(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	if !itemBelongsToList(list, itemID) {
		return nil, ErrNotFound
	}
	if err := s.lists.SetItemChecked(ctx, itemID, checked); err != nil {
		return nil, err
	}
	return s.reload(ctx, userID, listID)
}

func (s *shoppingListService) DeleteItem(ctx context.Context, userID, listID, itemID string) (*domain.ShoppingList, error) {
	list, err := s.authorize(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	if !itemBelongsToList(list, itemID) {
		return nil, ErrNotFound
	}
	if err := s.lists.DeleteItem(ctx, itemID); err != nil {
		return nil, err
	}
	return s.reload(ctx, userID, listID)
}

func (s *shoppingListService) ClearChecked(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	if _, err := s.authorize(ctx, userID, listID); err != nil {
		return nil, err
	}
	if err := s.lists.DeleteCheckedItems(ctx, listID); err != nil {
		return nil, err
	}
	return s.reload(ctx, userID, listID)
}

// Reorder は listID の項目表示順を itemIDs の並びで保存する。指定 ID は全てこのリストの
// 項目でなければならない(見えない項目の順序を書こうとしたら ErrNotFound)。
func (s *shoppingListService) Reorder(ctx context.Context, userID, listID string, itemIDs []string) (*domain.ShoppingList, error) {
	list, err := s.authorize(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(list.Items))
	for i := range list.Items {
		allowed[list.Items[i].ID] = struct{}{}
	}
	// このリストの項目でないものを弾きつつ重複を除く(最初の出現順を維持)。
	seen := make(map[string]struct{}, len(itemIDs))
	deduped := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		if _, ok := allowed[itemID]; !ok {
			return nil, ErrNotFound
		}
		if _, dup := seen[itemID]; dup {
			continue
		}
		seen[itemID] = struct{}{}
		deduped = append(deduped, itemID)
	}
	if err := s.lists.Reorder(ctx, listID, deduped); err != nil {
		return nil, err
	}
	return s.reload(ctx, userID, listID)
}

// reload は listID のリストを読み直し、SharedUsers を詰めて返す。
func (s *shoppingListService) reload(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	list, err := s.lists.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, ErrNotFound
	}
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.applyShared(ctx, group, list)
}

// applyShared はリストがグループの共有リストそのもの(list.OwnerID == group.OwnerID)の
// ときだけ、SharedUsers に「買い物リストを統合しているメンバー(owner を除く)」を詰める。
// 統合をやめて個人運用に戻したメンバー自身のリストは、グループに属していても誰とも
// 共有されていないので空のまま返す。
func (s *shoppingListService) applyShared(ctx context.Context, group *domain.ShareGroup, list *domain.ShoppingList) (*domain.ShoppingList, error) {
	if group == nil || list.OwnerID != group.OwnerID {
		list.SharedUsers = nil
		return list, nil
	}
	sharingIDs, err := s.groups.SharingMemberIDs(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	sharing := make(map[string]struct{}, len(sharingIDs))
	for _, id := range sharingIDs {
		sharing[id] = struct{}{}
	}
	shared := make([]domain.User, 0, len(group.Members))
	for _, m := range membersExcept(group.Members, list.OwnerID) {
		if _, ok := sharing[m.ID]; ok {
			shared = append(shared, m)
		}
	}
	list.SharedUsers = shared
	return list, nil
}

// authorize は listID のリストを取り出し、userID が操作できる(所有者、または owner と同じ
// シェアグループのメンバー)か検証する。見つからなければ ErrNotFound、権限が無ければ ErrForbidden。
func (s *shoppingListService) authorize(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	list, err := s.lists.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, ErrNotFound
	}
	ok, err := s.canModifyList(ctx, list, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return list, nil
}

// canModifyList は userID が list を編集できる(所有者、または「同じグループのメンバー」かつ
// 「自分自身が買い物リストをグループへ統合している」)かを返す。統合をやめたメンバーは、
// グループ所有者のリストに対する操作権を持たない(自分のリストは第一分岐でカバーされる)。
func (s *shoppingListService) canModifyList(ctx context.Context, l *domain.ShoppingList, userID string) (bool, error) {
	if l.OwnerID == userID {
		return true, nil
	}
	membership, err := s.groups.FindMembership(ctx, userID)
	if err != nil {
		return false, err
	}
	if membership == nil || !membership.ShareShoppingList {
		return false, nil
	}
	memberIDs, err := s.groups.MemberIDs(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, id := range memberIDs {
		if id == l.OwnerID {
			return true, nil
		}
	}
	return false, nil
}

func itemBelongsToList(l *domain.ShoppingList, itemID string) bool {
	for i := range l.Items {
		if l.Items[i].ID == itemID {
			return true
		}
	}
	return false
}
