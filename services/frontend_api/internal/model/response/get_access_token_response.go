package response

type GetAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Success     bool   `json:"success"`
}
