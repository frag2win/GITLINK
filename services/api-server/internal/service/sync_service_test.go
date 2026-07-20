package service_test

import (
	"testing"
	"time"

	"github.com/localrepo/api-server/internal/service"
)

func TestExponentialBackoffCalculation(t *testing.T) {
	svc := service.NewSyncService(nil, nil, nil)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 30 * time.Second},
		{2, 2 * time.Minute},
		{3, 10 * time.Minute},
		{4, 1 * time.Hour},
		{5, 24 * time.Hour},
	}

	for _, tt := range tests {
		got := svc.CalculateNextBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("CalculateNextBackoff(%d) = %v; want %v", tt.attempt, got, tt.expected)
		}
	}
}
