package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"esdi/cmd"
	"esdi/config"
	"esdi/telemetry"

	"github.com/arl/statsviz"
)

func initApplication() {
	// 1. This is the first initialization setup we do so we can log
	err := setupLogger()
	if err != nil {
		panic("failed to open logging file: " + err.Error())
	}

	// 2. This is the second thing we setup because we need some stuff to run
	// other things
	err = config.Setup("./config/config.yaml")
	if err != nil {
		panic("failed to setup config: " + err.Error())
	}

	// Performance analysis
	if config.GetCfg().MetricsServer {
		slog.Info("Setting up metrics server...")
		err = runStatviz()
		if err != nil {
			slog.Error(fmt.Sprintf("failed to setup metrics server: %+v", err))
		}
	}

	// Setting up some internal data structures
	telemetry.Init()
}

func setupLogger() error {
	output, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	slog.SetDefault(logger)

	return nil
}

func runStatviz() error {
	mux := http.NewServeMux()
	err := statsviz.Register(mux)
	if err != nil {
		return err
	}

	go func() {
		err := http.ListenAndServe("localhost:8001", mux)
		if err != nil {
			slog.Error("Failed to set up server for statsviz", "err", err)
		}
	}()

	return nil
}

func main() {
	initApplication()

	// Launches the cobra package stuff
	cmd.Execute()
}
