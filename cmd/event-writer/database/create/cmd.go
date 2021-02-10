package create

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/common/database/utils"
)

// Cmd to create the MCCP database
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create the MCCP database.",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runCommand(globalParams)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

type paramSet struct {
	dbURI string
}

var globalParams paramSet

func init() {
	Cmd.Flags().StringVar(&globalParams.dbURI, "db-uri", os.Getenv("DB_URI"), "URI of the database")
}

func runCommand(globalParams paramSet) error {
	if globalParams.dbURI == "" {
		return errors.New("--db-uri not provided and $DB_URI not set")
	}
	db, err := utils.Open(globalParams.dbURI)
	if err != nil {
		return err
	}
	// Set the Ref to the created DB
	err = utils.MigrateTables(db)
	if err != nil {
		return err
	}
	return nil
}
