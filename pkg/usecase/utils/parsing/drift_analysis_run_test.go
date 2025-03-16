package parsing

import (
	"driftive.cloud/api/pkg/repository/queries"
	"github.com/google/uuid"
	"testing"
	"time"
)

// test
func TestToDriftAnalysisRunDTO(t *testing.T) {
	var runs = []queries.DriftAnalysisRun{
		{
			Uuid:                   uuid.New(),
			RepositoryID:           0,
			TotalProjects:          0,
			TotalProjectsDrifted:   0,
			AnalysisDurationMillis: 0,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		},
	}

	dtos := ToDriftAnalysisRunDTOs(runs)
	if len(dtos) != 1 {
		t.Errorf("Expected 1, got %d", len(dtos))
	}
}
