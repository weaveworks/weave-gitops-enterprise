# Infrastructure 
This document outlines the architecture documentation for Infrastructure domain.

## Cluster Management

Bringing up a new Kubernetes Cluster is fairly easy, the [IaaS providers](https://azure.microsoft.com/en-gb/resources/cloud-computing-dictionary/what-is-iaas/)
provide APIs so that users can easily bring up clusters even without having to understand tools like
[`kubeadm`](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/).

Preparing that cluster for workloads can require a bit more work, the Cluster Management functionality provides
mechanisms for creating new [CAPI](https://cluster-api.sigs.k8s.io/) clusters from templates,
bootstrapping Flux into the clusters to start loading workloads from a git repository,
and installing packages of components (which we call [Profiles](https://docs.gitops.weave.works/docs/cluster-management/profiles/))
into newly bootstrapped clusters.

Our cluster-management functionality sets up a collaboration between CAPI, Flux and Helm (Profiles) for customer clusters,
and provides a single-pane-of-glass view of the workloads on these clusters.

```mermaid
C4Context
    title Weave GitOps Enterprise

    Boundary(gitopsOperation, "GitOps Operations") {
        Person(PlatformEngineer, "Platform Engineer", "operates WGE platform for applications")
        Person(ApplicationDeveloper, "Application Developer", "writes and operates Line-of-Business applications")

        Boundary(git, "Git") {
            System_Ext(gitProvider, "GitProvider", "Source storage in Git. Ex. GitHub")
        }

        Boundary(Kubernetes, "Kubernetes Cluster") {
            Boundary(fluxb, "Flux") {
                Component(sourceController, "Source Controller")
            }

            Boundary(wg, "Weave GitOps Enterprise") {
                Component(WeaveGitopsEnterpriseUI, "Weave GitOps Enterprise UI")
                Component(clusterController, "Cluster Controller")
                Component(clusterBootstrapController, "Cluster Bootstrap Controller")
            }
            Rel(PlatformEngineer, gitProvider, "Gitops flow for changes")
            UpdateRelStyle(PlatformEngineer, gitProvider, $offsetX="-140", $offsetY="-40")
            Rel(ApplicationDeveloper, gitProvider, "Gitops flow for changes")
            UpdateRelStyle(ApplicationDeveloper, gitProvider, $offsetX="80", $offsetY="-40")
            Rel(PlatformEngineer, WeaveGitopsEnterpriseUI, "Cluster overview")
            Rel(ApplicationDeveloper, WeaveGitopsEnterpriseUI, "Application view")


            Boundary(capib, "CAPI subsystem") {
                Component(capiController, "Cluster API Controller")
                Component(capiAWSController, "Cluster API for AWS Controller")
            }
            Rel(sourceController, gitProvider, "Archive source")
            UpdateRelStyle(sourceController, gitProvider, $offsetX="0", $offsetY="20")

            Rel(capiAWSController, cloudProvider, "Create and update clusters")
            Rel(clusterBootstrapController, clusterController, "Track cluster state")
            Rel(capiAWSController, capiController, "Update cluster state")
        }
        Boundary(cloud, "Cloud") {
            System_Ext(cloudProvider, "Cloud Provider", "Provide IaaS, ex. AWS")
        }

    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="3")
```

**In Action**
- Available via Weave GitOps Enterprise [clusters experience](https://demo-01.wge.dev.weave.works/clusters)

**Documentation and Next Steps**
- [API](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/cmd/clusters-service/api/cluster_services.proto)
- [Code](https://github.com/weaveworks/weave-gitops-enterprise)
- [User documentation](https://docs.gitops.weave.works/docs/cluster-management/intro/)

## Terraform

Addresses the problem of provisioning infrastructure beyond Kubernetes clusters for both platform  
and application developers. It uses Terraform as most prominent IaC solution nowadays. Leverages Weaveworks [TF-controller](https://github.com/weaveworks/tf-controller) 
to manage Terraform under Gitops principles and integrates with Weave GitOps.

Given a platform engineer or developer that wants to provision 
[Terraform infrastruture via TF-contorller](https://docs.gitops.weave.works/docs/terraform/using-terraform-cr/provision-resources-and-auto-approve/)

The common gitops flow applies:
- A PR is created to GitProvider (or other git provider) with the change.
- PR is reviewed and merged.
- Flux source controllers syncs it.
Then terraform flow kicks in:
- Terraform Controller reconciles Terraform Crs.
- Terraform Runners executes terraform jobs.
- The infrastructure is provisioned. 

```mermaid
C4Component
    title Weave GitOps Enterprise
    Boundary(gitopsOperation, "GitOps Operations") {
        Person(PlatformEngineer, "Platform Engineer", "operates WGE platform for applications")
        Person(ApplicationDeveloper, "Application Developer", "writes and operates Line-of-Business applications")
        Rel(PlatformEngineer, gitProvider, "Gitops flow for changes")
        UpdateRelStyle(PlatformEngineer, gitProvider, $offsetX="-140", $offsetY="-40")
        Rel(ApplicationDeveloper, gitProvider, "Gitops flow for changes")
        UpdateRelStyle(ApplicationDeveloper, gitProvider, $offsetX="80", $offsetY="-40")

        Boundary(git, "Git") {
            System_Ext(gitProvider, "GitProvider", "Source storage in Git. Ex. GitHub")
        }

        Boundary(Kubernetes, "Kubernetes Cluster") {
            Boundary(wg, "Weave GitOps Enterprise") {
                Component(WeaveGitopsEnterpriseUI, "Weave GitOps Enterprise UI")
            }
            Rel(ApplicationDeveloper, WeaveGitopsEnterpriseUI, "View Terraform")
            UpdateRelStyle(ApplicationDeveloper, WeaveGitopsEnterpriseUI, $offsetX="0", $offsetY="-40")
            Rel(PlatformEngineer, WeaveGitopsEnterpriseUI, "View Terraform")
            UpdateRelStyle(PlatformEngineer, WeaveGitopsEnterpriseUI, $offsetX="-110", $offsetY="-70")

            Boundary(kubecp, "Kube Control Plane") {
                Component(KubernetesApi, "Kubernetes API")
            }

            Boundary(fluxb, "Flux") {
                Component(sourceController, "Source Controller")
            }

            Rel(sourceController, gitProvider, "pull terraform source")

            Boundary(terraform, "Terraform") {
                Component(terraformController, "Terraform Controller", "manages terraform resources")
                Component(terraformRunner, "Terraform Runners", "terraform execution component")
            }

            Rel(terraformController, KubernetesApi, "read terraform manifests")
            Rel(terraformController, terraformRunner, "manage terraform executions")
            Rel(terraformRunner, infraProvider, "provisions infrastructure")
        }

        Boundary(cloud, "Cloud") {
            System_Ext(infraProvider, "Infrastructure Provider", "Any infrastructure that terraform supports")
        }

    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="2")
```

**In Action**
- Available via Weave GitOps Enterprise [terraform experience](https://demo-01.wge.dev.weave.works/terraform)

**Documentation and Next Steps**
- [API](https://github.com/weaveworks/weave-gitops-enterprise/tree/main/api/terraform)
- [WGE Terraform Code](https://github.com/weaveworks/weave-gitops-enterprise/tree/main/pkg/terraform)
- [Terraform Controller Code](https://github.com/weaveworks/tf-controller)
- [User documentation](https://docs.gitops.weave.works/docs/terraform/overview/)