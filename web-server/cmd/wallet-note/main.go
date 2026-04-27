package main

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ke1ta1to/wallet-note/internal/app"
	"github.com/ke1ta1to/wallet-note/internal/shared/config"
	"github.com/ke1ta1to/wallet-note/internal/shared/ddb"
)

func main() {
	setupLogger()
	cfg := config.Load()

	ctx := context.Background()
	db, err := ddb.NewClient(ctx)
	if err != nil {
		slog.Error("init dynamodb client", "err", err)
		os.Exit(1)
	}

	mux := app.NewMux(db, cfg.TableName)

	addr := ":" + cmp.Or(os.Getenv("PORT"), "8080")
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// JSON on Lambda for CloudWatch Logs Insights, text locally for humans.
func setupLogger() {
	var handler slog.Handler
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))
}
