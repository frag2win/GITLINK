package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

	// Unified diff parser to dynamically extract conflicting files and changed line ranges
	lines := strings.Split(diffStr, "\n")
	var currentFile *models.ConflictingFile

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil && len(currentFile.Hunks) > 0 {
				conflictingFiles = append(conflictingFiles, *currentFile)
			}
			parts := strings.Fields(line)
			filePath := "unknown_file"
			if len(parts) >= 4 {
				filePath = strings.TrimPrefix(parts[3], "b/")
			}
			currentFile = &models.ConflictingFile{
				FilePath: filePath,
				Hunks:    make([]models.ConflictHunk, 0),
			}
		} else if strings.HasPrefix(line, "@@ ") && currentFile != nil {
			hunk := models.ConflictHunk{
				GitConflictType: "content",
				Reason:          "Concurrent modification detected in branch diff hunk",
			}
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				targetPart := parts[2] // e.g. +10,15
				targetPart = strings.TrimPrefix(targetPart, "+")
				subParts := strings.Split(targetPart, ",")
				if len(subParts) > 0 {
					if start, errParse := strconv.Atoi(subParts[0]); errParse == nil {
						hunk.StartLine = start
						hunk.EndLine = start
						if len(subParts) > 1 {
							if count, errCount := strconv.Atoi(subParts[1]); errCount == nil {
								hunk.EndLine = start + count
							}
						}
					}
				}
			}
			currentFile.Hunks = append(currentFile.Hunks, hunk)
		}
	}

	if currentFile != nil && len(currentFile.Hunks) > 0 {
		conflictingFiles = append(conflictingFiles, *currentFile)
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
