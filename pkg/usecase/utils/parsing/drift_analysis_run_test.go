package parsing

import (
	"driftive.cloud/api/pkg/repository/queries"
	"encoding/json"
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

// The UI decides whether to poll off `status`, and renders the in-flight project tags off
// `running_projects`. Both are mapped by hand, so a forgotten field fails silently: the JSON stays
// valid, the page still renders, and live runs simply never update. These assert the serialized
// wire format rather than the struct fields, which is what the UI actually consumes.
func TestToDriftAnalysisRunWithProjectsDTO_SerializesLiveFields(t *testing.T) {
	run := queries.DriftAnalysisRun{
		Uuid:            uuid.New(),
		RepositoryID:    7,
		TotalProjects:   3,
		Status:          "RUNNING",
		RunningProjects: []string{"infra/a", "infra/b"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	body, err := json.Marshal(ToDriftAnalysisRunWithProjectsDTO(run, nil))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["status"] != "RUNNING" {
		t.Errorf("status = %v, want RUNNING — the UI polls off this field", got["status"])
	}

	running, ok := got["running_projects"].([]any)
	if !ok {
		t.Fatalf("running_projects missing or not an array: %v", got["running_projects"])
	}
	if len(running) != 2 || running[0] != "infra/a" || running[1] != "infra/b" {
		t.Errorf("running_projects = %v, want [infra/a infra/b]", running)
	}
}

// An empty TEXT[] must serialize as [] rather than null, so the UI can map over it unconditionally.
func TestToDriftAnalysisRunWithProjectsDTO_EmptyRunningProjectsIsArray(t *testing.T) {
	cases := map[string][]string{
		"empty slice": {},
		"nil slice":   nil,
	}
	for name, runningProjects := range cases {
		t.Run(name, func(t *testing.T) {
			run := queries.DriftAnalysisRun{
				Uuid:            uuid.New(),
				Status:          "COMPLETED",
				RunningProjects: runningProjects,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			body, err := json.Marshal(ToDriftAnalysisRunWithProjectsDTO(run, nil))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got struct {
				RunningProjects *[]string `json:"running_projects"`
			}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.RunningProjects == nil {
				t.Errorf("running_projects serialized as null; the UI would crash mapping over it: %s", body)
			}
		})
	}
}

// The runs list is rendered from DriftAnalysisRunDTO, which needs status to show the Running tag.
func TestToDriftAnalysisRunDTO_IncludesStatus(t *testing.T) {
	body, err := json.Marshal(ToDriftAnalysisRunDTO(queries.DriftAnalysisRun{
		Uuid:      uuid.New(),
		Status:    "RUNNING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != "RUNNING" {
		t.Errorf("status = %q, want RUNNING", got.Status)
	}
}
