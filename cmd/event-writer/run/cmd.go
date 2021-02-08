package run

import (
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/event-writer/liveness"
	"github.com/weaveworks/wks/cmd/event-writer/queue"
	"github.com/weaveworks/wks/cmd/event-writer/subscribe"
	"github.com/weaveworks/wks/common/database/utils"
)

var LogLevel int

// Cmd to start the event-writer process
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Start event-writer process",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runCommand(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	natsURL      string
	natsSubject  string
	dbURI        string
	timeInterval int
	batchSize    int
}

var globalParams paramSet

func init() {
	Cmd.Flags().StringVar(&globalParams.natsURL, "nats-url", os.Getenv("NATS_URL"), "URL of the NATS server to connect to")
	Cmd.Flags().StringVar(&globalParams.natsSubject, "nats-subject", os.Getenv("NATS_SUBJECT"), "NATS subject to subscribe to")
	Cmd.Flags().StringVar(&globalParams.dbURI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
	Cmd.Flags().IntVar(&globalParams.timeInterval, "time-interval", 3, "time interval in seconds for writing messages to the DB")
	Cmd.Flags().IntVar(&globalParams.batchSize, "batch-size", 100, "batch size of writes")
	Cmd.PersistentFlags().IntVar(&LogLevel, "log-level", 4, "logging level (0-6)")
}

func runCommand(globalParams paramSet) error {
	setupLogger()

	if globalParams.dbURI == "" {
		return errors.New("--db-uri not provided and $DB_URI not set")
	}
	if globalParams.natsSubject == "" {
		return errors.New("please specify the NATS subject the event-writer should subscribe to")
	}
	if globalParams.natsURL == "" {
		return errors.New("please specify the NATS server URL the event-writer should connect to")
	}

	initializeQueues(globalParams)

	// Connect to the DB
	_, err := utils.Open(globalParams.dbURI)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to connect to the database at %s", globalParams.dbURI))
	}

	go liveness.StartLivenessProcess()

	log.Info(fmt.Printf("subscribing to %s at NATS server %s\n", globalParams.natsSubject, globalParams.natsURL))
	err = subscribe.ToSubject(globalParams.natsURL, globalParams.natsSubject, subscribe.ReceiveEvent)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to subscribe to NATS server %s and subject %s", globalParams.natsURL, globalParams.natsSubject))
	}
	log.Info(fmt.Printf("unsubscribed from %s", globalParams.natsSubject))
	return nil
}

func setupLogger() {
	if LogLevel < 0 || LogLevel > 6 {
		fmt.Println("The log-level argument should have a value between 0 and 6.")
		os.Exit(1)
	} else {
		log.SetLevel(log.Level(LogLevel))
	}
}

func initializeQueues(p paramSet) {
	queue.BatchSize = p.batchSize
	queue.TimeInterval = time.Duration(p.timeInterval) * time.Second
	queue.LastWriteTimestamp = time.Now()
}
