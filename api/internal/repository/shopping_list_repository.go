package repository

import (
	"context"
	"errors"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type shoppingListRepository struct {
	db *gorm.DB
}

func NewShoppingListRepository(db *gorm.DB) domain.ShoppingListRepository {
	return &shoppingListRepository{db: db}
}

// preloadList は関連を全て eager load する。項目はチェック済みを末尾へ回し、
// 同グループ内は手動並び順(position)→ 採番順(id)で安定させる。
func preloadList(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Owner").
		Preload("SharedUsers").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("checked ASC").Order("position ASC").Order("id ASC")
		})
}

// sharedListIDs は共有先に userID を含むリスト ID のサブクエリを返す。
func sharedListIDs(db *gorm.DB, userID string) *gorm.DB {
	return db.Table("shopping_list_shares").
		Select("shopping_list_id").
		Where("user_id = ?", userID)
}

func (r *shoppingListRepository) FindForUser(ctx context.Context, userID string) (*domain.ShoppingList, error) {
	db := r.db.WithContext(ctx)
	// 共有されたリストを優先する(世帯で 1 つのリストを共有する運用のため、
	// 共有相手のページには共有元のリストを見せる)。同点は id で安定させる。
	var list domain.ShoppingList
	err := preloadList(db).
		Where("id IN (?)", sharedListIDs(db, userID)).
		Order("id ASC").
		First(&list).Error
	if err == nil {
		return &list, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// 共有が無ければ自分が所有するリストを返す。
	err = preloadList(db).
		Where("owner_id = ?", userID).
		Order("id ASC").
		First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *shoppingListRepository) FindByID(ctx context.Context, id string) (*domain.ShoppingList, error) {
	var list domain.ShoppingList
	err := preloadList(r.db.WithContext(ctx)).Where("id = ?", id).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *shoppingListRepository) Create(ctx context.Context, list *domain.ShoppingList) error {
	// 空のリストを作る。belongs-to(Owner)は FK 定義用で書き込みでは巻き込まない。
	return r.db.WithContext(ctx).Omit("Owner", "SharedUsers").Create(list).Error
}

func (r *shoppingListRepository) ReplaceSharedUsers(ctx context.Context, list *domain.ShoppingList) error {
	// SharedUsers(m2m)は既存ユーザーへの参照のため Replace で中間テーブルのみ操作する。
	return r.db.WithContext(ctx).Model(list).Association("SharedUsers").Replace(list.SharedUsers)
}

func (r *shoppingListRepository) AddItem(ctx context.Context, item *domain.ShoppingListItem) error {
	// 追加項目は末尾へ回すため、リスト内の現在の最大 position + 1 を割り当てる。
	var maxPos *int
	if err := r.db.WithContext(ctx).
		Model(&domain.ShoppingListItem{}).
		Where("shopping_list_id = ?", item.ShoppingListID).
		Select("MAX(position)").
		Scan(&maxPos).Error; err != nil {
		return err
	}
	if maxPos != nil {
		item.Position = *maxPos + 1
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *shoppingListRepository) Reorder(ctx context.Context, listID string, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}
	// 先頭から position 0,1,2... を振り直す。listID 条件で他リストの項目は触らない。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, itemID := range itemIDs {
			if err := tx.Model(&domain.ShoppingListItem{}).
				Where("id = ? AND shopping_list_id = ?", itemID, listID).
				Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *shoppingListRepository) SetItemChecked(ctx context.Context, itemID string, checked bool) error {
	return r.db.WithContext(ctx).
		Model(&domain.ShoppingListItem{ID: itemID}).
		Update("checked", checked).Error
}

func (r *shoppingListRepository) DeleteItem(ctx context.Context, itemID string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", itemID).
		Delete(&domain.ShoppingListItem{}).Error
}

func (r *shoppingListRepository) DeleteCheckedItems(ctx context.Context, listID string) error {
	return r.db.WithContext(ctx).
		Where("shopping_list_id = ? AND checked = ?", listID, true).
		Delete(&domain.ShoppingListItem{}).Error
}
