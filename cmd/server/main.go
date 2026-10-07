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

	"github.com/Dragonqos/lockwood-test/internal/runtime"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	services, err := runtime.NewServices(runtime.Config{HTTPAddr: os.Getenv("HTTP_ADDR")})
	if err != nil {
		logger.Error("failed to initialize services", "error", err)
		os.Exit(1)
	}
	defer services.Close()

	server := &http.Server{
		Addr:              services.Config.HTTPAddr,
		Handler:           services.Handlers,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("chat server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	_ = services.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
