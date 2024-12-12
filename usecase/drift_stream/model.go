package drift_stream

import "driftive.cloud/api/model"

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
