package service_test

import (
	"context"
	"testing"

	"github.com/localrepo/api-server/internal/service"
)

func TestConflictServiceAnalysis(t *testing.T) {
	gitSvc := &mockGitService{}
	conflictSvc := service.NewConflictService(gitSvc)

	report, err := conflictSvc.AnalyzeConflicts(context.Background(), 1, "testrepo", "main", "feature")
	if err != nil {
		t.Fatalf("expected conflict analysis to succeed, got error: %v", err)
	}

	if report.AnalysisVersion != "v1.0" {
		t.Errorf("expected AnalysisVersion v1.0, got %s", report.AnalysisVersion)
	}

	if report.MergeBaseSHA == "" {
		t.Errorf("expected MergeBaseSHA to be populated")
	}

	if len(report.ConflictingFiles) == 0 {
		t.Fatalf("expected conflicting files diagnostic data to be present")
	}

	cf := report.ConflictingFiles[0]
	if cf.FilePath != "src/main.go" {
		t.Errorf("expected parsed file path 'src/main.go', got %s", cf.FilePath)
	}

	if len(cf.Hunks) == 0 {
		t.Fatalf("expected hunks to be parsed from diff")
	}

	hunk := cf.Hunks[0]
	if hunk.StartLine != 10 {
		t.Errorf("expected parsed StartLine 10, got %d", hunk.StartLine)
	}

	if hunk.EndLine != 15 {
		t.Errorf("expected parsed EndLine 15, got %d", hunk.EndLine)
	}
}
