# Application Delivery 
This document outlines an architecture documentation for Weave Gitops Application Delivery domain.

## Motivation
As software business, value is delivered to the users or customers by enabling them to do actions.
That enablement of the user is rarely a thing that happens statically but evolves over time.

In software, as the enablement of the user comes through software components, the evolution of that enablement
involves evolution of the underlying software components. Traditionally called in several ways, application, services ,etc ..

As [our mission](https://www.weave.works/company/) states
> to empower developers and DevOps teams to build better software faster.

We require to provide capabilities to enable evolution of the software. Application Delivery enables that part of our
business domain.


## Audience
You would be interested in know about application delivery domain if
1. You are working in a capability within the domain.
2. You are working in a capability in another domain that has a dependency with it.
3. You are not working in the context of the domain nor dependent, but want to understand a bit more
of the wider weave gitops architecture.

## Glossary

- Application Delivery Domain
- Pipelines
- Progressive Delivery

## Application Delivery Architecture

Diagrams are based on [C4 Model](https://c4model.com/). Note that there are some limitations with the visualization of 
diagrams due to c4models integration with mermaid and markdown.

### Weave Gitops Enterprise - Application Delivery Domain - Context Diagram

This section shows the context where personas could make use of application delivery capabilities within weave gitops.

![Context Diagram](./imgs/application-delivery-context.svg)

```mermaid
C4Context

      title Application Delivery - System Context Diagram
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

### Weave Gitops Enterprise - Application Delivery Domain - Container Diagram

Weave Gitops Enterprise as tiered application that could be seen in the following diagram

![Container Diagram Capabilities](imgs/application-delivery-container-tiers.svg)

```mermaid
C4Container
  title "Application Delivery - Container Diagram - Tiers"
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

![Container Diagram Capabilities](imgs/application-delivery-container.svg)

```mermaid
C4Container
    title "Application Delivery - Container Diagram - Capabilities"
    Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
    Rel(weaveGitopsEnterpriseUi, Pipelines, "uses pipelines api")
    Rel(weaveGitopsEnterpriseUi, ProgressiveDelivery, "uses progressive delivery api")
    Container_Boundary(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend") {
      Container(Pipelines, "Pipelines", "golang", "provides pipelines capabilities")
      Container(ProgressiveDelivery, "Progressive Delivery", "golang", "flagger service provides progressive delivery capabilities")
      Rel(Pipelines, KubernetesCluster, "read pipeline resourcess")
      Rel(ProgressiveDelivery, KubernetesCluster, "read progressive delivery resources")      
    }
    Container_Boundary(external, "external") {
      System_Ext(KubernetesCluster, "Kubernetes Cluster")
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")          
```

- Pipelines: enables a user to deliver application changes across different environment in an orchestrated manner. 
- Progressive Delivery: enables a user to deliver an application change into a given environment in a safe manner to optimise for application availability.


### Application Delivery - Pipelines Capability - Component Diagram

Pipelines enables a user to deliver application changes across different environment in an orchestrated manner.

It is composed by the following sub-capabilities

![Pipelines](imgs/application-delivery-pipelines.svg)

```mermaid

C4Component
      title Application Delivery - Pipelines Domain Component Diagram
      Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI")
      Rel(weaveGitopsEnterpriseUi, Pipeline, "read pipeline definitions")
      Rel(weaveGitopsEnterpriseUi, PipelineStatus, "read pipeline status")
      Container_Boundary(Pipelines, "Pipelines") {
        Component(Pipeline, "Pipeline", "golang","in development")
        Component(PipelineStatus, "PipelineStatus","golang", "in development")
        Rel(Pipeline, KubernetesCluster, "reads pipeline resources")
        Rel(PipelineStatus, KubernetesCluster, "reads pipeline status")      
      }
      Container_Boundary(external, "external") {
        System_Ext(KubernetesCluster, "Kubernetes Cluster")
      }
      UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")        
                    
```

- pipeline: ability to define pipelines, environments and associations with applications. 
- pipeline status: ability to follow an application change along the environments defined in a pipeline specification.

//TODO: move me to master
Its api could be found [here](https://github.com/weaveworks/weave-gitops-enterprise/blob/af0da2a895d205d837d1c7afaf29977225e01957/api/pipelines/pipelines.proto)

Next Steps:
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)

Capability could be seen in action via:
- In development

#### Progressive Delivery Capability

Progressive Delivery enables a user to deliver an application change into a given environment in a safe manner to optimise for application availability.

It is composed by the following sub-capabilities

![Progressive Delivery](imgs/application-delivery-progressive-delivery.svg)

```mermaid
C4Component
      title Application Delivery - Progressive Delivery Domain Component Diagram
      Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI")
      Rel(weaveGitopsEnterpriseUi, Canary, "read canaries via api")
      Rel(weaveGitopsEnterpriseUi, MetricTemplate, "read metric templates via api")
      Container_Boundary(ProgressiveDelivery, "ProgressiveDelivery") {
        Component(Canary, "Canary", "golang", "service layer to read flagger canary resources")
        Component(MetricTemplate, "MetricTemplate", "golang", "service layer to read flagger metric template resources")
        Rel(Canary, KubernetesApi, "reads canary resources via api")
        Rel(MetricTemplate, KubernetesApi, "reads metric temaplate resources via api")      
      }
      Container_Boundary(Kubernetes, "Kubernetes Cluster") {
        System_Ext(Flagger, "Flagger","controller that provides the runtime for progressive delivery")
        System_Ext(KubernetesApi, "Kubernetes API")
      }
      UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")       
```

- canaries: allows to interact with flagger [canaries](https://docs.flagger.app/usage/how-it-works#canary-resource)
- metrics templates: allow to interact with flagger [metric templates](https://docs.flagger.app/usage/metrics#custom-metrics)

Its api could be seen [here](https://github.com/weaveworks/progressive-delivery/blob/main/api/prog/prog.proto)

Next Steps:
- [progressive delivery repo](https://github.com/weaveworks/progressive-delivery)
- [weave gitops enterprise](https://github.com/weaveworks/weave-gitops-enterprise)
- [user documentation](https://docs.gitops.weave.works/docs/guides/delivery/0)

This capability is available in weave gitops enterprise and could be seen in 
action in our [demo environments](https://demo-01.wge.dev.weave.works/applications/delivery)













