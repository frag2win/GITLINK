package repository

import (
	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type SSHRepository interface {
	CreateSSHKey(key *models.SSHKey) error
	FindByUserID(userID uint) ([]models.SSHKey, error)
	FindByFingerprint(fingerprint string) (*models.SSHKey, error)
	DeleteByIDAndUserID(id string, userID uint) error
}

type sshRepository struct {
	db *gorm.DB
}

func NewSSHRepository(db *gorm.DB) SSHRepository {
	return &sshRepository{db: db}
}

func (r *sshRepository) CreateSSHKey(key *models.SSHKey) error {
	return r.db.Create(key).Error
}

func (r *sshRepository) FindByUserID(userID uint) ([]models.SSHKey, error) {
	var keys []models.SSHKey
	if err := r.db.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *sshRepository) FindByFingerprint(fingerprint string) (*models.SSHKey, error) {
	var key models.SSHKey
	if err := r.db.Where("fingerprint = ?", fingerprint).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *sshRepository) DeleteByIDAndUserID(id string, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SSHKey{}).Error
}
