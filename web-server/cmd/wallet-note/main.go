package main

import (
	"cmp"
	"log/slog"
	"net/http"
	"os"

	"github.com/ke1ta1to/wallet-note/internal/shared/router"
	"github.com/ke1ta1to/wallet-note/internal/user"
)

func main() {
	setupLogger()

	addr := ":" + cmp.Or(os.Getenv("PORT"), "8080")
	slog.Info("listening", "addr", addr)

	mux := router.New(user.New())
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// JSON output on Lambda for CloudWatch Logs Insights, text locally for humans.
func setupLogger() {
	var handler slog.Handler
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))
}
