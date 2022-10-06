# Architecture Documentation

## Motivation

This documentation tries to make visible Weave Gitops Enterprise architecture in a simple way. 

You would benefit of this documentation is you are in any of the following journeys.

1. I want to understand how Weave Gitops Enterprise looks as a system in a simple way.
2. I want to understand the high level  building blocks or component within Weave Gitops Enterprise in a simple way. 
3. I want to understand the different business domains served by Weave Gitops Enterprise in a simple way.
4. For any of the domains, I want to go deeper in terms of behaviour, api or code in. 

## Assumptions and Limitations

Diagrams aim to be self-explanatory however:

1. They are based on [C4 Model](https://c4model.com/). If you have problems understanding them please take some time
   to get familiar via skimming [abstractions](https://c4model.com/#Abstractions) and [notation](https://c4model.com/#Notation)
   or  [watch this](https://www.youtube.com/watch?v=x2-rSnhpw0g).
2. They are using concepts from Domain Driven Design. If it gets difficult to read, please have a look to
   the following [article](https://medium.com/@ruxijitianu/summary-of-the-domain-driven-design-concepts-9dd1a6f90091).
3. We are using [mermaid](https://mermaid-js.github.io/mermaid/#/) for diagraming. It currently has an early support 
   for [C4](https://mermaid-js.github.io/mermaid/#/c4c) with known limitations like markdown rendering via github. 
   To overtake this limitation you will see both the image (in svg) and the source code as `mermaid-source` code-block.

## Glossary

TBA

## Weave Gitops Enterprise as System

It aims to represent Weave Gitops Enterprise as a whole system and how it fits into the world around it.

![Weave Gitops System Diagram](./imgs/system-context.svg)

```mermaid
C4Context
      title Weave Gitops Enterprise - System Context Diagram
      Person(platformOperator, "Platform Operator")
      Person(applicationDeveloper, "Application Developer")      
      System(weaveGitopsEnterprise, "Weave Gitops Enterprise")

      Rel(platformOperator, weaveGitopsEnterprise, "Manages Platform")
      Rel(applicationDeveloper, weaveGitopsEnterprise, "Delivers Application")
      Rel(weaveGitopsEnterprise, Github, "gitops flows")
      Rel(weaveGitopsEnterprise, KubernetesCluster, "read resources via api")
      Rel(weaveGitopsEnterprise, idp, "authenticate users via OIDC")

      Container_Boundary(runtime, "runtime") {
         System_Ext(KubernetesCluster, "Kubernetes Cluster", "run customer applications")

         System_Ext(Flagger, "Flagger", "controller for progressive delivery capabilities")
         Rel(Flagger, KubernetesCluster, "manage canary resources from")

         System_Ext(Flux, "Flux", "deploy customer applications")
         Rel(Flux, KubernetesCluster, "deploy apps to")
         Rel(Flux, Github, "read apps from")        
      }

      Container_Boundary(git, "git") {
          System_Ext(Github, "GitHub", "Source storage in Git")      
      }

      Container_Boundary(auth, "auth") {
        System_Ext(idp, "Identity Provider", "provides authn services for customer users")      
      }


      UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="3")    
```

## Weave Gitops Enterprise as Application Tiers

Weave Gitops Enterprise as tiered application that could be seen in the following diagram

![Container Diagram Capabilities](imgs/tiers.svg)

```mermaid-source
C4Container
  title Weave Gitops Enterprise
  Person(platformOperator, "Platform Operator")
  Person(developer, "Application Developer")      
  Rel(platformOperator, weaveGitopsEnterpriseUi, "Manages Platform")
  UpdateRelStyle(platformOperator, weaveGitopsEnterpriseUi, "", "", "-115", "-40")

  Rel(developer, weaveGitopsEnterpriseUi, "Delivers Application")
  UpdateRelStyle(developer, weaveGitopsEnterpriseUi, "", "", "50", "-40")

  Container_Boundary(weaveGitopsEnterprise, "Weave Gitops Enterprise") {
      Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
      Container(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend","golang","backend application with grpc api")
      Rel(weaveGitopsEnterpriseUi, weaveGitopsEnterpriseBackend, "consume api")
      UpdateRelStyle(weaveGitopsEnterpriseUi, weaveGitopsEnterpriseBackend, "", "", "-35", "-15")
      Rel(weaveGitopsEnterpriseBackend, KubernetesCluster, "get resources via api")
      UpdateRelStyle(weaveGitopsEnterpriseBackend, KubernetesCluster, "", "", "-55", "-15")
  }
  Rel(weaveGitopsEnterpriseBackend, idp, "authenticate users via oidc")
  UpdateRelStyle(weaveGitopsEnterpriseBackend, idp, "", "", "-105", "-15")
  Rel(weaveGitopsEnterpriseBackend, Github, "interacts for gitops flows")
  UpdateRelStyle(weaveGitopsEnterpriseBackend, Github, "", "", "5", "-15")

  Container_Boundary(runtime, "runtime") {
         System_Ext(KubernetesCluster, "Kubernetes Cluster", "run customer applications")
         Container(PolicyAgent, "Policy Agent", "", "components for enforcing policy resources")
         Rel(PolicyAgent, KubernetesCluster, "read policies from")
        UpdateRelStyle(PolicyAgent, KubernetesCluster, "", "", "-35", "-15")

         System_Ext(Flagger, "Flagger", "controller for progressive delivery capabilities")
         Rel(Flagger, KubernetesCluster, "manage canary resources from")
         UpdateRelStyle(Flagger, KubernetesCluster, "", "", "-35", "30")

         System_Ext(Flux, "Flux", "deploy customer applications")
         Rel(Flux, KubernetesCluster, "deploy apps to")
         UpdateRelStyle(Flux, KubernetesCluster, "", "", "75", "15")
         Rel(Flux, Github, "read apps from")        
        UpdateRelStyle(Flux, Github, "", "", "25", "15")
        
        System_Ext(CAPI, "CAPI", "manages kube clusters in the infra layer")
        Rel(CAPI, KubernetesCluster, "manage capi resources from")
        UpdateRelStyle(CAPI, KubernetesCluster, "", "", 10", "85")
        Rel(CAPI, aws, "manage clusters in cloud")
        UpdateRelStyle(CAPI, aws, "", "", "75", "35")
  }

  Container_Boundary(auth, "auth") {
    System_Ext(idp, "Identity Provider", "provides authn services for customer users")      
  }

  Container_Boundary(git, "git") {
    System_Ext(Github, "GitHub", "Source storage in Git")      
  }

 Container_Boundary(cloud, "cloud") {
          System_Ext(aws, "AWS", "aws cloud or any other public or private infra layer")
 }

UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="2") 
```

## Weave Gitops Enterprise as Business Domains

From the previous view, we could go a level deeper to understand the different 
business domains provided weave gitops enterprise.

![Container Diagram Capabilities](imgs/domains.svg)

```mermaid-source
C4Container
    title Weave Gitops Enterprise - Domains
    Container_Boundary(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend") {
          Container_Boundary(clusterManagement, "Cluster Management") {
            Component(ClusterDomain, "Cluster Domain",, "provides capabilities around managing kuberetens cluster via CAPI")            
          }
          Container_Boundary(applicationDelivery, "Application Delivery") {
            Component(ProgressiveDelivery, "Progressive Delivery Domain",, "canary and other progressive delivery capabilities using flagger")
            Component(Pipelines, "Pipelines Domain",, "continuous delivery pipeline capabilities based on flux")
          }
          Container_Boundary(trustedDelivery, "Trusted Delivery") {
            Component(Policy, "Policy Domain", "golang", "poilcy deifnition and enforcement capabilities based on magalix policy agent")
          }
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")               
```

Extend understanding on any of the domain by going into their page 

- [Cluster Management](cluster-management.md)
- [Application Delivery](application-delivery.md) 
- [Trusted Delivery](trusted-delivery.md)
