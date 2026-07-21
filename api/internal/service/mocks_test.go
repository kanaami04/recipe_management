package service

import (
	"context"
	"time"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/id"
)

// --- RecipeRepository のモック ---

type mockRecipeRepo struct {
	store        map[string]*domain.Recipe
	created      *domain.Recipe
	updated      *domain.Recipe
	deletedIDs   []string
	reorderedIDs []string
	// archived[userID][recipeID] = true でアーカイブ済み。
	archived map[string]map[string]bool
	// prunedUserIDs は PruneRecipeState を呼ばれたユーザーを呼び出し順に記録する。
	prunedUserIDs []string
}

func newMockRecipeRepo() *mockRecipeRepo {
	return &mockRecipeRepo{store: map[string]*domain.Recipe{}}
}

func (m *mockRecipeRepo) FindAllForUser(_ context.Context, _ string) ([]domain.Recipe, error) {
	var out []domain.Recipe
	for _, r := range m.store {
		out = append(out, *r)
	}
	return out, nil
}
func (m *mockRecipeRepo) FindByID(_ context.Context, id string) (*domain.Recipe, error) {
	r, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	// コピーを返す(サービスが SharedUsers を詰めても store / created / updated を汚さない)。
	cp := *r
	return &cp, nil
}
func (m *mockRecipeRepo) Create(_ context.Context, recipe *domain.Recipe) error {
	if recipe.ID == "" {
		recipe.ID = id.New()
	}
	cp := *recipe
	m.store[recipe.ID] = &cp
	m.created = &cp
	return nil
}
func (m *mockRecipeRepo) Update(_ context.Context, recipe *domain.Recipe) error {
	cp := *recipe
	m.store[recipe.ID] = &cp
	m.updated = &cp
	return nil
}
func (m *mockRecipeRepo) Delete(_ context.Context, recipe *domain.Recipe) error {
	delete(m.store, recipe.ID)
	m.deletedIDs = append(m.deletedIDs, recipe.ID)
	return nil
}
func (m *mockRecipeRepo) Reorder(_ context.Context, _ string, recipeIDs []string) error {
	m.reorderedIDs = recipeIDs
	return nil
}
func (m *mockRecipeRepo) SetArchived(_ context.Context, userID, recipeID string, archived bool) error {
	if m.archived == nil {
		m.archived = map[string]map[string]bool{}
	}
	if m.archived[userID] == nil {
		m.archived[userID] = map[string]bool{}
	}
	m.archived[userID][recipeID] = archived
	return nil
}
func (m *mockRecipeRepo) IsArchived(_ context.Context, userID, recipeID string) (bool, error) {
	return m.archived[userID][recipeID], nil
}
func (m *mockRecipeRepo) PruneRecipeState(_ context.Context, userID string) error {
	m.prunedUserIDs = append(m.prunedUserIDs, userID)
	return nil
}

// --- UserRepository のモック ---

type mockUserRepo struct {
	byName            map[string]*domain.User
	byEmail           map[string]*domain.User
	byID              map[string]*domain.User
	all               []domain.User
	updated           *domain.User
	passwordChangedID string
	newPasswordHash   string
	deletedUserID     string
	avatarKeyUpdates  []*string // UpdateAvatarKey に渡された key を呼び出し順に記録
	emailVerifiedSet  *bool     // SetEmailVerified に最後に渡された値(未呼び出しなら nil)
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	return m.byName[username], nil
}
func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return m.byEmail[email], nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	return m.byID[id], nil
}
func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = id.New()
	}
	if m.byName == nil {
		m.byName = map[string]*domain.User{}
	}
	if m.byEmail == nil {
		m.byEmail = map[string]*domain.User{}
	}
	m.byName[user.Username] = user
	m.byEmail[user.Email] = user
	if m.byID == nil {
		m.byID = map[string]*domain.User{}
	}
	m.byID[user.ID] = user
	return nil
}
func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.updated = user
	return nil
}
func (m *mockUserRepo) UpdatePassword(_ context.Context, userID, passwordHash string) error {
	m.passwordChangedID = userID
	m.newPasswordHash = passwordHash
	return nil
}
func (m *mockUserRepo) Delete(_ context.Context, userID string) error {
	m.deletedUserID = userID
	return nil
}
func (m *mockUserRepo) UpdateAvatarKey(_ context.Context, userID string, key *string) error {
	m.avatarKeyUpdates = append(m.avatarKeyUpdates, key)
	if u, ok := m.byID[userID]; ok {
		u.AvatarKey = key
	}
	return nil
}
func (m *mockUserRepo) SetEmailVerified(_ context.Context, userID string, verified bool) error {
	m.emailVerifiedSet = &verified
	if u, ok := m.byID[userID]; ok {
		u.EmailVerified = verified
	}
	return nil
}

// --- Mailer のモック ---

type mockMailer struct {
	verifyTo, verifyLink string // SendEmailVerification に渡された宛先・リンク
	resetTo, resetLink   string // SendPasswordReset に渡された宛先・リンク
	err                  error
}

func (m *mockMailer) SendEmailVerification(_ context.Context, toEmail, link string) error {
	m.verifyTo, m.verifyLink = toEmail, link
	return m.err
}
func (m *mockMailer) SendPasswordReset(_ context.Context, toEmail, link string) error {
	m.resetTo, m.resetLink = toEmail, link
	return m.err
}

// --- LabelRepository のモック ---

type mockLabelRepo struct {
	store      map[string]*domain.Label // id -> label
	created    *domain.Label
	renamedTo  string
	deletedIDs []string
	err        error
}

func newMockLabelRepo() *mockLabelRepo {
	return &mockLabelRepo{store: map[string]*domain.Label{}}
}

func (m *mockLabelRepo) FindAllForOwner(_ context.Context, ownerID string) ([]domain.Label, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []domain.Label
	for _, l := range m.store {
		if l.OwnerID == ownerID {
			out = append(out, *l)
		}
	}
	return out, nil
}
func (m *mockLabelRepo) FindByID(_ context.Context, id string) (*domain.Label, error) {
	l, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}
func (m *mockLabelRepo) FindByOwnerAndName(_ context.Context, ownerID, name string) (*domain.Label, error) {
	for _, l := range m.store {
		if l.OwnerID == ownerID && l.Name == name {
			return l, nil
		}
	}
	return nil, nil
}
func (m *mockLabelRepo) Create(_ context.Context, label *domain.Label) error {
	if label.ID == "" {
		label.ID = id.New()
	}
	cp := *label
	m.store[label.ID] = &cp
	m.created = &cp
	return nil
}
func (m *mockLabelRepo) Rename(_ context.Context, label *domain.Label, newName string) error {
	m.renamedTo = newName
	if l, ok := m.store[label.ID]; ok {
		l.Name = newName
	}
	return nil
}
func (m *mockLabelRepo) Delete(_ context.Context, label *domain.Label) error {
	delete(m.store, label.ID)
	m.deletedIDs = append(m.deletedIDs, label.ID)
	return nil
}

// --- ShoppingListRepository のモック ---

type mockShoppingListRepo struct {
	store           map[string]*domain.ShoppingList // id -> list
	created         *domain.ShoppingList
	addedItem       *domain.ShoppingListItem
	addedItems      []*domain.ShoppingListItem // AddItems に渡された項目を呼び出し順に記録
	checkedItems    map[string]bool            // itemID -> checked
	deletedItems    []string
	clearedLists    []string
	reorderedIDs    []string
	deletedOwnerIDs []string // DeleteByOwnerID に渡された ownerID を呼び出し順に記録
}

func newMockShoppingListRepo() *mockShoppingListRepo {
	return &mockShoppingListRepo{store: map[string]*domain.ShoppingList{}, checkedItems: map[string]bool{}}
}

func (m *mockShoppingListRepo) FindByOwnerID(_ context.Context, ownerID string) (*domain.ShoppingList, error) {
	for _, l := range m.store {
		if l.OwnerID == ownerID {
			return l, nil
		}
	}
	return nil, nil
}
func (m *mockShoppingListRepo) FindByID(_ context.Context, id string) (*domain.ShoppingList, error) {
	l, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}
func (m *mockShoppingListRepo) Create(_ context.Context, list *domain.ShoppingList) error {
	if list.ID == "" {
		list.ID = id.New()
	}
	cp := *list
	m.store[list.ID] = &cp
	m.created = &cp
	return nil
}
func (m *mockShoppingListRepo) AddItem(_ context.Context, item *domain.ShoppingListItem) error {
	if item.ID == "" {
		item.ID = id.New()
	}
	if l, ok := m.store[item.ShoppingListID]; ok {
		l.Items = append(l.Items, *item)
	}
	m.addedItem = item
	return nil
}
func (m *mockShoppingListRepo) AddItems(_ context.Context, items []*domain.ShoppingListItem) error {
	for _, item := range items {
		if item.ID == "" {
			item.ID = id.New()
		}
		if l, ok := m.store[item.ShoppingListID]; ok {
			l.Items = append(l.Items, *item)
		}
	}
	m.addedItems = append(m.addedItems, items...)
	return nil
}
func (m *mockShoppingListRepo) SetItemChecked(_ context.Context, itemID string, checked bool) error {
	m.checkedItems[itemID] = checked
	for _, l := range m.store {
		for i := range l.Items {
			if l.Items[i].ID == itemID {
				l.Items[i].Checked = checked
			}
		}
	}
	return nil
}
func (m *mockShoppingListRepo) DeleteItem(_ context.Context, itemID string) error {
	m.deletedItems = append(m.deletedItems, itemID)
	for _, l := range m.store {
		kept := l.Items[:0]
		for i := range l.Items {
			if l.Items[i].ID != itemID {
				kept = append(kept, l.Items[i])
			}
		}
		l.Items = kept
	}
	return nil
}
func (m *mockShoppingListRepo) Reorder(_ context.Context, _ string, itemIDs []string) error {
	m.reorderedIDs = itemIDs
	return nil
}
func (m *mockShoppingListRepo) DeleteCheckedItems(_ context.Context, listID string) error {
	m.clearedLists = append(m.clearedLists, listID)
	if l, ok := m.store[listID]; ok {
		kept := l.Items[:0]
		for i := range l.Items {
			if !l.Items[i].Checked {
				kept = append(kept, l.Items[i])
			}
		}
		l.Items = kept
	}
	return nil
}
func (m *mockShoppingListRepo) DeleteByOwnerID(_ context.Context, ownerID string) error {
	for lid, l := range m.store {
		if l.OwnerID == ownerID {
			delete(m.store, lid)
			m.deletedOwnerIDs = append(m.deletedOwnerIDs, ownerID)
		}
	}
	return nil
}

// --- ShareGroupRepository のモック ---

type mockShareGroupRepo struct {
	groups            map[string]*domain.ShareGroup // id -> group
	membership        map[string]string             // user_id -> group_id
	shareShoppingList map[string]bool               // user_id -> ShareShoppingList
	created           *domain.ShareGroup
	deletedIDs        []string
}

func newMockShareGroupRepo() *mockShareGroupRepo {
	return &mockShareGroupRepo{
		groups:            map[string]*domain.ShareGroup{},
		membership:        map[string]string{},
		shareShoppingList: map[string]bool{},
	}
}

// seed はテスト用に owner + メンバーを持つグループを 1 件登録する(全員 ShareShoppingList=true)。
func (m *mockShareGroupRepo) seed(groupID, ownerID string, memberIDs ...string) *domain.ShareGroup {
	members := []domain.User{{ID: ownerID}}
	m.membership[ownerID] = groupID
	m.shareShoppingList[ownerID] = true
	for _, uid := range memberIDs {
		members = append(members, domain.User{ID: uid})
		m.membership[uid] = groupID
		m.shareShoppingList[uid] = true
	}
	g := &domain.ShareGroup{ID: groupID, Name: "g", OwnerID: ownerID, Members: members}
	m.groups[groupID] = g
	return g
}

func (m *mockShareGroupRepo) Create(_ context.Context, g *domain.ShareGroup) error {
	if g.ID == "" {
		g.ID = id.New()
	}
	cp := *g
	cp.Members = []domain.User{{ID: g.OwnerID}}
	m.groups[g.ID] = &cp
	m.membership[g.OwnerID] = g.ID
	m.shareShoppingList[g.OwnerID] = true
	m.created = &cp
	return nil
}
func (m *mockShareGroupRepo) FindByUserID(_ context.Context, userID string) (*domain.ShareGroup, error) {
	gid, ok := m.membership[userID]
	if !ok {
		return nil, nil
	}
	return m.groups[gid], nil
}
func (m *mockShareGroupRepo) FindByID(_ context.Context, groupID string) (*domain.ShareGroup, error) {
	g, ok := m.groups[groupID]
	if !ok {
		return nil, nil
	}
	return g, nil
}
func (m *mockShareGroupRepo) FindByInviteCode(_ context.Context, code string) (*domain.ShareGroup, error) {
	for _, g := range m.groups {
		if g.InviteCode == code {
			return g, nil
		}
	}
	return nil, nil
}
func (m *mockShareGroupRepo) MemberIDs(_ context.Context, userID string) ([]string, error) {
	gid, ok := m.membership[userID]
	if !ok {
		return nil, nil
	}
	var ids []string
	for uid, g := range m.membership {
		if g == gid {
			ids = append(ids, uid)
		}
	}
	return ids, nil
}
func (m *mockShareGroupRepo) AddMember(_ context.Context, groupID, userID string, shareShoppingList bool) error {
	m.membership[userID] = groupID
	m.shareShoppingList[userID] = shareShoppingList
	if g, ok := m.groups[groupID]; ok {
		g.Members = append(g.Members, domain.User{ID: userID})
	}
	return nil
}
func (m *mockShareGroupRepo) RemoveMember(_ context.Context, groupID, userID string) error {
	delete(m.membership, userID)
	delete(m.shareShoppingList, userID)
	if g, ok := m.groups[groupID]; ok {
		kept := g.Members[:0]
		for _, u := range g.Members {
			if u.ID != userID {
				kept = append(kept, u)
			}
		}
		g.Members = kept
	}
	return nil
}
func (m *mockShareGroupRepo) UpdateInviteCode(_ context.Context, groupID, code string, _ time.Time) error {
	if g, ok := m.groups[groupID]; ok {
		g.InviteCode = code
	}
	return nil
}
func (m *mockShareGroupRepo) Delete(_ context.Context, groupID string) error {
	delete(m.groups, groupID)
	for uid, gid := range m.membership {
		if gid == groupID {
			delete(m.membership, uid)
			delete(m.shareShoppingList, uid)
		}
	}
	m.deletedIDs = append(m.deletedIDs, groupID)
	return nil
}
func (m *mockShareGroupRepo) FindMembership(_ context.Context, userID string) (*domain.ShareGroupMember, error) {
	gid, ok := m.membership[userID]
	if !ok {
		return nil, nil
	}
	return &domain.ShareGroupMember{GroupID: gid, UserID: userID, ShareShoppingList: m.shareShoppingList[userID]}, nil
}
func (m *mockShareGroupRepo) SharingMemberIDs(_ context.Context, groupID string) ([]string, error) {
	var ids []string
	for uid, gid := range m.membership {
		if gid == groupID && m.shareShoppingList[uid] {
			ids = append(ids, uid)
		}
	}
	return ids, nil
}
func (m *mockShareGroupRepo) UpdateShareShoppingList(_ context.Context, userID string, share bool) error {
	m.shareShoppingList[userID] = share
	return nil
}

// --- AvatarStorage のモック ---

type mockAvatarStorage struct {
	presignedKeys []string // PresignUpload に渡された key を呼び出し順に記録
	deletedKeys   []string // Delete に渡された key を呼び出し順に記録
	presignErr    error
}

func (m *mockAvatarStorage) PresignUpload(_ context.Context, key, _ string) (string, error) {
	if m.presignErr != nil {
		return "", m.presignErr
	}
	m.presignedKeys = append(m.presignedKeys, key)
	return "https://example.com/upload/" + key, nil
}
func (m *mockAvatarStorage) Delete(_ context.Context, key string) error {
	m.deletedKeys = append(m.deletedKeys, key)
	return nil
}
func (m *mockAvatarStorage) PublicURL(key string) string {
	return "https://example.com/public/" + key
}
