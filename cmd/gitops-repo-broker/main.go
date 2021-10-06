package main

import (
	stdlog "log"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/server"
	"go.uber.org/zap"
)

func main() {
	var log logr.Logger
	if os.Getenv("HUMAN_LOGS") != "" {
		if zl, err := zap.NewDevelopment(zap.AddCaller()); err != nil {
			stdlog.Fatalf("Failed to create a zap development logger: %v", err)
		} else {
			log = zapr.NewLogger(zl)
		}
	} else {
		if zl, err := zap.NewProduction(zap.AddCaller()); err != nil {
			stdlog.Fatalf("Failed to create a zap production logger: %v", err)
		} else {
			log = zapr.NewLogger(zl)
		}
	}

	cmd := server.NewAPIServerCommand(log)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
