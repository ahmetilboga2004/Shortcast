package middleware

import (
	"shortcast/internal/config"
	"shortcast/internal/repository"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	cfg      *config.Config
	authRepo *repository.AuthRepository
}

func NewAuthMiddleware(cfg *config.Config, authRepo *repository.AuthRepository) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:      cfg,
		authRepo: authRepo,
	}
}

func (am *AuthMiddleware) JWTMiddleware() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(am.cfg.SecretKey)},
		SuccessHandler: func(c *fiber.Ctx) error {
			// Token'ı kontrol et
			token := c.Get("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			// Token blacklist'te mi kontrol et
			exists, err := am.authRepo.IsTokenBlacklisted(token)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Token kontrolü yapılamadı",
				})
			}

			if exists {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Geçersiz token",
				})
			}

			return c.Next()
		},
	})
}

func (am *AuthMiddleware) GuestMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")

		if tokenString == "" {
			return c.Next()
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Önce blacklist kontrolü yapalım
		exists, err := am.authRepo.IsTokenBlacklisted(tokenString)
		if err == nil && exists {
			return c.Next() // Token blacklist'te ise guest olarak kabul et
		}

		// Token'ı doğrula
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(am.cfg.SecretKey), nil
		})

		if err != nil || !token.Valid {
			return c.Next() // Geçersiz token ise guest olarak kabul et
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Giriş yapmış kullanıcılar bu API'yi kullanamaz!",
		})
	}
}
