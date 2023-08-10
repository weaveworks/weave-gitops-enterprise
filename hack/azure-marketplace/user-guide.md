# User guide

We'll take you through installing EE on azure

## 1. Enable the GitOps cluster extension on your AKS cluster

Navigate to cluster extensions in the Azure portal and add the _GitOps_ extension to your AKS Cluster.

## 2. Install a CAPI provider

For example:

```
clusterctl init --infrastructure azure
```

If you don't plan to use CAPI, install the vcluster provider that is quite light, it still provides some CRDs that EE requires

```
clusterctl init --infrastructure vcluster
```

## 3. Apply the entitlements secret

Contact sales@weave.works for a valid entitlements secret. Then apply it to the cluster:

```bash
kubectl apply -f entitlements.yaml
```

## 4. Configure access for writing to git from the UI

Follow step _4. Configure access for writing to git from the UI_ from the installation guide here:

https://docs.gitops.weave.works/docs/installation/weave-gitops-enterprise/#4-configure-access-for-writing-to-git-from-the-ui

## 5. Configure password

In order to login to the WGE UI, you need to generate a bcrypt hash for your chosen password and store it as a secret in the Kubernetes cluster.

There are several different ways to generate a bcrypt hash, this guide uses `gitops get bcrypt-hash` from our CLI, which can be installed by following
the instructions [here](#gitops-cli).

```bash
PASSWORD="<your password>"
echo -n $PASSWORD | gitops get bcrypt-hash
$2a$10$OS5NJmPNEb13UgTOSKnMxOWlmS7mlxX77hv4yAiISvZ71Dc7IuN3q
```

Use the hashed output to create a Kubernetes username/password secret.

```bash
kubectl create secret generic cluster-user-auth \
  --namespace flux-system \
  --from-literal=username=wego-admin \
  --from-literal=password='$2a$.......'
```

## 6. Install the Weave Gitops Enterprise

Navigate to the Marketplace in the azure portal and add the Weave Gitops Enterprise Offering, during configuration select the cluster we've performed the configuration on.

## 7. Extra configuration

Additional configuration is done through an optional ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-service-extra-config
  namespace: flux-system
data:
  # disable TLS
  NO_TLS: "true"
```

Apply the configuration with:

```bash
kubectl apply -f cluster-service-extra-config.yaml

# restart the clusters-service for changes to take effect
kubectl -n flux-system rollout restart deploy/weave-gitops-enterprise-mccp-cluster-service
```

### Available configuration options

| value                 | default                            | description                                                                                                                                                                                        |
| --------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `NO_TLS`              | `"false"`                          | disable TLS                                                                                                                                                                                        |
| `CLUSTER_NAME`        | `"management"`                     | name of the management cluster                                                                                                                                                                     |
| `AUTH_METHODS`        | `"token-passthrough,user-account"` | Which auth methods to use, valid values are 'oidc', 'token-pass-through' and 'user-account'                                                                                                        |
| `OIDC_ISSUER_URL`     | ""                                 | The URL of the OpenID Connect issuer                                                                                                                                                               |
| `OIDC_CLIENT_ID`      | ""                                 | The client ID for the OpenID Connect client                                                                                                                                                        |
| `OIDC_CLIENT_SECRET`  | ""                                 | The client secret to use with OpenID Connect issuer                                                                                                                                                |
| `OIDC_REDIRECT_URL`   | ""                                 | The OAuth2 redirect URL                                                                                                                                                                            |
| `OIDC_TOKEN_DURATION` | `"1h"`                             | The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m                                                                                                 |
| `OIDC_CLAIM_USERNAME` | `"email"`                          | JWT claim to use as the user name. By default email, which is expected to be a unique identifier of the end user. Admins can choose other claims, such as sub or name, depending on their provider |
| `OIDC_CLAIM_GROUPS`   | `"groups"`                         | JWT claim to use as the user's group. If the claim is present it must be an array of strings                                                                                                       |
| `CUSTOM_OIDC_SCOPES`  | `"groups, openid, email, profile"` | Customise the requested scopes for then OIDC authentication flow - openid will always be requested                                                                                                 |
