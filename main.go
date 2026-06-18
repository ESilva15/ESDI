package main

import (
	"log/slog"
	"net/http"
	"os"

	"esdi/cmd"
	"esdi/telemetry"

	"github.com/arl/statsviz"
)

func main() {
	// NOTE: pass all these things to cobra
	output, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		panic("failed to open logging file: " + err.Error())
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	// Performance analysis
	mux := http.NewServeMux()
	statsviz.Register(mux)

	go func() {
		http.ListenAndServe("localhost:8001", mux)
	}()

	// Setting up some internal data structures
	telemetry.Init()

	// Launches the cobra package stuff
	cmd.Execute(logger)
}
