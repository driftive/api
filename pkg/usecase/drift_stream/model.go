package drift_stream

import (
	"driftive.cloud/api/pkg/model"
	"time"
)

type ProjectType int

const (
	Terraform ProjectType = iota
	Tofu
	Terragrunt
)

type ProjectDriftAnalysisState struct {
	Project    model.Project
	Drifted    bool
	Succeeded  bool
	PlanOutput string
	Adds       int
	Changes    int
	Destroys   int
}

type DriftAnalysisState struct {
	RunID         string
	ProjectStates []ProjectDriftAnalysisState
}

// TypedProject represents a TF/Tofu/Terragrunt project to be analyzed
type TypedProject struct {
	Dir  string      `json:"dir" yaml:"dir"`
	Type ProjectType `json:"type" yaml:"type"`
}

type DriftProjectResult struct {
	Project TypedProject `json:"project"`
	Drifted bool         `json:"drifted"`
	// Succeeded true if the drift analysis succeeded, even if the project had drifted.
	Succeeded  bool   `json:"succeeded"`
	InitOutput string `json:"init_output"`
	PlanOutput string `json:"plan_output"`
}

type DriftDetectionResult struct {
	ProjectResults []DriftProjectResult `json:"project_results"`
	TotalDrifted   int32                `json:"total_drifted"`
	TotalErrored   *int32               `json:"total_errored,omitempty"`
	TotalProjects  int32                `json:"total_projects"`
	TotalChecked   int32                `json:"total_checked"`
	Duration       time.Duration        `json:"duration"`
}
