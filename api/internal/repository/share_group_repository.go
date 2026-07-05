package repository

import (
	"context"
	"errors"
	"time"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type shareGroupRepository struct {
	db *gorm.DB
}

func NewShareGroupRepository(db *gorm.DB) domain.ShareGroupRepository {
	return &shareGroupRepository{db: db}
}

func (r *shareGroupRepository) Create(ctx context.Context, group *domain.ShareGroup) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// belongs-to(Owner)は FK 定義用で、書き込みでは巻き込まない。
		if err := tx.Omit("Owner", "Members").Create(group).Error; err != nil {
			return err
		}
		// 所有者自身もメンバー行を持つ。
		member := &domain.ShareGroupMember{GroupID: group.ID, UserID: group.OwnerID}
		return tx.Omit("Group", "User").Create(member).Error
	})
}

// membersOf は groupID のメンバー(参加順)を返す。
func (r *shareGroupRepository) membersOf(ctx context.Context, groupID string) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).
		Joins("JOIN share_group_members ON share_group_members.user_id = users.id").
		Where("share_group_members.group_id = ?", groupID).
		Order("share_group_members.joined_at ASC").
		Order("users.id ASC").
		Find(&users).Error
	return users, err
}

func (r *shareGroupRepository) FindByUserID(ctx context.Context, userID string) (*domain.ShareGroup, error) {
	var member domain.ShareGroupMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, member.GroupID)
}

func (r *shareGroupRepository) FindByID(ctx context.Context, id string) (*domain.ShareGroup, error) {
	var group domain.ShareGroup
	err := r.db.WithContext(ctx).Preload("Owner").Where("id = ?", id).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	members, err := r.membersOf(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.Members = members
	return &group, nil
}

func (r *shareGroupRepository) FindByInviteCode(ctx context.Context, code string) (*domain.ShareGroup, error) {
	var group domain.ShareGroup
	err := r.db.WithContext(ctx).Preload("Owner").Where("invite_code = ?", code).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	members, err := r.membersOf(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.Members = members
	return &group, nil
}

func (r *shareGroupRepository) MemberIDs(ctx context.Context, userID string) ([]string, error) {
	// userID の所属グループのメンバー全員(自分を含む)。未所属なら空。
	sub := r.db.Table("share_group_members").Select("group_id").Where("user_id = ?", userID)
	var ids []string
	err := r.db.WithContext(ctx).
		Table("share_group_members").
		Where("group_id IN (?)", sub).
		Pluck("user_id", &ids).Error
	return ids, err
}

func (r *shareGroupRepository) AddMember(ctx context.Context, groupID, userID string) error {
	member := &domain.ShareGroupMember{GroupID: groupID, UserID: userID}
	return r.db.WithContext(ctx).Omit("Group", "User").Create(member).Error
}

func (r *shareGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&domain.ShareGroupMember{}).Error
}

func (r *shareGroupRepository) UpdateInviteCode(ctx context.Context, groupID, code string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&domain.ShareGroup{ID: groupID}).
		Updates(map[string]any{"invite_code": code, "invite_code_expires_at": expiresAt}).Error
}

func (r *shareGroupRepository) Delete(ctx context.Context, groupID string) error {
	// メンバー行は FK の ON DELETE CASCADE で消える。
	return r.db.WithContext(ctx).Delete(&domain.ShareGroup{ID: groupID}).Error
}
