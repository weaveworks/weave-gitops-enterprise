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
- Policy: ability to define policies to enforce at runtime for any workload running in kubernetes. 

**In Action**
- Available via weave gitops enterprise [policy experience](https://demo-01.wge.dev.weave.works/policies)

**Documentation and Next Steps**
- [API](https://github.com/weaveworks/policy-agent/tree/dev/api)
- [code](https://github.com/weaveworks/weave-gitops-enterprise)
- [policy agent repo](https://github.com/weaveworks/policy-agent)
- [user documentation](https://docs.gitops.weave.works/docs/enterprise/intro/index.html)











