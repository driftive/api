package parsing

import (
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository/queries"
)

func ToGitRepositoryDTO(repository queries.GitRepository) dto.GitRepositoryDTO {
	return dto.GitRepositoryDTO{
		ID:               repository.ID,
		OrganizationID:   repository.OrganizationID,
		ProviderID:       repository.ProviderID,
		Name:             repository.Name,
		IsPrivate:        repository.IsPrivate,
		HasAnalysisToken: repository.AnalysisToken != nil,
	}
}

func ToGitRepositoryDTOs(repositories []queries.GitRepository) []dto.GitRepositoryDTO {
	repoDTOs := make([]dto.GitRepositoryDTO, 0)
	for _, repo := range repositories {
		repoDTOs = append(repoDTOs, ToGitRepositoryDTO(repo))
	}
	return repoDTOs
}
