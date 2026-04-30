package store

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// InviteCodeStore manages one-time invite codes.
type InviteCodeStore struct {
	db *gorm.DB
}

// InviteCode can be used once for registration.
type InviteCode struct {
	Code      string     `gorm:"primaryKey;size:64" json:"code"`
	CreatedBy string     `gorm:"column:created_by;size:64" json:"created_by"`
	UsedBy    string     `gorm:"column:used_by;size:64" json:"used_by"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (InviteCode) TableName() string { return "invite_codes" }

func NewInviteCodeStore(db *gorm.DB) *InviteCodeStore {
	return &InviteCodeStore{db: db}
}

func (s *InviteCodeStore) initTables() error {
	return s.db.AutoMigrate(&InviteCode{})
}

func normalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (s *InviteCodeStore) Create(code, createdBy string) error {
	row := &InviteCode{Code: normalizeInviteCode(code), CreatedBy: createdBy}
	return s.db.Create(row).Error
}

func (s *InviteCodeStore) IsUsable(code string) (bool, error) {
	var count int64
	err := s.db.Model(&InviteCode{}).
		Where("code = ? AND used_at IS NULL", normalizeInviteCode(code)).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Consume atomically marks invite code as used. Returns false when invalid/used.
func (s *InviteCodeStore) Consume(code, userID string) (bool, error) {
	now := time.Now().UTC()
	res := s.db.Model(&InviteCode{}).
		Where("code = ? AND used_at IS NULL", normalizeInviteCode(code)).
		Updates(map[string]interface{}{"used_by": userID, "used_at": now, "updated_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (s *InviteCodeStore) List(limit int) ([]InviteCode, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	var rows []InviteCode
	err := s.db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
