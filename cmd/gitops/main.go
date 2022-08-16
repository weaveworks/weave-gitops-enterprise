package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
)

func main() {
	client := adapters.NewHTTPClient().EnableCLIAuth()

	if err := root.Command(client).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
