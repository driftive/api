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
		CreatedAt:            run.CreatedAt,
		UpdatedAt:            run.UpdatedAt,
	}
}

func ToDriftAnalysisRunDTOs(runs []queries.DriftAnalysisRun) []dto.DriftAnalysisRunDTO {
	var dtos []dto.DriftAnalysisRunDTO
	for _, run := range runs {
		dtos = append(dtos, ToDriftAnalysisRunDTO(run))
	}
	return dtos
}
