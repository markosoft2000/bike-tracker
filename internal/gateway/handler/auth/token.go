package auth_handler

import (
	"context"

	"github.com/gofiber/fiber/v3"

	authv1 "github.com/markosoft2000/bike-tracker/pkg/gen/grpc/auth/sso"
)

func (h *authHandler) Refresh(c fiber.Ctx) error {
	var req authv1.RefreshTokenRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request JSON payload",
			},
		)
	}

	req.Ip = c.IP()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.RefreshTokenTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.RefreshToken(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
