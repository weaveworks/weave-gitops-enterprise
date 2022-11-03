package estimation

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Pricer implementations calculate the price for all products matching the
// filters.
type Pricer interface {
	ListPrices(ctx context.Context, service, currency string, filters map[string]string) ([]float32, error)
}

var _ Estimator = (*AWSClusterEstimator)(nil)

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
func (e *AWSClusterEstimator) Estimate(ctx context.Context, us []*unstructured.Unstructured) (*CostEstimate, error) {
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

	return reduceEstimates(estimates), nil
}

func (e *AWSClusterEstimator) estimateClusters(ctx context.Context, clusters []composedCluster) (map[string]*CostEstimate, error) {
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

func composeClusters(resources *clusterResources) ([]composedCluster, error) {
	clusters := []composedCluster{}
	for _, cluster := range resources.capiClusters {
		infrastructureRegionCode, ok := resources.awsClusters[cluster.infrastructure.String()]
		if !ok {
			return nil, fmt.Errorf("could not find infrastructure %s", cluster.infrastructure)
		}
		controlPlane, ok := resources.controlPlanes[cluster.controlPlane.String()]
		if !ok {
			return nil, fmt.Errorf("could not find control plane %s", cluster.controlPlane)
		}
		controlPlaneMachineTemplateInstanceType, ok := resources.machineTemplates[controlPlane.machineTemplate.String()]
		if !ok {
			return nil, fmt.Errorf("could not find AWSMachineTemplate for control plane %s", cluster.controlPlane)
		}
		controlPlaneInstances := clusterInstances{
			instances:    *controlPlane.replicas,
			instanceType: controlPlaneMachineTemplateInstanceType,
		}

		infrastructureMachineDeployment, foundMD := func(s string, deploys map[string]machineDeployment) (machineDeployment, bool) {
			for _, v := range deploys {
				if v.clusterName == s {
					return v, true
				}
			}

			return machineDeployment{}, false
		}(cluster.name, resources.machineDeployments)

		infrastructureMachinePool, foundPool := func(s string, pools map[string]machinePool) (machinePool, bool) {
			for _, v := range pools {
				if v.clusterName == s {
					return v, true
				}
			}

			return machinePool{}, false
		}(cluster.name, resources.machinePools)

		if !foundMD && !foundPool {
			return nil, fmt.Errorf("failed to find MachineDeployment or MachinePool for Cluster %s", cluster.name)
		}

		var infrastructureInstances clusterInstances

		if foundMD {
			infrastructureMachineTemplateInstanceType, ok := resources.machineTemplates[infrastructureMachineDeployment.infrastructure.String()]
			if !ok {
				return nil, fmt.Errorf("failed to find AWSMachineTemplate for MachineDeployment %s in cluster %s",
					infrastructureMachineDeployment.infrastructure.name, cluster.name)
			}

			infrastructureInstances = clusterInstances{
				instances:    infrastructureMachineDeployment.replicas,
				instanceType: infrastructureMachineTemplateInstanceType,
			}
		}

		if foundPool {
			// TODO: This should check that the infrastructure is an
			// AWSMachinepool
			pool, ok := resources.awsMachinePools[infrastructureMachinePool.infrastructure.String()]
			if !ok {
				return nil, fmt.Errorf("failed to find AWSMachinePool for MachinePool %s in cluster %s", infrastructureMachinePool.infrastructure, cluster.name)
			}

			infrastructureInstances = clusterInstances{
				instances:    infrastructureMachinePool.replicas,
				instanceType: pool.instanceType,
			}
		}

		// This assumes that the infrastructure and control-planes are in the
		// same region code.
		clusters = append(clusters, composedCluster{
			name: cluster.name, regionCode: infrastructureRegionCode,
			controlPlane:   controlPlaneInstances,
			infrastructure: infrastructureInstances,
		})
	}

	return clusters, nil
}

func (e *AWSClusterEstimator) priceRangeFromFilters(ctx context.Context, instanceType, regionCode string, instanceCount int32, additionalFilters ...map[string]string) (float32, float32, error) {
	unmergedFilters := append([]map[string]string{e.EC2Filters}, additionalFilters...)
	unmergedFilters = append(unmergedFilters, map[string]string{
		"instanceType": instanceType,
		"regionCode":   regionCode,
	})
	filters := mergeStringMaps(unmergedFilters...)
	prices, err := e.Pricer.ListPrices(ctx, "AmazonEC2", e.Currency, filters)
	if err != nil {
		return invalidPrice, invalidPrice, fmt.Errorf("error getting prices for estimation: %w", err)
	}
	totals := []float32{}
	for _, v := range prices {
		monthlyTotal := float32(instanceCount) * v * MonthlyHours
		totals = append(totals, monthlyTotal)
	}
	if len(totals) == 0 {
		return invalidPrice, invalidPrice, fmt.Errorf("no price data returned for instanceType %s in region %s", instanceType, regionCode)
	}
	min, max := minMax(totals)

	return min, max, nil
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
	if elemMap == nil {
		return nil, fmt.Errorf("missing reference: %s", strings.Join(elems, "."))
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

type clusterInstances struct {
	instances    int32
	instanceType string
}

type composedCluster struct {
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

type capiCluster struct {
	name           string
	infrastructure objectRef
	controlPlane   objectRef
}

type controlPlane struct {
	replicas        *int32
	machineTemplate objectRef
}

type machinePool struct {
	replicas       int32
	infrastructure objectRef
	clusterName    string
}

type awsMachinePool struct {
	maxSize      int32
	instanceType string
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

	// mapping of MachinePool ref to details
	machinePools map[string]machinePool

	// mapping of AWSMachinePool ref to details
	awsMachinePools map[string]awsMachinePool
}

func parseResources(items []*unstructured.Unstructured) (*clusterResources, error) {
	clusters := []capiCluster{}
	awsClusters := map[string]string{}
	machineTemplates := map[string]string{}
	controlPlanes := map[string]controlPlane{}
	machineDeployments := map[string]machineDeployment{}
	machinePools := map[string]machinePool{}
	awsMachinePools := map[string]awsMachinePool{}

	objectKey := func(u *unstructured.Unstructured) string {
		return fmt.Sprintf("%s:%s", u.GroupVersionKind().String(), u.GetName())
	}

	// TODO: validate missing fields
	// Go through the API definitions and check which of the fields
	// pulled below are not optional
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
				replicas:        replicas,
				machineTemplate: *machineTemplateRef,
			}
		case "AWSMachineTemplate.infrastructure.cluster.x-k8s.io":
			instanceType, err := nestedString(u, "spec", "template", "spec", "instanceType")
			if err != nil {
				return nil, fmt.Errorf("failed to parse AWSMachineTemplate %q: %w", u.GetName(), err)
			}
			machineTemplates[objectKey(u)] = instanceType
		case "MachineDeployment.cluster.x-k8s.io":
			replicas, err := nestedOptionalInt32(u, "spec", "replicas")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachineDeployment %q: %w", u.GetName(), err)
			}
			infrastructureRef, err := parseObjectRef(u, "spec", "template", "spec", "infrastructureRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachineDeployment infrastructureRef %q: %w", u.GetName(), err)
			}
			clusterName, err := nestedString(u, "spec", "clusterName")
			if err != nil {
				return nil, fmt.Errorf("failed to find clusterName in MachineDeployment %s", u.GetName())
			}
			machineDeployments[objectKey(u)] = machineDeployment{
				replicas: *replicas, infrastructure: *infrastructureRef,
				clusterName: clusterName,
			}
		case "MachinePool.cluster.x-k8s.io":
			optionalReplicas, err := nestedOptionalInt32(u, "spec", "replicas")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachinePool - missing replicas%q: %w", u.GetName(), err)
			}
			// v1beta1 MachinePool defaults to 1 if not provided
			replicas := int32(1)
			if optionalReplicas != nil {
				replicas = *optionalReplicas
			}
			clusterName, err := nestedString(u, "spec", "clusterName")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachinePool %q: %w", u.GetName(), err)
			}
			infrastructureRef, err := parseObjectRef(u, "spec", "template", "spec", "infrastructureRef")
			if err != nil {
				return nil, fmt.Errorf("failed to parse MachinePool infrastructureRef %q: %w", u.GetName(), err)
			}
			machinePools[objectKey(u)] = machinePool{
				replicas:       replicas,
				clusterName:    clusterName,
				infrastructure: *infrastructureRef,
			}
		case "AWSMachinePool.infrastructure.cluster.x-k8s.io":
			optionalMaxSize, err := nestedOptionalInt32(u, "spec", "maxSize")
			if err != nil {
				return nil, fmt.Errorf("failed to parse AWSMachinePool maxSize%q: %w", u.GetName(), err)
			}
			// v1beta1 AWSMachinePool defaults to 1 if not provided
			maxSize := int32(1)
			if optionalMaxSize != nil {
				maxSize = *optionalMaxSize
			}
			// TODO: instanceType is not required in the AWSLaunchTemplate it's
			// not clear what to do in this case.
			instanceType, err := nestedString(u, "spec", "awsLaunchTemplate", "instanceType")
			if err != nil {
				return nil, fmt.Errorf("failed to parse AWSMachinePool %q: %w", u.GetName(), err)
			}

			awsMachinePools[objectKey(u)] = awsMachinePool{
				maxSize:      maxSize,
				instanceType: instanceType,
			}
		}
	}

	return &clusterResources{
		capiClusters:       clusters,
		awsClusters:        awsClusters,
		controlPlanes:      controlPlanes,
		machineTemplates:   machineTemplates,
		machineDeployments: machineDeployments,
		machinePools:       machinePools,
		awsMachinePools:    awsMachinePools,
	}, nil
}
