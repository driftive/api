package parsing

import (
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository/queries"
)

func ToDriftAnalysisRunDTO(run queries.DriftAnalysisRun) dto.DriftAnalysisRunDTO {
	return dto.DriftAnalysisRunDTO{
		Uuid:                 run.Uuid.String(),
		RepositoryId:         run.RepositoryID,
		TotalProjects:        run.TotalProjects,
		TotalProjectsDrifted: run.TotalProjectsDrifted,
		DurationMillis:       run.AnalysisDurationMillis,
		CreatedAt:            run.CreatedAt,
		UpdatedAt:            run.UpdatedAt,
	}
}

func ToDriftAnalysisRunDTOs(runs []queries.DriftAnalysisRun) []dto.DriftAnalysisRunDTO {
	dtos := make([]dto.DriftAnalysisRunDTO, 0)
	for _, run := range runs {
		dtos = append(dtos, ToDriftAnalysisRunDTO(run))
	}
	return dtos
}

func ToDriftAnalysisProjectDTO(project queries.DriftAnalysisProject) dto.DriftAnalysisProjectDTO {
	return dto.DriftAnalysisProjectDTO{
		Id:         project.ID,
		RunId:      project.DriftAnalysisRunID.String(),
		Dir:        project.Dir,
		Type:       project.Type,
		Drifted:    project.Drifted,
		Succeeded:  project.Succeeded,
		InitOutput: project.InitOutput,
		PlanOutput: project.PlanOutput,
	}
}

func ToDriftAnalysisProjectDTOs(projects []queries.DriftAnalysisProject) []dto.DriftAnalysisProjectDTO {
	dtos := make([]dto.DriftAnalysisProjectDTO, 0)
	for _, project := range projects {
		dtos = append(dtos, ToDriftAnalysisProjectDTO(project))
	}
	return dtos
}

func ToDriftAnalysisRunWithProjectsDTO(run queries.DriftAnalysisRun, projects []queries.DriftAnalysisProject) dto.DriftAnalysisRunWithProjectsDTO {
	return dto.DriftAnalysisRunWithProjectsDTO{
		DriftAnalysisRunDTO: ToDriftAnalysisRunDTO(run),
		Projects:            ToDriftAnalysisProjectDTOs(projects),
	}
}
