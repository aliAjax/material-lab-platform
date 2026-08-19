package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"material-lab-platform/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ctx, cancel := workerContext(context.Background())
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

func workerContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-parent.Done()
		time.Sleep(time.Second)
		cancel()
	}()
	return ctx, cancel
}
