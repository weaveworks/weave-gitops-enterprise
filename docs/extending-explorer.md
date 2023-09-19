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
2. Add a entry to [`ToFluxObject`](https://github.com/weaveworks/weave-gitops-enterprise/blob/f36d549b6010afbd3c086c4955637586629ec589/pkg/query/configuration/objectkind.go#L284) as will be required 
for [Status and Message](https://github.com/weaveworks/weave-gitops-enterprise/blob/f36d549b6010afbd3c086c4955637586629ec589/pkg/query/internal/models/object.go#L113C1-L129)
3. Add an [RBAC entry](https://github.com/weaveworks/weave-gitops-enterprise/blob/9101b60a487e1f999b4e988e9ca27bdde4ac7538/charts/mccp/templates/clusters-service/collector.yaml#L13) to WeGO ServiceAccount for your kind

## Using the default Explorer UI component

If you would like to use default Explorer view with your Kind:

1. Add your Kind details Route to [`getKindRoute`](https://github.com/weaveworks/weave-gitops-enterprise/blob/f36d549b6010afbd3c086c4955637586629ec589/ui-cra/src/utils/nav.ts#L3)
2. Add a test case for your kind in the [integration test](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/pkg/query/server/server_integration_test.go#L44)
   - Add any CRDs to [testdata/crds](../pkg/query/server/testdata/crds)
   - Add the GV to [`AddToScheme`](../pkg/query/server/suite_test.go) to register the GVK.
3. Add the Kind or Resource to the [user documentation](https://docs.gitops.weave.works/docs/explorer/configuration/#kinds).  
   
## Using the Explorer UI component

If you would like to view your object list in a dedicated table:

1. Add a new [`ObjectCategory`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/internal/models/object.go#L13) variable
2. Add a category to the `ObjectKind` struct that is being added to `SupportedObjectKinds`

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
