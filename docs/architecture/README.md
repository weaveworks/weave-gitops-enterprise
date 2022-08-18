# Architecture Documentation

## Motivation and Audience
To make visible Weave Gitops Enterprise architecture. 

You would be interested on it 

2. You are working in a capability within the domain.
3. You are working in a capability in another domain that has a dependency with it.
4. You are not working in the context of the domain nor dependent, but want to understand a bit more
   of the wider weave gitops architecture.

## Glossary

TBA

## Weave Gitops Enterprise

### Assumptions

Diagrams aim to be self-explanatory but 

1. They are based on [C4 Model](https://c4model.com/). If you have problems understanding them please take some time 
to get familiar via skimming [abstractions](https://c4model.com/#Abstractions) and [notation](https://c4model.com/#Notation) 
or  [watch this](https://www.youtube.com/watch?v=x2-rSnhpw0g).
2. They are using concepts from Domain Driven Design. If it gets difficult to read, please have a look to 
the following [article](https://medium.com/@ruxijitianu/summary-of-the-domain-driven-design-concepts-9dd1a6f90091). 

### System Diagram 

![Weave Gitops System Diagram](./imgs/system-context.svg)

```mermaid
C4Context

      title Weave Gitops Enterprise - Context Diagram
      Person(platformOperator, "Platform Operator")
      Person(developer, "Application Developer")      
      System(weaveGitopsEnterprise, "Weave Gitops Enterprise")
      System(ignore, "ignore")

      Rel(platformOperator, weaveGitopsEnterprise, "Manages Platform")
      Rel(developer, weaveGitopsEnterprise, "Delivers Application")
      Rel(weaveGitopsEnterprise, Git, "sync resources from")
      Rel(weaveGitopsEnterprise, KubernetesCluster, "read resources via api")

      System_Ext(KubernetesCluster, "Kubernetes Cluster")
      System_Ext(Git, "Git") 
      UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")

```

### Tiers

Weave Gitops Enterprise as tiered application that could be seen in the following diagram

![Container Diagram Capabilities](imgs/tiers.svg)

```mermaid
C4Container
  title Weave Gitops Enterprise - Tiers
  Person(platformOperator, "Platform Operator")
  Person(developer, "Application Developer")      
  Container_Boundary(weaveGitopsEnterprise, "Weave Gitops Enterprise") {
      Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
      Container(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend","golang","monlith backend application with grpc api")
      Rel(weaveGitopsEnterpriseUi, weaveGitopsEnterpriseBackend, "consumes via grpc")
      Rel(weaveGitopsEnterpriseBackend, KubernetesCluster, "consumes delivery resources via kubernetes api")
  }
  Rel(platformOperator, weaveGitopsEnterpriseUi, "Manages Platform")
  Rel(developer, weaveGitopsEnterpriseUi, "Delivers Application")
  Container_Boundary(external, "external") {
    System_Ext(KubernetesCluster, "Kubernetes Cluster")
    System_Ext(Git, "Git")     
  }
  Rel(weaveGitopsEnterpriseBackend, Git, "sync resources from")
  Rel(weaveGitopsEnterpriseBackend, KubernetesCluster, "consumes resources from")  

  UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")              
```
Looking into application delivery domain capabilities we could see the following

### Business Domains 

Looking into application delivery domain capabilities we could see the following

![Container Diagram Capabilities](imgs/domains.svg)

```mermaid
C4Container
    title Weave Gitops Enterprise - Domains
    Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
    Container_Boundary(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend") {
      Container(ClusterManagement, "Cluster Management Domain",, "provides capabilities around managing kuberetens cluster via CAPI")
      Container(ApplicationDelivery, "Application Delivery Domain",, "provides capabilities around delivery application into clusters and release safety")
      Container(TrustedDelivery, "Trusted Delivery Domain", "golang", "provides capabilities around policy for workloads")
    }
    Container_Boundary(external, "external") {
      System_Ext(KubernetesCluster, "Kubernetes Cluster")
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")          
```

Business Domains 

- [Cluster Management](cluster-management.md)
- [Application Delivery](application-delivery.md) 
- [Trusted Delivery](trusted-delivery.md)
