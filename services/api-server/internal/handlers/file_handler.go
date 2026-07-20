package handlers

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

type FileHandler struct {
	gitService       service.GitService
	repoService      service.RepoService
	authorizationSvc service.AuthorizationService
}

func NewFileHandler(gitService service.GitService, repoService service.RepoService, authorizationSvc service.AuthorizationService) *FileHandler {
	return &FileHandler{
		gitService:       gitService,
		repoService:      repoService,
		authorizationSvc: authorizationSvc,
	}
}

type FileEntryDTO struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Type      string `json:"type"` // "file" or "directory"
	SizeBytes int64  `json:"sizeBytes"`
}

type FileContentResponseDTO struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
	Content   string `json:"content"`
	IsBinary  bool   `json:"isBinary"`
	Ref       string `json:"ref"`
}

func (h *FileHandler) getAuthorizedRepo(c *fiber.Ctx) (string, error) {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return "", c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repoIDStr := c.Params("id")
	repoID, err := strconv.ParseUint(repoIDStr, 10, 32)
	if err != nil {
		return "", c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid repository ID"})
	}

	repo, err := h.repoService.GetRepoByID(c.Context(), uint(repoID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Repository not found"})
		}
		return "", c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	isAllowed := false
	var authErr error
	if c.Method() == fiber.MethodGet {
		isAllowed, authErr = h.authorizationSvc.AuthorizeRead(c.Context(), userID, repo.ID)
	} else {
		isAllowed, authErr = h.authorizationSvc.AuthorizeWrite(c.Context(), userID, repo.ID)
	}

	if authErr != nil || !isAllowed {
		return "", c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Access denied"})
	}

	return repo.Name, nil
}

func (h *FileHandler) BrowseFiles(c *fiber.Ctx) error {
	repoName, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	ref := c.Query("ref", "main")
	// The path can be in wildcard * or in the query param
	path := c.Params("*")
	if path == "" {
		path = c.Query("path", "")
	}
	path = strings.Trim(path, "/")

	// Try to get tree entries
	entries, err := h.gitService.GetTree(c.Context(), repoName, ref, path)
	if err != nil {
		// If it failed because it is a file rather than a directory, serve file content!
		if strings.Contains(err.Error(), "not a tree") || strings.Contains(err.Error(), "ObjectNotFound") {
			return h.serveFile(c, repoName, ref, path)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Map to frontend DTO
	dtoList := make([]FileEntryDTO, len(entries))
	for i, entry := range entries {
		entryType := "file"
		if entry.Type == "tree" {
			entryType = "directory"
		}
		dtoList[i] = FileEntryDTO{
			Path:      entry.Path,
			Name:      entry.Name,
			Type:      entryType,
			SizeBytes: entry.Size,
		}
	}

	return c.JSON(dtoList)
}

func (h *FileHandler) GetFileContent(c *fiber.Ctx) error {
	repoName, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	ref := c.Query("ref", "main")
	path := c.Params("*")
	path = strings.Trim(path, "/")

	return h.serveFile(c, repoName, ref, path)
}

func (h *FileHandler) serveFile(c *fiber.Ctx, repoName, ref, path string) error {
	content, err := h.gitService.GetFile(c.Context(), repoName, ref, path)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	// Determine MIME type
	name := filepath.Base(path)
	mimeType := "text/plain"
	switch {
	case strings.HasSuffix(name, ".png"):
		mimeType = "image/png"
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		mimeType = "image/jpeg"
	case strings.HasSuffix(name, ".gif"):
		mimeType = "image/gif"
	case strings.HasSuffix(name, ".svg"):
		mimeType = "image/svg+xml"
	case strings.HasSuffix(name, ".pdf"):
		mimeType = "application/pdf"
	case strings.HasSuffix(name, ".json"):
		mimeType = "application/json"
	}

	// Determine if binary (null byte heuristic)
	isBinary := false
	for _, b := range content {
		if b == 0 {
			isBinary = true
			break
		}
	}

	// Binary content is base64-encoded to produce valid JSON UTF-8.
	var contentStr string
	if isBinary {
		contentStr = base64.StdEncoding.EncodeToString(content)
	} else {
		contentStr = string(content)
	}

	resp := FileContentResponseDTO{
		Name:      name,
		Path:      path,
		MimeType:  mimeType,
		SizeBytes: int64(len(content)),
		Content:   contentStr,
		IsBinary:  isBinary,
		Ref:       ref,
	}

	return c.JSON(resp)
}
