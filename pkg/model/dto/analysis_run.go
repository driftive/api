package dto

import "time"

type DriftAnalysisRunDTO struct {
	Uuid                 string    `json:"uuid"`
	RepositoryId         int64     `json:"repository_id"`
	TotalProjects        int32     `json:"total_projects"`
	TotalProjectsDrifted int32     `json:"total_projects_drifted"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
