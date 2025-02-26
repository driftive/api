package dto

import "time"

type DriftAnalysisRunDTO struct {
	Uuid                 string    `json:"uuid"`
	RepositoryId         int64     `json:"repository_id"`
	TotalProjects        int32     `json:"total_projects"`
	TotalProjectsDrifted int32     `json:"total_projects_drifted"`
	DurationMillis       int64     `json:"duration_millis"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type DriftAnalysisProjectDTO struct {
	Id         int64   `json:"id"`
	RunId      string  `json:"run_id"`
	Dir        string  `json:"dir"`
	Type       string  `json:"type"`
	Drifted    bool    `json:"drifted"`
	Succeeded  bool    `json:"succeeded"`
	InitOutput *string `json:"init_output"`
	PlanOutput *string `json:"plan_output"`
}

type DriftAnalysisRunWithProjectsDTO struct {
	DriftAnalysisRunDTO
	Projects []DriftAnalysisProjectDTO `json:"projects"`
}
