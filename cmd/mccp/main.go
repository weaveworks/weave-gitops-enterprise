package main

import (
	"github.com/weaveworks/wks/cmd/mccp/cmd"
	"github.com/weaveworks/wks/pkg/cmdutil"
)

func main() {
	if err := cmd.Execute(); err != nil {
		cmdutil.ErrorExit("Error", err)
	}
}
