package dto

type OrganizationDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	AvatarURL string `json:"avatar_url"`
}
