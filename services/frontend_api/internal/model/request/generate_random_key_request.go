package request

type GenerateRandomKeyRequest struct {
	Length int `form:"length" binding:"required,min=16,max=64"`
}
