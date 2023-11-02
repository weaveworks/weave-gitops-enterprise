# Explorer Indexer

You are here cause you want to understand a bit more on the indexer component in Explorer with any of the following questions:

- Why an Indexer
- What Indexer uses Explorer
- How Indexer for Explorer works

# Glossary

Based on https://blevesearch.com/docs/Terminology/

>Term:A term is a sequence of unicode characters. Typically the word “term” is reserved for uses describing the things we write into indexes or the things we’re looking for in indexes. For example, the text “mary had a little lamb”, might result in 3 terms being inserted into the index: “mary”, “little”, and “lamb”.

# Why an indexer

Index-based searching is at core of any modern search experience as it provides performant, flexible and powerful searching
approach. Before introducing an indexer, Explorer was querying based on SQL-like semantics which has evident issues when 
used as search engine. 

# What indexer uses Explorer

Given an index is there for fast queries, we want to ensure that the indexer is designed on purposes for those queries. Therefore,
we are going to start by selecting a set of query scenarios that matters to us to drive the rest of the understanding.

1) As a developer, you own the application `shopping-cart` and you want to assess the health of all the resources belonging to it. 
To discover resources, you have annotated your resources with the label `app=shopping-cart`, so you would like Explorer to provide
you with a feature similar to kubectl of filtering by label:  

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

This case applies where the user knows the dimension to search (key) and value.  

2) As a platform engineer, you are doing on-call. You have been notified that are deployment failures. You 
want to discover deployment failure across the platform searching for `failed`. In this case you want to search by the term across any dimension. 
This is the case for full-text search. Other examples of this type of scenario could be:

- You want to find which OCI repositories are verified or not https://fluxcd.io/flux/components/source/ocirepositories/#verification, so you are going to search for `verified`.
- You want to find all apps with a commitID `74c60a927c7588900c28960d37ff2e5118d0eedf`.

These two cases set the type of queries that we want to support:

1. `key/value`: `app=shopping-cart`
2. `full-text`: search all `failed`, `verified` or `74c60a927c7588900c28960d37ff2e5118d0eedf`

We could compose those searches with multiple `key\value` and `full-text` with the results as the intersection of all the queries.

In Explorer, we use https://blevesearch.com/ as indexing library.

# How indexer for Explorer works

Given that we know now the queries we want to serve lets look at: 

- Indexing: how the document and indexing looks like to serve those queries
- Querying or Search: how querying looks like for satisfying those searches
- Navigation: how the user would traverse the collection and or filter 

## Indexing: how get documents in

We index documents [here](https://github.com/weaveworks/weave-gitops-enterprise/blob/2fdb9b9455787f5a0c5469556f366f72ddbba890/pkg/query/store/indexer.go#L109). 
Our document is called [`object`](https://github.com/weaveworks/weave-gitops-enterprise/blob/d502edf7b80e622800835c27d91deb2b78dc70be/pkg/query/internal/models/object.go#L15):

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

The indexer is created [here](https://github.com/weaveworks/weave-gitops-enterprise/blob/2fdb9b9455787f5a0c5469556f366f72ddbba890/pkg/query/store/indexer.go#L56). 
When you index a document, you create an inverted index (like the [book index](https://en.wikipedia.org/wiki/Index_(publishing))) to the document where the key is the field. 
Notice that a `field` here is an `indexed field`. It is important to remark this as search operations by fields should be done referencing to `indexed fields` and not `object fields`:

You could understand this difference by for example using [bleve cli](https://blevesearch.com/docs/bleve/) that has 
features to list the indexed fields:

```
✗ bleve fields /var/folders/9b/bkrspzws5xgd7x_ldtc880pr0000gn/T/index735061655/index.db

0 - Object.spec.renderType
1 - Object.metadata.labels.templateType
2 - Object.kind
3 - Object.metadata.managedFields.manager
4 - Object.metadata.managedFields.operation
5 - Object.metadata.namespace
6 - category
7 - kind.facet
8 - namespace.facet
9 - _id
10 - ID
11 - Object.apiVersion
12 - Object.metadata.labels.weave.works/template-type
13 - _all
14 - labels.value
15 - DeletedAt.Valid
16 - Object.metadata.generation
17 - Object.metadata.managedFields.apiVersion
18 - Object.metadata.name
19 - Object.metadata.resourceVersion
20 - apiGroup
21 - cluster.facet
22 - labels.key
23 - Object.metadata.creationTimestamp
24 - name
25 - Object.metadata.uid
26 - kind
27 - unstructured
28 - Object.metadata.managedFields.time
29 - tenant
30 - Object.metadata.managedFields.fieldsType
31 - cluster
32 - message
33 - apiVersion
34 - namespace
35 - status
36 - Object.metadata.selfLink
```

In our case they came from two documents:

1. `object` document

```
        if err := batch.Index(obj.GetID(), obj); err != nil {
			i.log.Error(err, "failed to index object", "object", obj.GetID())
			continue
		}
```

Whose field mapping creates fields like:

```
7 - kind.facet

26 - kind
```

2. `object.unstructured` field

```
            if err := batch.Index(obj.GetID()+unstructuredSuffix, data); err != nil {
				i.log.Error(err, "failed to index unstructured object", "object", obj.GetID())
				continue
			}
```

whose field mapping creates fields like:

```
2 - Object.kind
```

## Search: how we find data 

You could do two type of searches:

- `key-value`: when you want to search those documents whose `field` (key) has a given `value`
- `full-text`: when you want to search anywhere in the index for `terms`

### Key-Value searches

Key-value searches are supported by the indexer where:
- `key` is the path of an indexed field
- `value` is the dictionary term to search for the indexed field

Querying for those fields would give us the same values. For example:

![query-object.png](imgs%2Fquery-object.png)
![query-object-unstructured.png](imgs%2Fquery-object-unstructured.png)

### Full-text search or term search

Happens when you don't indicate the field to use but the whole index.

![query-object-term.png](imgs%2Fquery-object-term.png)

## Navigation: how traverse a collection

TBA

## Explorer Search Features
With indexing and search we are able to explore some of the features that we would provide or think 
to provide in Explorer and how they could be brought on.

### Search by labels

This is the case requested for templates where the type is a label `weave.works/template-type`.

Searching by label means that:

1. Explorer api accepts searches by label 
2. Explorer ui offers filtering by label
3. Explorer UI offers customisation for better user experience

We will be using the following template:

```go
    &gapiv1.GitOpsTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       gapiv1.Kind,
						APIVersion: "templates.weave.works/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-template-1",
						Namespace: "default",
						Labels: map[string]string{
							"weave.works/template-type": "cluster",
						},
					},
				},
```

#### Explorer api accepts searches by label

To offer that we require that:

1. labels are indexed
2. query api endpoint accepts label-based-fields queries

**Labels are indexed**

As we have seen earlier, any unstructured fields is indexed that includes the label, so 
we meet this point:

```
10 - ID
11 - Object.apiVersion
12 - Object.metadata.labels.weave.works/template-type
13 - _all
14 - labels.value
```

if we get the values for that field 

```bash
bleve dictionary /var/folders/9b/bkrspzws5xgd7x_ldtc880pr0000gn/T/index735061655/index.db Object.metadata.labels.weave.works/template-type
cluster - 1
```
**Query api endpoint accepts label-based-fields queries**

We have created a test case in `server_integration_test.go` for doing where a query like `Object.metadata.labels.weave.works/template-type:cluster`

```go
{
			name:   "should support gitops templates by label",
			access: allowTemplatesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				&gapiv1.GitOpsTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       gapiv1.Kind,
						APIVersion: "templates.weave.works/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-template-1",
						Namespace: "default",
						Labels: map[string]string{
							"weave.works/template-type": "cluster",
						},
					},
				},
			},
			query:              "Object.metadata.labels.weave.works/template-type:cluster",
			expectedNumObjects: 1,
		},
```

```bash
=== RUN   TestQueryServer/should_support_gitops_templates_by_label
--- PASS: TestQueryServer/should_support_gitops_templates_by_label (1.46s)
PASS
````

#### Explorer ui offers filtering by label

Explorer offers pre-defined filters by using bleve feature facets
https://blevesearch.com/docs/Result-Faceting/

>Facets allow you to include aggregated information about the documents matching your query.

Facets are just indexed fields that shows its dictionary term, so it could be used to create key-value queries for particular <indexed field , term>

Explorer UI calculates the faces by using `ListFacets` api endpoint. We could 
get the label as facet by adding them to the kind configuration and requesting 
for them during `ListFacets` request as follows:

```
	// adding facets for labels
	for _, objectKind := range configuration.SupportedObjectKinds {
		for _, label := range objectKind.Labels {
			labelFacet := fmt.Sprintf("Object.metadata.labels.%s", label)
			req.AddFacet(labelFacet, bleve.NewFacetRequest(labelFacet, 100))
		}
	}
```
Then we could see the fields showing as facet 

![Explorer-label-facet.png](imgs%2Fexplorer-label-facet.png)

The current limitation is that the visualization shows
the field name completely `Object.metadata.labels.weave.works/template-type`
where the user should see `weave.works/template-type`

the reasons are:

1) we are using unstructured so we have the full indexed field that includes the path 
2) we are not using the non-unstructured

the solution:

1) **by unstructured**: the api should do a mapping between indexed fields and api fields in and out from api queries
- for an api query request -> we add the indexed field path
- for an api query response -> we delete the indexed field path

2) **by normalised**: add `label` as part of the normalised object. this would likely require to:
 - add the label to the normalised schema 
 - add a field mapping for `label` has as `indexed field id` = `label-key`

**solution by normalised**:

Added labels to object struct as map 

```
...
	Tenant              string                       `json:"tenant" gorm:"type:text"`
	Labels              map[string]string            `json:"labels" gorm:"-"`
}
```
Now we have the same situation as before for unstructured with indexed field like `labels.labelKey` 

```
 ✗ bleve dictionary /Users/enekofb/projects/github.com/blevesearch/bleve-Explorer/data/index.db labels.weave.works/template-type
cluster - 1
```

and we could filter and querying using the label 

![Explorer-label-facet-normalised.png](imgs%2Fexplorer-label-facet-normalised.png)

#### Explorer UI offers customisation for better user experience

To be discussed within Tangerine if the previous looks fine