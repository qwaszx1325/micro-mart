package request

type RegisterRequest struct {
	UserName string `json:"username" binding:"required,min=3,max=50,one_num"`
	Email    string `json:"email" binding:"required_without=MobileNumber,omitempty,email"`
	Password string `json:"password" binding:"required,min=6,max=15"`
}
