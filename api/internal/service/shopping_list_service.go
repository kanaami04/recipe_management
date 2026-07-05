package service

import (
	"context"
	"fmt"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
)

type ShoppingListService interface {
	// Get は userID が見るべき買い物リストを返す。無ければ空のリストを作成して返す。
	Get(ctx context.Context, userID string) (*domain.ShoppingList, error)
	AddItem(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error)
	SetItemChecked(ctx context.Context, userID, listID, itemID string, checked bool) (*domain.ShoppingList, error)
	DeleteItem(ctx context.Context, userID, listID, itemID string) (*domain.ShoppingList, error)
	ClearChecked(ctx context.Context, userID, listID string) (*domain.ShoppingList, error)
	Reorder(ctx context.Context, userID, listID string, itemIDs []string) (*domain.ShoppingList, error)
	UpdateShares(ctx context.Context, userID, listID string, sharedUsers []request.SharedUserInput) (*domain.ShoppingList, error)
}

type shoppingListService struct {
	lists domain.ShoppingListRepository
	users domain.UserRepository
}

func NewShoppingListService(lists domain.ShoppingListRepository, users domain.UserRepository) ShoppingListService {
	return &shoppingListService{lists: lists, users: users}
}

func (s *shoppingListService) Get(ctx context.Context, userID string) (*domain.ShoppingList, error) {
	existing, err := s.lists.FindForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	// 見えるリストがまだ無ければ、自分が所有する空のリストを作って返す。
	list := &domain.ShoppingList{OwnerID: userID}
	if err := s.lists.Create(ctx, list); err != nil {
		// 別リクエストが同時に作成した場合は owner_id の一意制約で弾かれる。
		// その勝者を読み直して返す(二重作成を避ける)。
		if again, ferr := s.lists.FindForUser(ctx, userID); ferr == nil && again != nil {
			return again, nil
		}
		return nil, err
	}
	return s.lists.FindByID(ctx, list.ID)
}

func (s *shoppingListService) AddItem(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error) {
	if _, err := s.authorize(ctx, userID, listID); err != nil {
		return nil, err
	}
	item := &domain.ShoppingListItem{ShoppingListID: listID, Name: name}
	if err := s.lists.AddItem(ctx, item); err != nil {
		return nil, err
	}
	return s.lists.FindByID(ctx, listID)
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
	return s.lists.FindByID(ctx, listID)
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
	return s.lists.FindByID(ctx, listID)
}

func (s *shoppingListService) ClearChecked(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	if _, err := s.authorize(ctx, userID, listID); err != nil {
		return nil, err
	}
	if err := s.lists.DeleteCheckedItems(ctx, listID); err != nil {
		return nil, err
	}
	return s.lists.FindByID(ctx, listID)
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
	return s.lists.FindByID(ctx, listID)
}

func (s *shoppingListService) UpdateShares(ctx context.Context, userID, listID string, sharedUsers []request.SharedUserInput) (*domain.ShoppingList, error) {
	list, err := s.authorize(ctx, userID, listID)
	if err != nil {
		return nil, err
	}
	shared := make([]domain.User, 0, len(sharedUsers))
	for _, su := range sharedUsers {
		u, err := s.users.FindByUsername(ctx, su.Username)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, fmt.Errorf("%w: %s", ErrSharedUserNotFound, su.Username)
		}
		shared = append(shared, *u)
	}
	list.SharedUsers = shared
	if err := s.lists.ReplaceSharedUsers(ctx, list); err != nil {
		return nil, err
	}
	return s.lists.FindByID(ctx, listID)
}

// authorize は listID のリストを取り出し、userID が操作できる(所有 or 共有先)か検証する。
// 見つからなければ ErrNotFound、権限が無ければ ErrForbidden。
func (s *shoppingListService) authorize(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	list, err := s.lists.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, ErrNotFound
	}
	if !canModifyList(list, userID) {
		return nil, ErrForbidden
	}
	return list, nil
}

// canModifyList は owner もしくは共有先ユーザーであれば true(recipe の canModify 相当)。
func canModifyList(l *domain.ShoppingList, userID string) bool {
	if l.OwnerID == userID {
		return true
	}
	for i := range l.SharedUsers {
		if l.SharedUsers[i].ID == userID {
			return true
		}
	}
	return false
}

func itemBelongsToList(l *domain.ShoppingList, itemID string) bool {
	for i := range l.Items {
		if l.Items[i].ID == itemID {
			return true
		}
	}
	return false
}
