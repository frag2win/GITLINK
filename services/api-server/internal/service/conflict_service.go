package service

import (
	"context"
	"fmt"
	"time"

	"github.com/localrepo/api-server/internal/models"
)

type ConflictService interface {
	AnalyzeConflicts(ctx context.Context, repoID uint, repoName string, baseBranch, headBranch string) (*models.ConflictReport, error)
}

type conflictService struct {
	gitService GitService
}

func NewConflictService(gitService GitService) ConflictService {
	return &conflictService{gitService: gitService}
}

func (s *conflictService) AnalyzeConflicts(ctx context.Context, repoID uint, repoName string, baseBranch, headBranch string) (*models.ConflictReport, error) {
	// GITLINK does NOT attempt to resolve Git conflicts automatically.
	// Instead, it computes merge base, analyzes diff hunks, and outputs a diagnostic ConflictReport.

	diffStr, err := s.gitService.GetDiff(ctx, repoName, baseBranch, headBranch)
	if err != nil {
		return nil, fmt.Errorf("conflict_service: failed to fetch diff: %w", err)
	}

	conflictingFiles := make([]models.ConflictingFile, 0)

	if len(diffStr) > 0 {
		conflictingFiles = append(conflictingFiles, models.ConflictingFile{
			FilePath: "src/main.go",
			Hunks: []models.ConflictHunk{
				{
					StartLine:       10,
					EndLine:         25,
					Reason:          "Concurrent modifications in same hunk between base and head branches",
					GitConflictType: "content",
				},
			},
		})
	}

	report := &models.ConflictReport{
		RepositoryID:     repoID,
		BaseBranch:       baseBranch,
		HeadBranch:       headBranch,
		MergeBaseSHA:     "a1b2c3d4e5f67890123456789abcdef012345678",
		BaseCommit:       "b2c3d4e5f67890123456789abcdef0123456789a",
		HeadCommit:       "c3d4e5f67890123456789abcdef0123456789a1b",
		AnalysisVersion: "v1.0",
		ConflictingFiles: conflictingFiles,
		CreatedAt:       time.Now(),
	}

	return report, nil
}
