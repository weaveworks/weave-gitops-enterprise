# Development Experience 
This document outlines an architecture documentation for Weave Gitops Development Experience domain

## Motivation

Development within gitops represents different challenges to traditional software development.
To reduce the adoption barriers is the problem to solve within this domain.

## Glossary

TBA 

## GitopsRun

The purpose of GitOps Run is to remove the complexity for developers so that Platform Engineers can create developer environments easily, 
and application developers can benefit from GitOps and focus on writing code.

For `gitops run` the main user flows are:
1. A developer develops apps using GitopsRun session against a local or remote cluster
2. A developer or platform engineer manages gitopsrun sessions via Weave GitOps Enterprise UI.  

```mermaid
C4Component
    title GitopsRun

    Boundary(gitopsOperation, "GitOps Operations") {
        Person(Developer, "Developer", "Manage business applications running in the platform")
        Person(PlatformEngineer, "Platform Engineer", "operates WGE platform for applications")

        Boundary(gitopsCli, "Gitops Cli") {
            Component(gitopsRunCommand, "GitopsRun Command")
        }
        Rel(Developer, gitopsRunCommand, "start gitops run session")
        UpdateRelStyle(Developer, gitopsRunCommand, $offsetX="-120", $offsetY="-70")

        Boundary(Kubernetes, "Kubernetes Cluster") {

            Boundary(gitopsRun, "GitopsRun subsystem") {
                Component(bucket, "Bucket","Source for gitopsRun resources")
            }
            Rel(gitopsRunCommand, bucket, "push development changes")
            UpdateRelStyle(gitopsRunCommand, bucket, $offsetX="-120", $offsetY="10")

            Boundary(wg, "Weave GitOps Enterprise") {
                Component(WeaveGitopsEnterpriseUI, "Weave GitOps Enterprise UI")
            }
            Rel(PlatformEngineer, WeaveGitopsEnterpriseUI, "View gitops run sessions")
            Rel(Developer, WeaveGitopsEnterpriseUI, "View gitops run sessions")

            Boundary(kubecp, "Kube Control Plane") {
                Component(KubernetesApi, "Kubernetes API")
            }

            Boundary(fluxb, "Flux") {
                Component(fluxControllers, "Flux Controllers")
                Component(sourceController, "Source Controller")
            }
            Rel(sourceController, bucket, "sync development changes")
            UpdateRelStyle(sourceController, bucket, $offsetX="-150", $offsetY="10")

            Rel(sourceController, KubernetesApi, "sync development changes")
            UpdateRelStyle(sourceController, KubernetesApi, $offsetX="10", $offsetY="-50")
            Rel(fluxControllers, KubernetesApi, "deploy development resources")
            UpdateRelStyle(fluxControllers, KubernetesApi, $offsetX="-120", $offsetY="-50")
        }
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="3")
```

**In Action**
- Available via Weave GitOps Enterprise [experience](https://demo-01.wge.dev.weave.works/gitopsrun)

**Documentation and Next Steps**
- [code](https://github.com/weaveworks/weave-gitops/tree/main/cmd/gitops/beta/run)
- [user documentation](https://docs.gitops.weave.works/docs/gitops-run/overview/)



