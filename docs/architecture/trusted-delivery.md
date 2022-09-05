# Trusted Delivery 
This document outlines an architecture documentation for Weave Gitops Trusted Delivery domain.

## Motivation

TBA

## Glossary

- Trusted Delivery
- Policy 

## Trusted Delivery Domains

### Policy Domain

It is composed by the following aggregates or capabilities

![Component Diagram Capabilities](./imgs/trusted-delivery-container.svg)

```mermaid-source
C4Component
    title "Application Delivery - Container Diagram - Capabilities"
    Container(weaveGitopsEnterpriseUi, "Weave Gitops Enterprise UI","javascript and reactJs","weave gitops experience via web browser")
    Rel(weaveGitopsEnterpriseUi, Policy, "uses policy api")
    Container_Boundary(weaveGitopsEnterpriseBackend, "Weave Gitops Enterprise Backend") {
      Container(Policy, "Policy", "golang", "policy entity business logic provided via api")
      Rel(Policy, KubernetesApi, "read policy resourcess")
      Container(Violation, "Violation", "golang", "violation entity business logic provided via api")
      Rel(Violation, KubernetesApi, "read kube events")
    }
    Container_Boundary(KubernetesCluster, "KubernetesCluster") {
      Component(KubernetesApi, "Kubernetes Api")
      Component(PolicyAgent, "Policy Agent")
      Rel(PolicyAgent, KubernetesApi, "enforce policies from")
      Rel(PolicyAgent,ElasticSearch, "send validations")
    }
    Container_Boundary(PolicyProfile, "Policy Profile Helm Chart") {
        Component(PolicySet, "PolicySet","", "set of polices to provide customer baseline")
        Component(PolicyCRD, "Policy CRD","", "policy crd api")
        Component(PolicyAgentManifest, "Policy Agent Manifeset","", "installs policy agent")
    }
    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1") 
   Boundary(elk, "ELK") {
      System_Ext(Kibana, "Kibana", "visualices validations events") 
      System_Ext(ElasticSearch, "ElasticSearch", "stores for validation events") 
      Rel(Kibana,ElasticSearch, "read validation events")
   }         
```
- Policy: ability to define policies to enforce at runtime for any workload running in kubernetes. 

**In Action**
- Available via weave gitops enterprise [policy experience](https://demo-01.wge.dev.weave.works/policies)

**Documentation and Next Steps**
- [API](https://github.com/weaveworks/policy-agent/tree/dev/api)
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [policy agent repo](https://github.com/weaveworks/policy-agent)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)











