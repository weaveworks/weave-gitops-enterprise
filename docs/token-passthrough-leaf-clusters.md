# Token passthrough for leaf clusters

## Note

_This is not a publicly available feature. It was developed for Deutsche Telekom (DT) and their particular k8s-cluster auth setup._

## DT's k8s setup

As of 2023-01-09 DT configures OIDC on all of their k8s clusters' api-servers. The OIDC clientID and clientSecret are the same for all clusters.

This is not ideal and not an approach we want to promote. See https://gist.github.com/bigkevmcd/7492b5d67dd1a9edb5f46b1e2f47cf82

DT wants to avoid configuring a ServiceAccount for each cluster and instead use the user's OIDC token to authenticate to the leaf clusters. This works with some technical caveats mentioned in this document below.

## Configuration

The following configuration is required to enable token passthrough for leaf clusters:

```yaml
oidc:
  enabled: true
  # fill in your issuer URL
  issuerURL: "https://dev-dox5prxhgkda6bz8.us.auth0.com/"
  # fill in the address you access the UI with
  redirectURL: "http://localhost:9001/oauth2/callback"
  # choose the username and group claims to use
  claimUsername: "email"
  # defaults to "groups"
  claimGroups: ""
  # Name of secret in flux-system namespace that contains a clientId and clientSecret
  clientCredentialsSecret: "oidc-auth"
  # Customise the requested scopes for then OIDC authentication flow - openid will always be requested
  customScopes: ""
auth:
  # disable user-account authentication (username/password)
  userAccount:
    enabled: false
  # enable token passthrough
  tokenPassthrough:
    enabled: true

envVars:
  # enable using the token passthrough to derive available namespaces
  - name: WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES
    value: "true"
```

The important part is the `WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES` env var.

Make sure to also save the `clientId` and `clientSecret` in the `oidc-auth` secret in the `flux-system` namespace.

```
kubectl create secret generic --namespace flux-system oidc-auth \
 --from-literal=clientID=MY_CLIENT_ID \
 --from-literal=clientSecret=MY_CLIENT_SECRET
```

Once enabled the current user's OIDC token will be used to authenticate and _list the namespaces_ for _all_ clusters, including the management cluster.

The management cluster's service account will no longer be used to request the full list of namespaces for any cluster including itself.

## Caveats

- All OIDC users must have RBAC permissions to list namespaces on any GitopsCluster they have access to.
- The _Add application_ feature lets you install HelmReleases onto management and leaf clusters. It caches the list of available helm charts using the management cluster's SA in the background as this is an expensive operation. The index.yaml of HelmRepo can be many megabytes.
  - This works for the management cluster as its SA will have permissions to cache the list of available charts.
  - This will not work for leaf clusters as the SA does not have access to them, only the user will have access at request time.
