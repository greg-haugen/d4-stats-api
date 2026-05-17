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

	"github.com/ghaugen/d4-stats-api/internal/config"
	"github.com/ghaugen/d4-stats-api/internal/handler"
	"github.com/ghaugen/d4-stats-api/internal/middleware"
	"github.com/ghaugen/d4-stats-api/internal/platform"
	"github.com/ghaugen/d4-stats-api/internal/platform/battlenet"
	"github.com/ghaugen/d4-stats-api/internal/platform/psn"
	"github.com/ghaugen/d4-stats-api/internal/platform/steam"
	"github.com/ghaugen/d4-stats-api/internal/platform/xbox"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config error", "err", err)
		os.Exit(1)
	}

	clients := map[string]platform.Client{
		"battlenet": battlenet.NewClient(
			"https://oauth.battle.net",
			"https://us.api.blizzard.com",
			cfg.BnetClientID,
			cfg.BnetClientSecret,
		),
		"steam": steam.NewClient(
			"https://api.steampowered.com",
			cfg.SteamAPIKey,
		),
		"psn": psn.NewClient(
			"https://m.np.playstation.com",
			cfg.PSNAuthToken,
		),
		"xbox": xbox.NewClient(
			"https://login.microsoftonline.com",
			"https://user.auth.xboxlive.com",
			"https://xboxapi.example.com",
			cfg.XboxClientID,
			cfg.XboxClientSecret,
		),
	}

	router := handler.NewRouter(clients)

	chain := middleware.Logging(logger)(
		middleware.RateLimit(cfg.RateLimitRPS)(
			middleware.Auth(cfg.JWTSecret)(router),
		),
	)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      chain,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
}
