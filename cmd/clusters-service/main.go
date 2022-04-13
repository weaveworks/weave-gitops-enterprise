package main

import (
	stdlog "log"
	"math/rand"
	"os"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
	"github.com/weaveworks/weave-gitops/core/logger"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		stdlog.Fatalf("Failed to create a temp directory for Helm: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	log, err := logger.New(logger.DefaultLogLevel, os.Getenv("HUMAN_LOGS") != "")
	if err != nil {
		stdlog.Fatalf("Couldn't set up logger: %v", err)
	}

	command := app.NewAPIServerCommand(log, tempDir)

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
