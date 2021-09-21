package main

import (
	stdlog "log"
	"math/rand"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/app"
	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UnixNano())

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

	command := app.NewAPIServerCommand(log)

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
