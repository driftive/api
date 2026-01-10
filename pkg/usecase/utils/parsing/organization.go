package parsing

import (
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/strutils"
)

func ToOrganizationDTO(organization queries.GitOrganization) dto.OrganizationDTO {
	return dto.OrganizationDTO{
		ID:        organization.ID,
		Name:      organization.Name,
		Installed: organization.InstallationID != nil,
		AvatarURL: strutils.OrEmpty(organization.AvatarUrl),
	}
}

func ToOrganizationDTOs(organizations []queries.GitOrganization) []dto.OrganizationDTO {
	orgDTOs := make([]dto.OrganizationDTO, 0, len(organizations))
	for _, org := range organizations {
		orgDTOs = append(orgDTOs, ToOrganizationDTO(org))
	}
	return orgDTOs
}
