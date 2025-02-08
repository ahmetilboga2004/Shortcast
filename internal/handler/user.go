package handler

import (
	"shortcast/internal/dto"
	"shortcast/internal/service"
	"shortcast/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserService(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetByID godoc
//
//	@Summary		Get user by ID
//	@Description	Get user details by user ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	dto.UserResponse
//	@Failure		400	{object}	map[string]string	"error"
//	@Failure		404	{object}	map[string]string	"error"
//	@Router			/users/{id} [get]
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user, err := h.userService.GetUserById(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userDto := &dto.UserResponse{
		ID:        user.ID,
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": userDto,
	})
}
