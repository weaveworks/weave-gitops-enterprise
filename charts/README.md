## How to install the MCCP Helm chart on a kind cluster

Follow these steps to deploy and test MCCP on a kind cluster:
1. Create a kind cluster that exposes a port to the host machine.

    Here's an example of such a config:
    ```yaml
    # kind-config.yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
    - role: worker
      extraPortMappings:
      - containerPort: 31490
        hostPort: 31490
    ```

    Take note of the container port as that port will be exposed by NATS once the MCCP chart has been installed.

    Use the config above to create a cluster by executing the following command:
    ```bash
    > kind create cluster --name test-kind-001 --config kind-config.yaml
    ```
    You should now have a kind cluster running locally with container port `31490` exposed to the same host port.

2.  Create a secret that contains your docker repository credentials that will be used to pull down the images. You can find instructions on how to generate this secret [here](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

    Add this secret to the target namespace. This needs to be the same namespace that the Helm chart will be installed.

    ```bash
    > kubectl apply -f docker-io-pull-secret.yaml --namespace mccp
    ```

    Take note of the secret name as you will need to supply it later when installing the chart.

3.  Determine your host IP address. If you are on wifi run this command:

    ```bash
    > ipconfig getifaddr en0
    > 192.168.0.1
    ```

    Take note of the IP address as you will need to supply it later when installing the chart. This is necessary in order to establish connectivity between agents and your MCCP instance.

4. Finally install the Helm chart to the target namespace by running the following command:

    ```bash
    cd ./charts
    helm install my-mccp ./mccp --namespace mccp \
        --set "imagePullSecrets[0].name=docker-io-pull-secret" \
        --set "nats.client.service.nodePort=31490" \
        --set "agentTemplate.natsURL=192.168.0.1:31490"
    ```

    In this example we specify the secret to use in order to pull down the images. We also specify that the port that NATS will accept connections to will be port `31490`. This needs to be the same port that we have exposed in our kind cluster above. Finally we specify the URL that agents will connect to, which also uses the same port as above.

5. You should now be able to load the MCCP UI by running the following command:

    ```bash
    > kubectl port-forward --namespace mccp deployments.apps/mccp-nginx-ingress-controller 8000:80
    ```
    The MCCP UI should now be accessible at `http://localhost:8000`.

6. (Optional) To enable the latest git activity column in the clusters table, provide:
    - a secret containing a git deploy key that has read access in the organization containing the cluster repos
    - a configmap with any file that should be loaded into the `.ssh/` config folder of the containers that need to talk to git.
      For example, if you need to provide a customized `known_hosts` file include it in the configmap.

   ```bash
   kubectl create secret generic git-deploy-key \
   --namespace <target-namespace> \
   --type=Opaque \
   --from-literal="identity=$(cat <my-git-deploy-key>)"

   kubectl create configmap gitops-broker-ssh-config \
   --namespace <target-namespace> \
   --from-file=<my-known-hosts>
   ```

    Then update the `gitopsRepoBroker` fields in `values.yaml` with the names of the created objects.
