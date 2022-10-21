package estimation

import (
	"context"
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Pricer implementations calculate the price for all products matching the
// filters.
type Pricer interface {
	ListPrices(ctx context.Context, service, currency string, filters map[string]string) ([]float32, error)
}

// NewAWSClusterEstimator creates and returns a new AWS estimator that can parse
// price Clusters from resources.
func NewAWSClusterEstimator(pricer Pricer, filters map[string]string) *AWSClusterEstimator {
	return &AWSClusterEstimator{Pricer: pricer, EC2Filters: filters, Currency: "USD"}
}

// AWSClusterEstimator estimates the costs for EC2 instances in AWS Clusters.
type AWSClusterEstimator struct {
	Pricer     Pricer
	EC2Filters map[string]string
	Currency   string
}

// Estimate calculates the estimate for the set of provided resources.
//
// It does this by parsing the resources into a set of clusters with their
// infrastructure and
func (e *AWSClusterEstimator) Estimate(ctx context.Context, us []*unstructured.Unstructured) (map[string]*CostEstimate, error) {
	resources, err := parseResources(us)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	composed, err := composeClusters(resources)
	if err != nil {
		return nil, err
	}
	estimates, err := e.estimateClusters(ctx, composed)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate clusters: %w", err)
	}

	return estimates, nil
}

func (e *AWSClusterEstimator) estimateClusters(ctx context.Context, clusters []cluster) (map[string]*CostEstimate, error) {
	estimates := map[string]*CostEstimate{}
	for _, cluster := range clusters {
		controlPlaneMin, controlPlaneMax, err := e.priceRangeFromFilters(ctx, cluster.controlPlane.instanceType, cluster.regionCode, cluster.controlPlane.instances)
		if err != nil {
			return nil, err
		}

		infrastructureMin, infrastructureMax, err := e.priceRangeFromFilters(ctx, cluster.infrastructure.instanceType, cluster.regionCode, cluster.infrastructure.instances)
		if err != nil {
			return nil, err
		}
		estimates[cluster.name] = &CostEstimate{Low: infrastructureMin + controlPlaneMin, High: infrastructureMax + controlPlaneMax, Currency: e.Currency}
	}

	return estimates, nil
}

func composeClusters(resources *clusterResources) ([]cluster, error) {
	clusters := []cluster{}
	for _, c := range resources.capiClusters {
		infrastructureRegionCode, ok := resources.awsClusters[c.infrastructure.String()]
		if !ok {
			return nil, fmt.Errorf("could not find infrastructure %s", c.infrastructure)
		}
		controlPlane, ok := resources.controlPlanes[c.controlPlane.String()]
		if !ok {
			return nil, fmt.Errorf("could not find control plane %s", c.controlPlane)
		}
		controlPlaneMachineTemplateInstanceType, ok := resources.machineTemplates[controlPlane.machineTemplateRef.String()]
		if !ok {
			return nil, fmt.Errorf("could not find AWSMachineTemplate for control plane %s", c.controlPlane)
		}
		controlPlaneInstances := clusterInstances{
			instances:    *controlPlane.replicas,
			instanceType: controlPlaneMachineTemplateInstanceType,
		}

		infrastructureMachineDeployment, ok := func(s string, deploys map[string]machineDeployment) (machineDeployment, bool) {
			for _, v := range deploys {
				if v.clusterName == s {
					return v, true
				}
			}

			return machineDeployment{}, false
		}(c.name, resources.machineDeployments)
		if !ok {
			return nil, fmt.Errorf("failed to find MachineDeployment for Cluster %s", c.name)
		}

		infrastructureMachineTemplateInstanceType, ok := resources.machineTemplates[infrastructureMachineDeployment.infrastructure.String()]
		if !ok {
			return nil, fmt.Errorf("failed to find AWSMachineTemplate for MachineDeployment %s in cluster %s",
				infrastructureMachineDeployment.infrastructure.name, c.name)
		}

		infrastructureInstances := clusterInstances{
			instances:    infrastructureMachineDeployment.replicas,
			instanceType: infrastructureMachineTemplateInstanceType,
		}

		// This assumes that the infrastructure and control-planes are in the
		// same region code.
		clusters = append(clusters, cluster{
			name: c.name, regionCode: infrastructureRegionCode,
			controlPlane:   controlPlaneInstances,
			infrastructure: infrastructureInstances,
		})
	}

	return clusters, nil
}

func (e *AWSClusterEstimator) priceRangeFromFilters(ctx context.Context, instanceType, regionCode string, instanceCount int32) (float32, float32, error) {
	filters := mergeStringMaps(e.EC2Filters, map[string]string{
		"instanceType": instanceType,
		"regionCode":   regionCode,
	})
	prices, err := e.Pricer.ListPrices(ctx, "AmazonEC2", e.Currency, filters)
	if err != nil {
		return -1, -1, fmt.Errorf("error getting prices for estimation: %w", err)
	}
	totals := []float32{}
	for _, v := range prices {
		monthlyTotal := float32(instanceCount) * v * MonthlyHours
		totals = append(totals, monthlyTotal)
	}
	min, max := minMax(totals)

	return min, max, nil
}

type clusterInstances struct {
	instances    int32
	instanceType string
}

type cluster struct {
	name           string
	regionCode     string
	infrastructure clusterInstances
	controlPlane   clusterInstances
}

type machineDeployment struct {
	infrastructure objectRef
	replicas       int32
	clusterName    string
}

type objectRef struct {
	name string
	schema.GroupVersionKind
}

func (o objectRef) String() string {
	return fmt.Sprintf("%s:%s", o.GroupVersionKind.String(), o.name)
}

func unstructuredKind(u *unstructured.Unstructured) string {
	return u.GetObjectKind().GroupVersionKind().GroupKind().String()
}

func parseObjectRef(u *unstructured.Unstructured, elems ...string) (*objectRef, error) {
	elemMap, _, err := unstructured.NestedStringMap(u.UnstructuredContent(), elems...)
	if err != nil {
		return nil, fmt.Errorf("failed to get infrastructureRef from %s %q: %w", u.GetKind(), u.GetName(), err)
	}
	groupVersion, err := schema.ParseGroupVersion(elemMap["apiVersion"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse infrastructureRef from %s %q: %w", u.GetKind(), u.GetName(), err)
	}

	return &objectRef{
		name:             elemMap["name"],
		GroupVersionKind: groupVersion.WithKind(elemMap["kind"]),
	}, nil
}

// this is used because the unstructured version drops errors.
// it can do this because the resource is in the cluster (normally) and so it's
// already been validated.
// We are parsing unvalidated (by the cluster) resources here.
func nestedString(u *unstructured.Unstructured, elems ...string) (string, error) {
	v, _, err := unstructured.NestedString(u.UnstructuredContent(), elems...)

	return v, err
}

func nestedOptionalInt32(u *unstructured.Unstructured, elems ...string) (*int32, error) {
	v, ok, err := unstructured.NestedInt64(u.UnstructuredContent(), elems...)
	if err != nil {
		return nil, err
	}
	if ok {
		s := int32(v)

		return &s, err
	}

	return nil, nil
}

type clusterResources struct {
	// list of CAPI clusters including the referenced resources
	capiClusters []capiCluster

	// mapping of AWS Cluster -> region
	awsClusters map[string]string

	// mapping of ControlPlane ref to details
	controlPlanes map[string]controlPlane

	// mapping of MachineTemplate ref to instanceType
	machineTemplates map[string]string

	// mapping of MachineDeployment ref to details
	machineDeployments map[string]machineDeployment
}

type capiCluster struct {
	name           string
	infrastructure objectRef
	controlPlane   objectRef
}

type controlPlane struct {
	replicas           *int32
	machineTemplateRef objectRef
}

func parseResources(items []*unstructured.Unstructured) (*clusterResources, error) {
	clusters := []capiCluster{}
	awsClusters := map[string]string{}
	machineTemplates := map[string]string{}
	controlPlanes := map[string]controlPlane{}
	machineDeployments := map[string]machineDeployment{}

	objectKey := func(u *unstructured.Unstructured) string {
		return fmt.Sprintf("%s:%s", u.GroupVersionKind().String(), u.GetName())
	}

	for _, u := range items {
		k := unstructuredKind(u)
		switch k {
		case "Cluster.cluster.x-k8s.io":
			infrastructureRef, err := parseObjectRef(u, "spec", "infrastructureRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse Cluster infrastructureRef %q: %w", u.GetName(), err)
			}
			controlPlaneRef, err := parseObjectRef(u, "spec", "controlPlaneRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse Cluster controlPlaneRef %q: %w", u.GetName(), err)
			}
			clusters = append(clusters, capiCluster{
				name:           u.GetName(),
				infrastructure: *infrastructureRef,
				controlPlane:   *controlPlaneRef,
			})
		case "AWSCluster.infrastructure.cluster.x-k8s.io":
			regionCode, err := nestedString(u, "spec", "region")
			if err != nil {
				return nil, fmt.Errorf("failed to parse AWSCluster %q: %w", u.GetName(), err)
			}
			awsClusters[objectKey(u)] = regionCode
		case "KubeadmControlPlane.controlplane.cluster.x-k8s.io":
			machineTemplateRef, err := parseObjectRef(u, "spec", "machineTemplate", "infrastructureRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse KubeadmControlPlane infrastructureRef %q: %w", u.GetName(), err)
			}
			replicas, err := nestedOptionalInt32(u, "spec", "replicas")
			if err != nil {
				return nil, fmt.Errorf("failed to parse KubeadmControlPlane replicas %q: %w", u.GetName(), err)
			}
			controlPlanes[objectKey(u)] = controlPlane{
				replicas:           replicas,
				machineTemplateRef: *machineTemplateRef,
			}
		case "AWSMachineTemplate.infrastructure.cluster.x-k8s.io":
			instanceType, err := nestedString(u, "spec", "template", "spec", "instanceType")
			if err != nil {
				return nil, fmt.Errorf("failed to parse AWSCluster %q: %w", u.GetName(), err)
			}
			machineTemplates[objectKey(u)] = instanceType
		case "MachineDeployment.cluster.x-k8s.io":
			replicas, err := nestedOptionalInt32(u, "spec", "replicas")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachineDeployment %q: %w", u.GetName(), err)
			}
			infrastructureRef, err := parseObjectRef(u, "spec", "template", "spec", "infrastructureRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse KubeadmControlPlane infrastructureRef %q: %w", u.GetName(), err)
			}
			clusterName, err := nestedString(u, "spec", "clusterName")
			if err != nil {
				return nil, fmt.Errorf("failed to find clusterName in MachineDeployment %s", u.GetName())
			}
			machineDeployments[objectKey(u)] = machineDeployment{
				replicas: *replicas, infrastructure: *infrastructureRef,
				clusterName: clusterName,
			}
		}
	}

	return &clusterResources{
		capiClusters:       clusters,
		awsClusters:        awsClusters,
		controlPlanes:      controlPlanes,
		machineTemplates:   machineTemplates,
		machineDeployments: machineDeployments,
	}, nil
}

func mergeStringMaps(origin, update map[string]string) map[string]string {
	cloned := map[string]string{}
	for k, v := range origin {
		cloned[k] = v
	}
	for k, v := range update {
		if v != "" {
			cloned[k] = v
		}
	}

	return cloned
}

func minMax(vals []float32) (float32, float32) {
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })

	return vals[0], vals[len(vals)-1]
}
