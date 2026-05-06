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
	Code         string     `gorm:"primaryKey;size:64" json:"code"`
	CreatedBy    string     `gorm:"column:created_by;size:64" json:"created_by"`
	DurationDays int        `gorm:"column:duration_days;default:0" json:"duration_days"`
	UsedBy       string     `gorm:"column:used_by;size:64" json:"used_by"`
	UsedAt       *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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

func (s *InviteCodeStore) Create(code, createdBy string, durationDays int) error {
	if durationDays < 0 {
		durationDays = 0
	}
	row := &InviteCode{Code: normalizeInviteCode(code), CreatedBy: createdBy, DurationDays: durationDays}
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

// Consume atomically marks invite code as used and returns consumed row details.
func (s *InviteCodeStore) Consume(code, userID string) (*InviteCode, bool, error) {
	norm := normalizeInviteCode(code)
	now := time.Now().UTC()
	updates := map[string]interface{}{"used_by": userID, "used_at": now, "updated_at": now}
	res := s.db.Model(&InviteCode{}).
		Where("code = ? AND used_at IS NULL", norm).
		Updates(updates)
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, false, nil
	}
	var row InviteCode
	if err := s.db.Where("code = ?", norm).First(&row).Error; err != nil {
		return nil, false, err
	}
	return &row, true, nil
}

func (s *InviteCodeStore) List(limit int) ([]InviteCode, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	var rows []InviteCode
	err := s.db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
