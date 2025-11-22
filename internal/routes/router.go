package routes

import (
	"log/slog"
	"os"

	"agnos_demo/internal/handlers"
	"agnos_demo/internal/middleware"
	"agnos_demo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Config struct {
	Port int
}

func InitConfig() (*Config, error) {
	return &Config{
		Port: viper.GetInt("API.HTTPServerPort"),
	}, nil
}

func NewRouter(service *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SlogMiddleware())

	// Create logger for handlers
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	h := handlers.NewHandlers(service.DB, logger)

	// Public routes
	r.GET("/health", h.HealthCheck)
	r.POST("/staff/login", h.LoginStaff)

	// Protected routes
	staffProtectedRoute := r.Group("/staff")
	staffProtectedRoute.Use(middleware.AuthMiddleware())
	{
		staffProtectedRoute.POST("/create", h.CreateStaff)
	}
	patientProtectedRoute := r.Group("/patient")
	patientProtectedRoute.Use(middleware.AuthMiddleware())
	{
		patientProtectedRoute.GET("/search", h.SearchPatient)
		patientProtectedRoute.GET("/search/:id", h.GetPatientByID)
	}

	return r
}
