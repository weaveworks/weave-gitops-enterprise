package main

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/root"
)

func main() {
	cobra.CheckErr(root.NewRootCmd().Execute())
}
