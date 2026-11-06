// Package mapper преобразует внутренние модели в HTTP response DTO.
package mapper

import (
	"github.com/rinatkh/classroom15/internal/httpapi/dto"
	"github.com/rinatkh/classroom15/internal/user"
)

func UserToResponse(value user.User) dto.UserResponse {
	return dto.UserResponse{ID: value.ID, Email: value.Email}
}
