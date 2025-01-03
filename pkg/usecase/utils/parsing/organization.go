package parsing

import (
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository/queries"
)

func ToOrganizationDTO(organization queries.GitOrganization) dto.OrganizationDTO {
	return dto.OrganizationDTO{
		ID:   organization.ID,
		Name: organization.Name,
	}
}

func ToOrganizationDTOs(organizations []queries.GitOrganization) []dto.OrganizationDTO {
	orgDTOs := make([]dto.OrganizationDTO, 0)
	for _, org := range organizations {
		orgDTOs = append(orgDTOs, ToOrganizationDTO(org))
	}
	return orgDTOs
}
