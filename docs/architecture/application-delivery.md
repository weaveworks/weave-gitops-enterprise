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

## Glossary

- Application Delivery Domain
- Pipelines
- Progressive Delivery

## Application Delivery Domains

Where in the context of application delivery we could find two domains

- Pipelines: enables a user to deliver application changes across different environment in an orchestrated manner. 
- Progressive Delivery: enables a user to deliver an application change into a given environment in a safe manner to optimise for application availability.

### Pipelines Domain

Pipelines enables a user to deliver application changes across different environment in an orchestrated manner.

It is composed by the following aggregates or capabilities

```mermaid-source
C4Component
      title Application Delivery - Pipelines Domain Component Diagram
      Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI")
      Rel(weaveGitopsEnterpriseUi, Pipeline, "read pipeline definitions")
      Container_Boundary(Pipelines, "Pipelines") {
        Component(Pipeline, "Pipeline", "golang","in development")
        Rel(Pipeline, KubernetesCluster, "reads pipeline resources")
      }
      Container_Boundary(external, "external") {
        System_Ext(KubernetesCluster, "Kubernetes Cluster")
      }
      UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")                           
```

- Pipeline: ability to define deployment pipelines for applications that you could follow across environments.  

**In Action**
- In development

**Documentation and Next Steps**

- [api](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/api/pipelines/pipelines.proto)
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)

### Progressive Delivery Domain

Progressive Delivery enables a user to deliver an application change into a given environment in a safe manner to optimise for application availability.

It is composed by the following aggregates or capabilities

```mermaid-source
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

- Canary: allows to read flagger [canaries](https://docs.flagger.app/usage/how-it-works#canary-resource).
- Metric Template: allow to read flagger [metric templates](https://docs.flagger.app/usage/metrics#custom-metrics).


**In Action**
- Available via weave gitops enterprise [delivery experience](https://demo-01.wge.dev.weave.works/applications/delivery)

**Documentation and Next Steps**
- [API](https://github.com/weaveworks/progressive-delivery/blob/main/api/prog/prog.proto)
- [progressive delivery repo](https://github.com/weaveworks/progressive-delivery)
- [weave gitops enterprise](https://github.com/weaveworks/weave-gitops-enterprise)
- [user documentation](https://docs.gitops.weave.works/docs/guides/delivery/0)














