package utils

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ParamAsUint(c *fiber.Ctx, param string) (uint, error) {
	paramValue := c.Params(param)

	parsedValue, err := strconv.ParseUint(paramValue, 10, 32)
	if err != nil {
		return 0, errors.New("geçersiz ID parametresi")
	}

	return uint(parsedValue), nil
}
