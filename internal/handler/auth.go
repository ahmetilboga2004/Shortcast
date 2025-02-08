package handler

import (
	"shortcast/internal/dto"
	"shortcast/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user with email/username and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			login	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	map[string]string	"token"
//	@Failure		400		{object}	map[string]string	"error"
//	@Failure		401		{object}	map[string]string	"error"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request dto.LoginRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz istek verisi",
		})
	}

	token, err := h.authService.Login(request.EmailOrUsername, request.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

// Register godoc
//
//	@Summary		Register new user
//	@Description	Create a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		dto.RegisterRequest	true	"Registration details"
//	@Success		201			{object}	map[string]string	"message"
//	@Failure		400			{object}	map[string]string	"error"
//	@Failure		409			{object}	map[string]string	"error"
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var request dto.RegisterRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz istek verisi",
		})
	}

	err := h.authService.Register(request)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Kullanıcı zaten mevcut veya bir hata oluştu",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Kullanıcı başarıyla kayıt edildi",
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Invalidate user's JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "message"
// @Failure      401  {object}  map[string]string  "error"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)

	err := h.authService.Logout(token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Çıkış yapılırken bir hata oluştu",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Başarıyla çıkış yapıldı",
	})
}
