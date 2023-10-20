# Explorer Indexer

Here we document the Indexer component in Explorer based on the following questions:

- why an indexer
- what indexer uses explorer
- how indexer for explorer works

# Why an indexer

# What indexer uses explorer

Given an index is there for fast queries, we want to ensure that the indexer is designed on purposes for those queries. Therefore,
we are going to start by selecting a set of query scenarios that matters to us to drive the rest of the understanding.

1) As developer, you own the application `shopping-cart` and you want to assess the health of all the resources for your application
across the platform. To discover resources you have for compliance to use a label `app=<application-name>` for each resource, so would 
like to introduce `app=shopping-cart` and to show you all the application like  

```bash 
k get all -l app=registry -A                                                                                                                          
NAMESPACE   NAME                                   READY   STATUS    RESTARTS   AGE
gitlab      pod/gitlab-registry-5f75ff9f95-gxfbd   1/1     Running   0          40m

NAMESPACE   NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
gitlab      service/gitlab-registry   ClusterIP   10.96.38.109   <none>        5000/TCP   40m

NAMESPACE   NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
gitlab      deployment.apps/gitlab-registry   1/1     1            1           40m

NAMESPACE   NAME                                         DESIRED   CURRENT   READY   AGE
gitlab      replicaset.apps/gitlab-registry-5f75ff9f95   1         1         1       40m

NAMESPACE   NAME                                                  REFERENCE                    TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
gitlab      horizontalpodautoscaler.autoscaling/gitlab-registry   Deployment/gitlab-registry   <unknown>/75%   1         1         1          40m
```

This case applies where the user knows the dimension to search as key, and value or likely value to search. this is the scenario of key/value search. to any search whether the user knows a key and value to search  

2) As platform engineer, you want to discover all applications you are doing support and have been notified of an incident about deployment failures. You 
want to discover deployment failure across the platform searching for `failed`. 

Or As compliance, you want to find which ocirpositoreis are verified or not https://fluxcd.io/flux/components/source/ocirepositories/#verification,
so you are going to search for `verified`.

Or all apps with a commitID `74c60a927c7588900c28960d37ff2e5118d0eedf`
 
In this case you want to search by the term across any dimension: this is the case for full text search

Therefore, we have a set of type of queries:

- `key/value`: `app=shopping-cart`
- `full-text`: search all `failed`, `verified` or `74c60a927c7588900c28960d37ff2e5118d0eedf`

We could compose those searches with multiple `key\value` and `full-text` with the results as the intersection of all the queries.

# How indexer for explorer works:

Given that we know now the queries we want to serve lets look at 

- indexing: how the document and indexing looks like to serve those queries
- querying: how querying looks like for satisfying those searches
- navigation: how the user would traverse the collection and or filter 

## Indexing: how the document and indexing looks like to serve those queries

We index in https://github.com/weaveworks/weave-gitops-enterprise/blob/2fdb9b9455787f5a0c5469556f366f72ddbba890/pkg/query/store/indexer.go#L109

Our document is called `object` https://github.com/weaveworks/weave-gitops-enterprise/blob/d502edf7b80e622800835c27d91deb2b78dc70be/pkg/query/internal/models/object.go#L15

```go
type Object struct {
   gorm.Model
   ID                  string                       `gorm:"primaryKey;autoIncrement:false"`
   Cluster             string                       `json:"cluster" gorm:"type:text"`
   Namespace           string                       `json:"namespace" gorm:"type:text"`
   APIGroup            string                       `json:"apiGroup" gorm:"type:text"`
   APIVersion          string                       `json:"apiVersion" gorm:"type:text"`
   Kind                string                       `json:"kind" gorm:"type:text"`
   Name                string                       `json:"name" gorm:"type:text"`
   Status              string                       `json:"status" gorm:"type:text"`
   Message             string                       `json:"message" gorm:"type:text"`
   Category            configuration.ObjectCategory `json:"category" gorm:"type:text"`
   KubernetesDeletedAt time.Time                    `json:"kubernetesDeletedAt"`
   Unstructured        json.RawMessage              `json:"unstructured" gorm:"type:blob"`
   Tenant              string                       `json:"tenant" gorm:"type:text"`
}
```

And the indexer is created https://github.com/weaveworks/weave-gitops-enterprise/blob/2fdb9b9455787f5a0c5469556f366f72ddbba890/pkg/query/store/indexer.go#L56 

with 

```go

var indexFile = "index.db"
var filterFields = []string{"cluster", "namespace", "kind"}

func NewIndexer(s Store, path string, log logr.Logger) (Indexer, error) {
	idxFileLocation := filepath.Join(path, indexFile)
	mapping := bleve.NewIndexMapping()

	addFieldMappings(mapping, filterFields)
```

:warning: 

https://blevesearch.com/docs/Index-Mapping/

>IndexMappings contain DocumentMappings for each of the different types of documents you want to support. Further, it contains a DefaultDocumentMapping that will be used for any type which does not have an explicit mapping.

Do we need different document mappings?

- no for normalised objects -> we have a single type 
- might be for denormalized object -> as the indexing category will depend on

An example here is what we want to do with templates:

- a gitopstemplate has a type that comes from a label that would require a document mapping itself
- other resources would be fine 

>FieldMappings
>Documents are hierarchical and contain named fields. These fields could be values or nested sub-documents. We customize the behavior for a named field by setting a DocumentMapping for it. Once we have a DocumentMapping for the named field, we can attach 0 or more FieldMappings to it. The FieldMappings describe how we want the field to be interpreted and what we want inserted into the index.


An strategy: 
- have a default documentmapping for json documents that allows 


https://github.com/weaveworks/weave-gitops-private/pull/132#issuecomment-1773001598


## Querying: how querying looks like for satisfying those searches


### key-value search 

For normalised filter 

- Give me all resources by name. For example `name=shopping-cart` where name is the key and shopping-cart is the value. 

Navigation via It comes as part of the sear 



For denormalized filter 

- Give me all resources by label. For example `app=shopping-cart` where app is the key and shopping-cart is the value.


### Simple full-text

- Give me all resources with commitId. For example `74c60a927c7588900c28960d37ff2e5118d0eedf` that is the commitId that could appear in any part of the document. 

### Simple 






## Navigation: how the user would traverse the collection and or filter








Here is a quick and dirty guide for adding new objects to Explorer.

Things to know:

- All objects are stored in a central store in a single table
- All objects are normalized to a standard format
- To differentiate which objects should show up for a given request, we use the `category` field
- Adding a new object kind creates a new `watch` that will listen for updates on all clusters
- Here is [an architecture doc](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/docs/architecture/explore.md) that goes into detail

## Adding a new Object Kind

1. Add a new entry in the [`SupportedObjectKinds`](https://github.com/weaveworks/weave-gitops-enterprise/blob/253256c16c777b0d488ca0ba8068b8f80b1b4c07/pkg/query/configuration/objectkind.go#L119) slice
2. (Optional) Add a entry to [`ToFluxObject`](https://github.com/weaveworks/weave-gitops-enterprise/blob/f36d549b6010afbd3c086c4955637586629ec589/pkg/query/configuration/objectkind.go#L284) if your 
kind manages to meet the [FluxObject interface](https://github.com/weaveworks/weave-gitops-enterprise/blob/9534aa348ac40928e18fe741de0c7b3c0bb89d14/pkg/query/configuration/objectkind.go#L83)
3. Add an [RBAC entry](https://github.com/weaveworks/weave-gitops-enterprise/blob/9101b60a487e1f999b4e988e9ca27bdde4ac7538/charts/mccp/templates/clusters-service/collector.yaml#L13) to WeGO ServiceAccount for your kind

## Using the default Explorer UI component

If you would like to use default Explorer view with your Kind:

1. Add your Kind details Route to [`getKindRoute`](https://github.com/weaveworks/weave-gitops-enterprise/blob/f36d549b6010afbd3c086c4955637586629ec589/ui-cra/src/utils/nav.ts#L3)
2. Add a test case for your kind in the [integration test](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/pkg/query/server/server_integration_test.go#L44)
   - Add any CRDs to [testdata/crds](../pkg/query/server/testdata/crds)
   - Add the GV to [`AddToScheme`](../pkg/query/server/suite_test.go) to register the GVK.
3. Add Kind or Resource to the [user documentation](https://docs.gitops.weave.works/docs/explorer/configuration/#kinds).  
   
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
