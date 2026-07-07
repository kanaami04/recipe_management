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
// SharedUsers はグループメンバーから service が詰める計算値のため preload しない。
func preloadList(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Owner").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("checked ASC").Order("position ASC").Order("id ASC")
		})
}

func (r *shoppingListRepository) FindByOwnerID(ctx context.Context, ownerID string) (*domain.ShoppingList, error) {
	var list domain.ShoppingList
	err := preloadList(r.db.WithContext(ctx)).
		Where("owner_id = ?", ownerID).
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
	return r.db.WithContext(ctx).Omit("Owner").Create(list).Error
}

// nextPosition は listID の項目を末尾へ回すための position(現在の最大 + 1、項目が無ければ 0)を返す。
func nextPosition(db *gorm.DB, listID string) (int, error) {
	var maxPos *int
	if err := db.
		Model(&domain.ShoppingListItem{}).
		Where("shopping_list_id = ?", listID).
		Select("MAX(position)").
		Scan(&maxPos).Error; err != nil {
		return 0, err
	}
	if maxPos == nil {
		return 0, nil
	}
	return *maxPos + 1, nil
}

func (r *shoppingListRepository) AddItem(ctx context.Context, item *domain.ShoppingListItem) error {
	// 追加項目は末尾へ回すため、リスト内の現在の最大 position + 1 を割り当てる。
	pos, err := nextPosition(r.db.WithContext(ctx), item.ShoppingListID)
	if err != nil {
		return err
	}
	item.Position = pos
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *shoppingListRepository) AddItems(ctx context.Context, items []*domain.ShoppingListItem) error {
	if len(items) == 0 {
		return nil
	}
	// 全項目が同じリストに属する前提。現在の最大 position + 1 から連番で末尾へ積む。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		next, err := nextPosition(tx, items[0].ShoppingListID)
		if err != nil {
			return err
		}
		for i, item := range items {
			item.Position = next + i
		}
		return tx.Create(items).Error
	})
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

func (r *shoppingListRepository) DeleteByOwnerID(ctx context.Context, ownerID string) error {
	// 項目は FK の ON DELETE CASCADE で一緒に消える。無ければ 0 件影響で成功扱い。
	return r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Delete(&domain.ShoppingList{}).Error
}
