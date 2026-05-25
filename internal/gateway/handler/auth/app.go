package auth_handler

import (
	"context"
	"encoding/base64"

	"github.com/gofiber/fiber/v3"

	authv1 "github.com/markosoft2000/bike-tracker/pkg/gen/grpc/auth/sso"
)

func (h *authHandler) AddApp(c fiber.Ctx) error {
	// anonymous JSON struct to accept standard Base64 string for proto raw bytes
	var payload struct {
		Name   string `json:"name"`
		Secret string `json:"secret"` // incoming base64 encoded string representation
	}

	if err := c.Bind().JSON(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request JSON payload",
			},
		)
	}

	// decode Base64 string into native []byte raw array slice
	rawSecret, err := base64.StdEncoding.DecodeString(payload.Secret)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid secret: value must be a valid base64 encoded string",
			},
		)
	}

	req := authv1.AddAppRequest{
		Name:   payload.Name,
		Secret: rawSecret,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.AppAddTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.AddApp(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *authHandler) RemoveApp(c fiber.Ctx) error {
	req := authv1.RemoveAppRequest{
		Id: c.Params("id"),
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.AppRemoveTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.RemoveApp(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
