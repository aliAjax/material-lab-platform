package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"material-lab-platform/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	runner := worker.NewRunner(worker.NewMemoryQueue(), "material-lab-worker", 15*time.Second)
	runner.Register("notification", worker.NotificationHandler(logger))
	runner.Register("certificate_html", worker.CertificateHandler("./data/certificates", logger))
	slog.Info("worker started")
	if err := runner.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("worker stopped unexpectedly", "error", err)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}
