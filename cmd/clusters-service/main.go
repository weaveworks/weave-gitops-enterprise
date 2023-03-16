package main

import (
	stdlog "log"
	"math/rand"
	"os"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		stdlog.Fatalf("Failed to create a temp directory for Helm: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	command := app.NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
