package outputs

import "back/domain"

type UserResponse struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Activo   bool     `json:"activo"`
	Permisos []string `json:"permisos"`
}

func ToUserResponse(u *domain.User) UserResponse {
	if u.Permisos == nil {
		u.Permisos = []string{}
	}
	return UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Activo:   u.Activo,
		Permisos: u.Permisos,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
