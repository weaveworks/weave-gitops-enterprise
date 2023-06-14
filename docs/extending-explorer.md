# Extending Explorer

Here is a quick and dirty guide for adding new objects to Explorer.

Things to know:

- All objects are stored in a central store in a single table
- All objects are normalized to a standard format
- To differentiate which objects should show up for a given request, we use the `category` field
- Adding a new object kind creates a new `watch` that will listen for updates on all clusters
- Here is [an architecture doc](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/docs/architecture/explore.md) that goes into detail

## Adding a new Object Kind

1. Add a new entry in the [`SupportedObjectKinds`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/configuration/objectkind.go#L119) slice
2. Add an entry to the [`flux` adapters `switch`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/internal/adapters/fluxobject.go#L27)
3. Add an [RBAC entry](https://github.com/weaveworks/weave-gitops-enterprise/blob/9101b60a487e1f999b4e988e9ca27bdde4ac7538/charts/mccp/templates/clusters-service/collector.yaml#L13) to WeGO ServiceAccount for your kind

## Using the Explorer UI component

If you would like to view your object list in a dedicated table:

1. Add a new [`ObjectCategory`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/internal/models/object.go#L13) variable
2. Add a category entry to the [`adapters`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/internal/adapters/fluxobject.go#L72)

3. Instantiate the `<Explorer />` UI component with the `category` prop, like so:

```jsx
function MyComponent() {
  return (
    <div>
      <Explorer category="gitopssets" />
    </div>
  );
}
```