package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
	models "github.com/localrepo/api-server/internal/db"
	"golang.org/x/crypto/ssh"
)

// AddSSHKeyRequest holds data for adding a new SSH key
type AddSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
}

// AddSSHKey adds a new SSH public key to the authenticated user's account
func AddSSHKey(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req AddSSHKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	req.PublicKey = strings.TrimSpace(req.PublicKey)
	if req.Name == "" || req.PublicKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and PublicKey are required"})
	}

	// Parse and validate the SSH public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid SSH public key format"})
	}

	// Calculate fingerprint (SHA256)
	hash := sha256.Sum256(pubKey.Marshal())
	fingerprint := "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])

	sshKey := models.SSHKey{
		UserID:      userID.(uint),
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Fingerprint: fingerprint,
	}

	if err := models.DB.Create(&sshKey).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "This SSH key is already registered"})
	}

	return c.Status(fiber.StatusCreated).JSON(sshKey)
}

// ListSSHKeys returns all SSH keys for the authenticated user
func ListSSHKeys(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var keys []models.SSHKey
	if err := models.DB.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch SSH keys"})
	}

	return c.JSON(keys)
}

// DeleteSSHKey removes an SSH key
func DeleteSSHKey(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	keyID := c.Params("id")
	if keyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Key ID required"})
	}

	if err := models.DB.Where("id = ? AND user_id = ?", keyID, userID).Delete(&models.SSHKey{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete SSH key"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
