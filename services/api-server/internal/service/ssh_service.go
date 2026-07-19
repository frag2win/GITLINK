package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"golang.org/x/crypto/ssh"
)

type SSHService interface {
	AddSSHKey(userID uint, name, publicKey string) (*models.SSHKey, error)
	ListSSHKeys(userID uint) ([]models.SSHKey, error)
	DeleteSSHKey(keyID string, userID uint) error
	AuthenticateByKey(key ssh.PublicKey) (*models.SSHKey, error)
}

type sshService struct {
	sshRepo repository.SSHRepository
}

func NewSSHService(sshRepo repository.SSHRepository) SSHService {
	return &sshService{sshRepo: sshRepo}
}

func (s *sshService) AddSSHKey(userID uint, name, publicKey string) (*models.SSHKey, error) {
	publicKey = strings.TrimSpace(publicKey)
	if name == "" || publicKey == "" {
		return nil, errors.New("name and publicKey are required")
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return nil, errors.New("invalid SSH public key format")
	}

	hash := sha256.Sum256(pubKey.Marshal())
	fingerprint := "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])

	sshKey := &models.SSHKey{
		UserID:      userID,
		Name:        name,
		PublicKey:   publicKey,
		Fingerprint: fingerprint,
	}

	if err := s.sshRepo.CreateSSHKey(sshKey); err != nil {
		return nil, errors.New("this SSH key is already registered")
	}

	return sshKey, nil
}

func (s *sshService) ListSSHKeys(userID uint) ([]models.SSHKey, error) {
	return s.sshRepo.FindByUserID(userID)
}

func (s *sshService) DeleteSSHKey(keyID string, userID uint) error {
	if keyID == "" {
		return errors.New("key ID required")
	}
	return s.sshRepo.DeleteByIDAndUserID(keyID, userID)
}

func (s *sshService) AuthenticateByKey(key ssh.PublicKey) (*models.SSHKey, error) {
	hash := sha256.Sum256(key.Marshal())
	fingerprint := "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])

	sshKey, err := s.sshRepo.FindByFingerprint(fingerprint)
	if err != nil {
		return nil, errors.New("invalid SSH key")
	}
	return sshKey, nil
}
