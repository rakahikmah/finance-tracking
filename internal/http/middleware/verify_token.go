package middleware

import (
	"github.com/gofiber/fiber/v2"
	apperr "github.com/rakahikmah/finance-tracking/error"
	"github.com/rakahikmah/finance-tracking/internal/http/auth"
)

func VerifyJWTToken(c *fiber.Ctx) error {
	if err := auth.VerifyToken(c); err != nil {
		return c.Status(apperr.ErrInvalidToken().HTTPCode).JSON(apperr.ErrInvalidToken())
	}

	return c.Next()
}
