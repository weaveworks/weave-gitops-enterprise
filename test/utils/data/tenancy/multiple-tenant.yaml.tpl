tenants:
  - name: test-team
    namespaces:
      - test-kustomization
      - test-system
    allowedRepositories:
      - kind: GitRepository
        url: https://github.com/stefanprodan/podinfo
      - kind: HelmRepository
        url: https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages
      - kind: GitRepository
        url: {{ .MainRepoURL }}
    allowedClusters:
      - kubeConfig: wge-leaf-tenant-kind-kubeconfig
      - kubeConfig: workspaces-leaf-cluster-test-kubeconfig
    teamRBAC:
      groupNames:
        - wge-test-org:ci-test # CI bot user group (GitHub)
        - wge-test-org # CI bot user group (Gitlab on-prem)
        - weaveworks:pesto
        - weaveworks:QA
        - wge-test
        - {{ .Org }}
      rules:
        - apiGroups: [""]
          resources: ["secrets", "pods", "services"]
          verbs: ["get", "list"]
        - apiGroups: ["apps"]
          resources: ["deployments", "replicasets"]
          verbs: ["get", "list"]
        - apiGroups: ["kustomize.toolkit.fluxcd.io"]
          resources: ["kustomizations"]
          verbs: ["get", "list", "patch"]
        - apiGroups: ["helm.toolkit.fluxcd.io"]
          resources: ["helmreleases"]
          verbs: ["get", "list", "patch"]
        - apiGroups: ["source.toolkit.fluxcd.io"]
          resources:
            [
              "buckets",
              "helmcharts",
              "gitrepositories",
              "helmrepositories",
              "ocirepositories",
            ]
          verbs: ["get", "list", "patch"]
        - apiGroups: [""]
          resources: ["events"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["autoscaling"]
          resources: ["horizontalpodautoscalers"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["infra.contrib.fluxcd.io"]
          resources: ["terraforms"]
          verbs: ["get", "watch", "list", "patch"]
        - apiGroups: [""]
          resources: ["configmaps"]
          verbs: ["get", "list", "watch"]
        - apiGroups: ["gitops.weave.works"]
          resources: ["gitopsclusters"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["cluster.x-k8s.io"]
          resources: ["clusters"]
          verbs: ["get", "list", "watch"]
        - apiGroups: ["capi.weave.works"]
          resources: ["capitemplates"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["templates.weave.works"]
          resources: ["gitopstemplates"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["pac.weave.works"]
          resources: ["policies"]
          verbs: ["get", "list"]

  - name: dev-team
    namespaces:
      - dev-system
    allowedRepositories:
      - kind: GitRepository
        url: https://github.com/example-org/example-app
      - kind: HelmRepository
        url: https://raw.githubusercontent.com/weaveworks/example-catalogue/gh-pages
      - kind: GitRepository
        url: {{ .MainRepoURL }}
    allowedClusters:
      - kubeConfig: fake-kubeconfig
    teamRBAC:
      groupNames:
        - wge-test-org:ci-test # CI bot user group (GitHub)
        - wge-test-org # CI bot user group (Gitlab on-prem)
        - weaveworks:pesto
        - weaveworks:QA
        - wge-test
        - {{ .Org }}
      rules:
        - apiGroups: [""]
          resources: ["secrets", "pods", "services"]
          verbs: ["get", "list"]
        - apiGroups: ["apps"]
          resources: ["deployments", "replicasets"]
          verbs: ["get", "list"]
        - apiGroups: ["kustomize.toolkit.fluxcd.io"]
          resources: ["kustomizations"]
          verbs: ["get", "list", "patch"]
        - apiGroups: ["helm.toolkit.fluxcd.io"]
          resources: ["helmreleases"]
          verbs: ["get", "list", "patch"]
        - apiGroups: ["source.toolkit.fluxcd.io"]
          resources:
            [
              "buckets",
              "helmcharts",
              "gitrepositories",
              "helmrepositories",
              "ocirepositories",
            ]
          verbs: ["get", "list", "patch"]
        - apiGroups: [""]
          resources: ["events"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["autoscaling"]
          resources: ["horizontalpodautoscalers"]
          verbs: ["get", "watch", "list"]
        - apiGroups: ["infra.contrib.fluxcd.io"]
          resources: ["terraforms"]
          verbs: ["get", "watch", "list", "patch"]
