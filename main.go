package main

import (
	"esdi/cmd"
	"log/slog"
	"os"
)

func main() {
	output, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic("failed to open logging file: " + err.Error())
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	// Launches the cobra package stuff
	cmd.Execute(logger)
}
