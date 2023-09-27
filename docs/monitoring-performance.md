# Monitoring Performance for Weave Gitops Enterprise

This document tries to provide an overview of performance monitoring for Weave Gitops Enterprise from two angles:
- how it looks in general and 
- how it could be used for troubleshooting performance issues.

## Monitoring Overview

### Metrics

Performance monitoring for Weave Gitops Enterprise happens mostly driven by metrics. Both the management console and controllers
are instrumented to generate Prometheus metrics. In addition, given that our applications are deployed to Kubernetes, 
Kubernetes metrics for workloads are also used. In summary, we have the main three monitoring layers:

 **Go runtime metrics via [prometheus client_golang](https://github.com/prometheus/client_golang/blob/1bae6c1e6314f6a20be183a7277059630780232a/prometheus/collectors/go_collector_latest.go)**

![overview-golang-runtime.png](monitoring%2Fimgs%2Foverview-golang-runtime.png)

 **API and Control Plane metrics via [go-http-metrics](https://github.com/slok/go-http-metrics) and [kubebuilder metrics](https://book.kubebuilder.io/reference/metrics-reference)**

![overview-wge.png](monitoring%2Fimgs%2Foverview-wge.png)

 **Component server metrics for example [Explorer](https://github.com/weaveworks/weave-gitops-enterprise/blob/b643619464104e59a17e77a697cd7c290f96889a/pkg/query/collector/metrics/recorder.go)**

![explorer emtrics](monitoring%2Fimgs%2Fexplorer-query-metrics-87ba3ddbfb12169b31b27e4f9ea8c722.png)

 - Kubernetes Workload metrics
![overview-kubernetes.png](monitoring%2Fimgs%2Foverview-kubernetes.png)
The monitoring stack is deployed as [Flux Kustomization](https://github.com/weaveworks/weave-gitops-quickstart/tree/add-monitoring) that includes:
- Prometheus 
- Grafana
- Kubernetes Dashboards
- Flux Dashboards 
- Weave Gitops Grafana dashboards

This is included in:
- Dev environment (via Tilt) so it could be used during development for understanding feature performance.
- [Staging cluster](https://github.com/weaveworks/weave-gitops-clusters/tree/main/k8s/clusters/internal-dev-gke/monitoring) so it could be used to long-live monitoring a feature or the app. 

### Profiling 

Apart from metrics, Weave Gitops Enterprise leverages golang profiling capabilities [pprof](https://pkg.go.dev/runtime/pprof) 
for complementing the understanding provided via metrics. For an example on using metrics and profiling for troubleshooting 
memory leaks, see [troubleshooting performance](#troubleshooting-performance-issues).

Any environment could by profiled by enabling the configuration [`WEAVE_GITOPS_ENABLE_PROFILING`](https://github.com/weaveworks/weave-gitops-enterprise/blob/b643619464104e59a17e77a697cd7c290f96889a/cmd/clusters-service/app/server.go#L843)
that exposes an [http endpoint for pprof](https://pkg.go.dev/net/http/pprof) for any of the available profiles. 

Then it could be used remote or locally used via pprof tool. An example a memory heap dump visualised via pprof `go tool pprof -http=:8082 heap` could be:

![pprof-web-ui.png](monitoring%2Fimgs%2Fpprof-web-ui.png)

Profiling is enabled by default in [dev via Tilt](../tools/dev-values.yaml) 

## Troubleshooting Performance Issues

As developer, we build up features that requires compute resources. Apart from functional requirements, we
expect to behave in en efficient way in terms of performance and compute resources usage.

This document guides you on an approach that could be useful to determine performance issues. In particular, we are going 
to focus on memory leaks as an example based on the experience gathered out of [this issue](https://github.com/weaveworks/weave-gitops-enterprise/issues/3189).

### Requirements

- An instance of Weave Gitops Enterprise with Monitoring stack deployed.
- The monitoring stack deployed [Flux Kustomization](https://github.com/weaveworks/weave-gitops-quickstart/tree/add-monitoring) that includes:
- Enabled [metrics](https://docs.gitops.weave.works/docs/references/helm-reference/) 
- Enabled [profiling](https://github.com/weaveworks/weave-gitops-enterprise/blob/b643619464104e59a17e77a697cd7c290f96889a/cmd/clusters-service/app/server.go#L843)

## Detect memory leaks

There could be different ways to detect that you might be facing a memory leak. One of them could have an ever-growing 
memory usage for you container as shown by the following picture:

![memory usage ever growing](imgs/memory-leak-profile.png)

At this point you determine how you memory heap looks like via

1. Download a heap dump for weave gitops by adding `/debug/pprof/heap` to your WGE url (For exampleh ttps://wge-3189-fix.eng-sandbox.weave.works/debug/pprof/heap).
2. Start pprof web interface by `go tool pprof -http=:8082 heap`
3. Navigate to your browser http://localhost:8082/ui/ and a UI like 

![pprof web ui overview](imgs/pprof-web-ui.png)

4. Use the pprof view that better helps you understand these two questions:
 a) what is the function that is generating objects for the heap that are not freed
 b) what is the call chain that ends up calling the function









