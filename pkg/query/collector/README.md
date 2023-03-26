# Collector 

# Get Started



# FAQ

## What is?

It is a component part of wego search or query engine that 
- watches gvks from remote clusters
- adapts them to a common schema 
- stores it in an indexer  

## What kinds are supported?

TBA

## How the common schema looks like?

TBA

## How can I add a new kind?

1. add a case to the [acceptance tests](collector_acceptance_test.go)
2. add a client object to the [reconciler](./reconciler/reconciler.go)
```go
func GetClientObjectByKind(gvk schema.GroupVersionKind) (client.Object, error) {
	switch gvk.Kind {
    ...
	case "ClusterRole":
		return &rbacv1.ClusterRole{}, nil
	default:
		return nil, fmt.Errorf("gvk not supported: %s", gvk.Kind)
	}
	return nil, fmt.Errorf("invalid gvk")
}
```




