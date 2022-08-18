# Trusted Delivery 
This document outlines an architecture documentation for Weave Gitops Trusted Delivery domain.

## Motivation


## Audience
You would be interested in know about Trusted Delivery Domain if
1. You are working in a capability within the domain.
2. You are working in a capability in another domain that has a dependency with it.
3. You are not working in the context of the domain nor dependent, but want to understand a bit more
of the wider weave gitops architecture.

## Glossary

- Trusted Delivery
- Policy 

## Trusted Architecture

Diagrams are based on [C4 Model](https://c4model.com/). Note that there are some limitations with the visualization of 
diagrams due to c4models integration with mermaid and markdown.

### Weave Gitops Enterprise - Trusted Delivery Domain - Context Diagram

This section shows the context where personas could make use of application delivery capabilities within weave gitops.

![Context Diagram](./imgs/trusted-delivery-context.svg)

```mermaid
C4Context

      title Trusted Delivery - System Context Diagram
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

### Weave Gitops Enterprise - Trusted Delivery Domain - Container Diagram

Weave Gitops Enterprise as tiered application that could be seen in the following diagram

![Container Diagram Capabilities](imgs/trusted-delivery-container-tiers.svg)

```mermaid
C4Container
  title "Trusted Delivery - Container Diagram - Tiers"
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

### Trusted Delivery - Policy Capabilities - Component Diagram

Looking into Trusted Delivery Domain capabilities we could see the following

![Component Diagram Capabilities](./imgs/trusted-delivery-container.svg)

```mermaid
C4Component
    title "Application Delivery - Container Diagram - Capabilities"
    Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
    Rel(weaveGitopsEnterpriseUi, Policy, "uses policy api")
    Container_Boundary(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend") {
      Container(Policy, "Policy", "golang", "provides policy capabilities")
      Rel(Policy, KubernetesApi, "read policy resourcess")
    }
    Container_Boundary(KubernetesCluster, "KubernetesCluster") {
      Component(KubernetesApi, "Kubernetes Api")
      Component(PolicyAgent, "Policy Agent")
      Rel(PolicyAgent, KubernetesApi, "enforce policies from")
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")          
```
Its api could be found [here](https://github.com/weaveworks/policy-agent/tree/dev/api)

Next Steps:
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [policy agent repo](https://github.com/weaveworks/policy-agent)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)











