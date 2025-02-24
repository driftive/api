package dto

type GitRepositoryDTO struct {
	ID               int64  `json:"id"`
	OrganizationID   int64  `json:"organization_id"`
	ProviderID       string `json:"provider_id"`
	Name             string `json:"name"`
	IsPrivate        bool   `json:"is_private"`
	HasAnalysisToken bool   `json:"has_analysis_token"`
}
