package auth_handler

import (
	"context"

	"github.com/gofiber/fiber/v3"

	authv1 "github.com/markosoft2000/bike-tracker/pkg/gen/grpc/auth/sso"
)

func (h *authHandler) Register(c fiber.Ctx) error {
	var req authv1.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request JSON payload"},
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.UserRegisterTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.Register(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *authHandler) Login(c fiber.Ctx) error {
	var req authv1.LoginRequest
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
		h.cfg.Services.SLO.Auth.UserLoginTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.Login(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *authHandler) Logout(c fiber.Ctx) error {
	var req authv1.LogoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request JSON payload",
			},
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.UserLogoutTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.Logout(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *authHandler) IsAdmin(c fiber.Ctx) error {
	req := authv1.IsAdminRequest{
		UserId: c.Params("userId"),
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		h.cfg.Services.SLO.Auth.UserIsAdminTimeout,
	)
	defer cancel()

	resp, err := h.gRPCClient.IsAdmin(ctx, &req)
	if err != nil {
		return h.handleGrpcError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
