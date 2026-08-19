package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopspring/decimal"
	httpadapter "material-lab-platform/internal/adapters/http"
	"material-lab-platform/internal/adapters/memory"
	"material-lab-platform/internal/adapters/postgres"
	"material-lab-platform/internal/application"
	"material-lab-platform/internal/auth"
	"material-lab-platform/internal/config"
	"material-lab-platform/internal/domain"
	"material-lab-platform/internal/health"
	"material-lab-platform/internal/ports"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration invalid", "error", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	var store ports.Store = memory.New()
	var database *postgres.Pool
	if cfg.DatabaseURL != "" {
		database, err = postgres.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			slog.Error("database unavailable", "error", err)
			os.Exit(1)
		}
		store, err = postgres.NewRepository(context.Background(), database)
		if err != nil {
			database.Close()
			slog.Error("database repository unavailable", "error", err)
			os.Exit(1)
		}
		defer database.Close()
		slog.Info("postgres repository enabled")
	}
	if os.Getenv("RUN_STARTUP_CHECKS") == "1" {
		startupChecks := health.Run(context.Background(), buildHealthChecks(database))
		for _, result := range startupChecks {
			if !result.Healthy {
				slog.Error("startup health check failed", "check", result.Name, "error", result.Error)
				os.Exit(1)
			}
		}
	}
	authService := auth.New(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	seedUsers(authService)
	app := application.New(store)
	seedDomain(app, authService)
	handler := httpadapter.New(app, authService, cfg)
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       25 * time.Second,
		WriteTimeout:      25 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server started", "address", cfg.HTTPAddress, "lab", cfg.LabName)
		errCh <- server.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	select {
	case signal := <-signals:
		slog.Info("shutdown requested", "signal", signal.String())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func buildHealthChecks(database *postgres.Pool) map[string]health.Check {
	checks := map[string]health.Check{
		"process": func(ctx context.Context) error { return ctx.Err() },
	}
	if database != nil {
		checks["database"] = database.Ready
	}
	return checks
}

func seedUsers(service *auth.Service) {
	users := []struct {
		username string
		name     string
		password string
		role     domain.Role
	}{
		{"manager", "林实验室负责人", "Manager123!", domain.RoleManager},
		{"registrar", "陈登记员", "Registrar123!", domain.RoleRegistrar},
		{"tester", "周试验员", "Tester123!", domain.RoleTester},
		{"reviewer", "吴复核员", "Reviewer123!", domain.RoleReviewer},
	}
	for _, seed := range users {
		if _, err := service.CreateUser(seed.username, seed.name, seed.password, seed.role); err != nil {
			slog.Error("seed user failed", "username", seed.username, "error", err)
			os.Exit(1)
		}
	}
}

func seedDomain(app *application.Service, authService *auth.Service) {
	users := authService.Users()
	var manager string
	for _, user := range users {
		if user.Role == domain.RoleManager {
			manager = user.ID
		}
	}
	minimum := decimalPtr("0")
	maximum := decimalPtr("1000")
	method := domain.Method{
		Code:          "GB-T228-TENSILE",
		Name:          "金属材料室温拉伸试验",
		MaterialScope: "金属材料",
		Version:       1,
		Fields: []domain.MethodField{
			{Name: "width", Label: "试样宽度", Type: domain.FieldNumber, Unit: "mm", Min: minimum, Max: maximum, Required: true},
			{Name: "thickness", Label: "试样厚度", Type: domain.FieldNumber, Unit: "mm", Min: minimum, Max: maximum, Required: true},
			{Name: "force", Label: "最大力", Type: domain.FieldNumber, Unit: "N", Min: minimum, Max: maximum, Required: true},
		},
		Formula:   "force / (width * thickness)",
		Precision: 2,
		Rule:      "result >= 400",
	}
	created, err := app.CreateMethod(context.Background(), manager, "startup-seed", method)
	if err != nil {
		slog.Error("seed method failed", "error", err)
		return
	}
	if _, err := app.PublishMethod(context.Background(), manager, created.ID, "startup-seed"); err != nil {
		slog.Error("publish seed method failed", "error", err)
	}
}

func decimalPtr(raw string) *decimal.Decimal {
	value := decimal.RequireFromString(raw)
	return &value
}
