## Disclosure

This product requires an internet connection to deploy properly. The following packages are downloaded on deployment:

## Storage of Customer Data

Weave GitOps Enterprise does not store any customer data.

### Dependencies

- Cert-Manager is installed in your Cluster. =>v1.0.0
- Flux is installed in your Cluster. =>v0.34.0
- Valid Entitlement provided by our commercial team for WGE
- GitOps CLI

## 1. Install Flux onto the cluster if not already installed

```
flux install
```

## 2. Apply the entitlement provided by Weaveworks to your cluster

```
kubectl apply -f entitlements.yaml
```

## 3. Create a password for the admin user

We need to geneate a bcrypt hash of your chosen password. You can optionally install the `gitops` cli to help generate this password hash or use some other tool.

To install the `gitops` cli:

```
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.19.0/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

Alternatively, users can install the `gitops` cli with Homebrew:

```
brew tap weaveworks/tap
brew install weaveworks/tap/gitops
```

Now lets use the `gitops` command to generate a secret to store the hash of the admin password

```
PASSWORD="<your password>"
echo -n $PASSWORD | gitops get bcrypt-hash
$2a$10$OS5NJmPNEb13UgTOSKnMxOWlmS7mlxX77hv4yAiISvZ71Dc7IuN3q

kubectl create secret generic cluster-user-auth \
  --namespace flux-system \
  --from-literal=username=wego-admin \
  --from-literal=password='$2a$.......'
```

## 4. Run Helm Install

```
helm install --set global.capiEnabled=false --namespace flux-system mccp oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/weaveworks/weave-gitops-enterprise-production --version 0.19.0
```

## 5. Get to the GUI

Use as sign-in the user and password created in the secret.

```
kubectl port-forward --namespace flux-system svc/clusters-service 8000:8000
```

Now you have a running Weave GitOps Enterprise instance. You can access the UI at https://localhost:8000

## Usage

### Changing the admin password

Delete the auth secret created during installation:

```
kubectl delete secret cluster-user-auth --namespace flux-system
```

Then follow the steps above to create a new password using the `gitops` tool.

Logout and login again to the UI using the new password.

### Assessing the health of your installation

#### CLI via `kubectl`

You can check the health of your installation by running:

```
kubectl get deploy -n flux-system
```

Check that all deployments are ready. If not, you can check the logs of the failing deployments to see what went wrong. For example

```
kubectl logs -n flux-system deploy/mccp-weave-gitops-enterprise-development-cluster-service
```

#### EKS via AWS Console

You can also check the health of your installation via the AWS Console. Go to the EKS service and select your cluster. Then go to the `Resources` tab and select `Deployments`. You should see all the deployments in the `flux-system` namespace. Check that all deployments are ready.

If there are any issues, you can check the status of the deployment by clicking on it. This will often give you a clue as to what went wrong.
