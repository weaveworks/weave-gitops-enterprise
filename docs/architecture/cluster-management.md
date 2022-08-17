# Cluster Management 
This document outlines an architecture documentation for Weave Gitops Cluster Management domain.

## Motivation

TBA

## Audience
You would be interested in know about Cluster Management Domain if
1. You are working in a capability within the domain.
2. You are working in a capability in another domain that has a dependency with it.
3. You are not working in the context of the domain nor dependent, but want to understand a bit more
of the wider weave gitops architecture.

## Glossary

- Cluster Management

## Cluster Management Architecture

Diagrams are based on [C4 Model](https://c4model.com/). Note that there are some limitations with the visualization of 
diagrams due to c4models integration with mermaid and markdown.

### Weave Gitops Enterprise - Cluster Management Domain - Context Diagram

This section shows the context where personas could make use of application delivery capabilities within weave gitops.

![Context Diagram](./imgs/cluster-mangement-context.svg)

```mermaid
C4Context
  title Weave Gitops Enterprise

    Boundary(b0, "GitOps Operations") {
    Person(platformOperator, "Platform Operator", "Operates Kubernetes clusters for applications")
    Person(applicationDeveloper, "Application Developer", "Writes and operates Line-of-Business applications")

    System_Ext(github, "GitHub", "Source storage in Git")      
    System_Ext(aws, "Amazon WebServices", "AWS APIs for managing clusters") 

    Boundary(kubernetes, "Kubernetes Cluster") {
        Boundary(wg, "Weave Gitops Enterprise") {  
            Component(weaveGitopsEnterpriseUI, "Weave Gitops Enterprise UI")
            Component(clusterController, "Cluster Controller")
            Component(clusterBootstrapController, "Cluster Bootstrap Controller")
        }

        Boundary(fluxb, "Flux subsystem") {  
            Component(sourceController, "Source Controller")
        }

        Boundary(capib, "CAPI subsystem") {  
            Component(capiController, "Cluster API Controller")
            Component(capiAWSController, "Cluster API for AWS Controller")
        }
    }
  }

Rel(platformOperator, github, "GitHub Flow for changes")
Rel(applicationDeveloper, github, "GitHub Flow for changes")
Rel(platformOperator, weaveGitopsEnterpriseUI, "Cluster overview")
Rel(applicationDeveloper, weaveGitopsEnterpriseUI, "Application view")

Rel(sourceController, github, "Archive source")
Rel(capiAWSController, aws, "Create and update clusters")
Rel(clusterBootstrapController, clusterController, "Track cluster state")
Rel(capiAWSController, capiController, "Update cluster state")

%% UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")
```

Its api could be found [here](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/cmd/clusters-service/api/cluster_services.proto)

Next Steps:
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)











