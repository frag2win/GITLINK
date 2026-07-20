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
		t.Errorf("expected conflicting files diagnostic data to be present")
	}
}
