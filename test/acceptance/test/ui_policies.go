package acceptance

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func installPolicyAgent(clusterName string) {
	ginkgo.By(fmt.Sprintf("And install cert-manager to %s cluster", clusterName), func() {
		stdOut, _ := runCommandAndReturnStringOutput("helm search repo cert-manager")
		if !strings.Contains(stdOut, `cert-manager/cert-manager`) {
			err := runCommandPassThrough("helm", "repo", "add", "cert-manager", "https://charts.jetstack.io")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to add cert-manage repositoy")
		}

		err := runCommandPassThrough("helm", "upgrade", "--install", "cert-manager", "cert-manager/cert-manager", "--namespace", "cert-manager", "--create-namespace", "--version", "1.10.0", "--set", "installCRDs=true")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install cert-manager to leaf cluster '%s'", clusterName))
	})

	ginkgo.By(fmt.Sprintf("And install policy agent to %s cluster", clusterName), func() {
		stdOut, _ := runCommandAndReturnStringOutput("helm search repo policy-agent")
		if !strings.Contains(stdOut, `policy-agent/policy-agent`) {
			err := runCommandPassThrough("helm", "repo", "add", "policy-agent", "https://weaveworks.github.io/policy-agent")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to add policy-agent repositoy")
		}

		err := runCommandPassThrough("helm", "upgrade", "--install", "weave-policy-agent", "policy-agent/policy-agent", "--namespace", "policy-system", "--create-namespace", "--version", "2.2.x", "--set", "config.accountId=weaveworks", "--set", "config.clusterId="+clusterName)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install policy agent to leaf cluster '%s'", clusterName))
		_ = runCommandPassThrough("kubectl", "wait", "--for=condition=Ready", "--timeout=60s", "--namespace", "policy-system", "pod", "-l", "name=policy-agent")
	})
}
