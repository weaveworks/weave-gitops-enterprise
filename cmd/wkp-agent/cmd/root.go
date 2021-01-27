package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	NatsURL        string
	KubeconfigFile string
	LogLevel       int
)

var rootCmd = &cobra.Command{
	Use:   "wkp-agent",
	Short: "Agent that connects to Weave WKP to send Kubernetes events",
}

func Execute() {
	setupLogger()
	autodetectKubeconfig()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&NatsURL, "nats-url", nats.DefaultURL, "NATS url")
	rootCmd.PersistentFlags().StringVar(&KubeconfigFile, "kubeconfig", "", "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().IntVar(&LogLevel, "log-level", 4, "logging level (0-6)")
	_ = rootCmd.ParseFlags(os.Args)
}

func autodetectKubeconfig() {
	defaultKubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// file is there, lets use it
	if _, err := os.Stat(defaultKubeconfigPath); err == nil {
		KubeconfigFile = defaultKubeconfigPath
	}
}

func setupLogger() {
	if LogLevel < 0 || LogLevel > 6 {
		fmt.Println("The log-level argument should have a value between 0 and 6.")
		os.Exit(1)
	} else {
		log.SetLevel(log.Level(LogLevel))
	}
}
