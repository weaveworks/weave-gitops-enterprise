package main

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/cmd"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cmdutil"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	if err := cmd.Execute(); err != nil {
		cmdutil.ErrorExit("Error", err)
	}
}
