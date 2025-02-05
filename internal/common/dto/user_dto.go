package dto

type UserLoginRequestDTO struct {
	Login    string `json:"login" example:"login"`
	Password string `json:"password" example:"password"`
}

type CreateUserRequestDTO struct {
	UserLoginRequestDTO
}

type CreateUserDTO struct {
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}

type UserDTO struct {
	Login string `json:"login" example:"login"`
	ID    string `json:"id" example:"id-1"`
}
