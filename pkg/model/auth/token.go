package auth

type UserToken struct {
	ID       int64  `json:"id"`
	Provider string `json:"provider"`
}
