## 1. Apply the entitlement provided by Weaveworks to your cluster

```
kubectl apply -f entitlements.yaml
```

## 2. Install Flux

```
flux bootstrap github \
  --owner=<github username> \
  --repository=fleet-infra \
  --branch=main \
  --path=./clusters/management \
  --personal
```

## 3. Create a password for the admin user

We need to geneate a bcrypt hash of your chosen password. You can optionally install the `gitops` cli to help generate this password hash or use some other tool.

To install the `gitops` cli:

```
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.9.5/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
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
helm install oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/weaveworks/weave-gitops-eks-accelerator:chart-v0.9.4 --namespace flux-system
```

## 5. Get to the GUI
```
kubectl port-forward --namespace flux-system svc/clusters-service 8000:8000
```

Now you have a running Weave GitOps EKS Accelerator


