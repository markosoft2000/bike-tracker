package auth_handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/markosoft2000/bike-tracker/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authv1 "github.com/markosoft2000/bike-tracker/pkg/gen/grpc/auth/sso"
)

type AuthHandlerService interface {
	Close() error // Close auth service (grpc client connections, etc...)

	// handlers
	Register(c fiber.Ctx) error
	Login(c fiber.Ctx) error
	Refresh(c fiber.Ctx) error
	Logout(c fiber.Ctx) error
	IsAdmin(c fiber.Ctx) error
	AddApp(c fiber.Ctx) error
	RemoveApp(c fiber.Ctx) error
}

type authHandler struct {
	log        *slog.Logger
	cfg        *config.Config
	gRPCClient authv1.AuthClient
	conn       *grpc.ClientConn
}

// NewAuthHandler initializes a long-lived connection pool to the backend gRPC Auth service
func NewAuthHandler(log *slog.Logger, cfg *config.Config) AuthHandlerService {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(cfg.Services.AuthServiceAddr, opts...)
	if err != nil {
		log.Error("failed to initialize backend gRPC auth client", slog.Any("error", err))
	}

	return &authHandler{
		log:        log,
		cfg:        cfg,
		gRPCClient: authv1.NewAuthClient(conn),
		conn:       conn,
	}
}

// Close gracefully closes the underlying gRPC network connection channel
func (h *authHandler) Close() error {
	if h.conn != nil {
		return h.conn.Close()
	}

	return nil
}

// handleGrpcError converts backend gRPC status codes into corresponding REST HTTP status codes
func (h *authHandler) handleGrpcError(c fiber.Ctx, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	// Log error response context cleanly
	h.log.Warn("gRPC processing execution rejected",
		slog.String("code", st.Code().String()),
		slog.String("message", st.Message()),
	)

	// Map according to standard RPC -> REST status rules
	switch st.Code() {
	case 3: // InvalidArgument
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": st.Message()})
	case 5: // NotFound
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": st.Message()})
	case 6: // AlreadyExists
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": st.Message()})
	case 7: // PermissionDenied
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": st.Message()})
	case 16: // Unauthenticated
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": st.Message()})
	case 8: // ResourceExhausted
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": st.Message()})
	default:
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Bad Gateway: communication error with auth backends"})
	}
}
