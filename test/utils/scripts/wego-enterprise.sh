#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'setup' ] && [ ${args[0]} != 'reset' ] && [ ${args[0]} != 'reset_mccp' ])
then 
    echo "Invalid option, valid values => [ setup, reset, reset_mccp ]"
    exit 1
fi

set -x 

function setup {
  if [ ${#args[@]} -ne 2 ]
  then
    echo "Workspace path is a required argument"
    exit 1
  fi

  UI_NODEPORT=30080
  NATS_NODEPORT=31490

  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    WORKER_NAME=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1 | cut -d '/' -f2-)
    WORKER_NODE_EXTERNAL_IP=$(kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='${WORKER_NAME}')].status.addresses[?(@.type=='ExternalIP')].address}")

    # Configure inbound NATS and UI node ports
    if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ]; then
      INSTANCE_SECURITY_GROUP=$(aws ec2 describe-instances --filter "Name=ip-address,Values=${WORKER_NODE_EXTERNAL_IP}" --query 'Reservations[*].Instances[*].NetworkInterfaces[0].Groups[0].{sg:GroupId}' --output text)
      aws ec2 authorize-security-group-ingress --group-id ${INSTANCE_SECURITY_GROUP}  --ip-permissions FromPort=${NATS_NODEPORT},ToPort=${NATS_NODEPORT},IpProtocol=tcp,IpRanges='[{CidrIp=0.0.0.0/0}]',Ipv6Ranges='[{CidrIpv6=::/0}]'
      aws ec2 authorize-security-group-ingress --group-id ${INSTANCE_SECURITY_GROUP}  --ip-permissions FromPort=${UI_NODEPORT},ToPort=${UI_NODEPORT},IpProtocol=tcp,IpRanges='[{CidrIp=0.0.0.0/0}]',Ipv6Ranges='[{CidrIpv6=::/0}]'
    else
      gcloud compute firewall-rules create nats-node-port --allow tcp:${NATS_NODEPORT}
      gcloud compute firewall-rules create ui-node-port --allow tcp:${UI_NODEPORT}
    fi
  elif [ -z ${WORKER_NODE_EXTERNAL_IP} ]; then
  # MANAGEMENT_CLUSTER_KIND is a KIND cluster
    if [ "$(uname -s)" == "Linux" ]; then
      WORKER_NODE_EXTERNAL_IP=$(ifconfig eth0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:)
    elif [ "$(uname -s)" == "Darwin" ]; then
      WORKER_NODE_EXTERNAL_IP=$(ifconfig en0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:)
    fi
  fi

  # Sets the UI and CAPI endpoint URL environment variables for acceptance tests
  echo "TEST_UI_URL=http://${WORKER_NODE_EXTERNAL_IP}:${UI_NODEPORT}" >> $GITHUB_ENV
  echo "TEST_CAPI_ENDPOINT_URL=http://${WORKER_NODE_EXTERNAL_IP}:${UI_NODEPORT}" >> $GITHUB_ENV

  kubectl create namespace prom
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  helm install my-prom prometheus-community/kube-prometheus-stack \
    --namespace prom \
    --version 14.4.0 \
    --values test/utils/data/mccp-prometheus-values.yaml
  
  if [ "$GITHUB_EVENT_NAME" == "schedule" ]; then
    helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/nightly/charts-v3/
  else
    helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3/
  fi
  helm repo update  
    
  kubectl create ns wego-system
  if [ ${GIT_PROVIDER} == "github" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY.git

    kubectl create secret generic git-provider-credentials --namespace=wego-system \
    --from-literal="GIT_PROVIDER_TOKEN=${GITHUB_TOKEN}"
  elif [ ${GIT_PROVIDER} == "gitlab" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY.git

    if [ -z ${GITOPS_GIT_HOST_TYPES} ]; then
      kubectl create secret generic git-provider-credentials --namespace=wego-system \
      --from-literal="GIT_PROVIDER_TOKEN=$GITLAB_TOKEN" \
      --from-literal="GITLAB_CLIENT_ID=$GITLAB_CLIENT_ID" \
      --from-literal="GITLAB_CLIENT_SECRET=$GITLAB_CLIENT_SECRET"
    else
      kubectl create secret generic git-provider-credentials --namespace=wego-system \
      --from-literal="GIT_PROVIDER_TOKEN=$GITLAB_TOKEN" \
      --from-literal="GITLAB_CLIENT_ID=$GITLAB_CLIENT_ID" \
      --from-literal="GITLAB_CLIENT_SECRET=$GITLAB_CLIENT_SECRET" \
      --from-literal="GITLAB_HOSTNAME=$GIT_PROVIDER_HOSTNAME" \
      --from-literal="GIT_HOST_TYPES=$GITOPS_GIT_HOST_TYPES" 
    fi
  fi

  kubectl apply -f ${args[1]}/test/utils/scripts/entitlement-secret.yaml
  kubectl apply -f ${args[1]}/test/utils/data/gitlab-on-prem-ssh-config.yaml
  # CHART_VERSION=$(git describe --always --abbrev=7 | sed 's/^[^0-9]*//')
    CHART_VERSION=0.0.17-36-ga103c48

  # Install weave gitops enterprise controllers
  helmArgs=()
  helmArgs+=( --set "nats.client.service.nodePort=${NATS_NODEPORT}" )
  helmArgs+=( --set "agentTemplate.natsURL=${WORKER_NODE_EXTERNAL_IP}:${NATS_NODEPORT}" )
  helmArgs+=( --set "nginx-ingress-controller.service.nodePorts.http=${UI_NODEPORT}" )
  helmArgs+=( --set "nginx-ingress-controller.service.type=NodePort" )
  helmArgs+=( --set "config.git.type=${GIT_PROVIDER}" )
  helmArgs+=( --set "config.git.hostname=${GIT_PROVIDER_HOSTNAME}" )
  helmArgs+=( --set "config.capi.repositoryURL=${GIT_REPOSITORY_URL}" )
  helmArgs+=( --set "config.capi.repositoryPath=./management" )
  helmArgs+=( --set "config.cluster.name=$(kubectl config current-context)" )
  helmArgs+=( --set "config.capi.baseBranch=main" )

  if [ ${ACCEPTANCE_TESTS_DATABASE_TYPE} == "postgres" ]; then
    # Create postgres DB
    kubectl apply -f test/utils/data/postgres-manifests.yaml
    kubectl wait --for=condition=available --timeout=600s deployment/postgres
    POSTGRES_CLUSTER_IP=$(kubectl get service postgres -ojsonpath={.spec.clusterIP})
    kubectl create secret generic mccp-db-credentials --namespace wego-system --from-literal=username=postgres --from-literal=password=password
    
    helmArgs+=( --set "dbConfig.databaseType=postgres" )
    helmArgs+=( --set "postgresConfig.databaseName=postgres" )
    helmArgs+=( --set "dbConfig.databaseURI=${POSTGRES_CLUSTER_IP}" )
  fi

  if [ ! -z $GITOPS_GIT_HOST_TYPES ]; then
    helmArgs+=( --set "config.extraVolumes[0].name=ssh-config" )
    helmArgs+=( --set "config.extraVolumes[0].configMap.name=ssh-config" )
    helmArgs+=( --set "config.extraVolumeMounts[0].name=ssh-config" )
    helmArgs+=( --set "config.extraVolumeMounts[0].mountPath=/root/.ssh" )
  fi

  helm install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace wego-system ${helmArgs[@]}

  # Install capi infrastructure provider
  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    echo "Capi infrastructure provider support is not implemented"
  else
    # enable cluster resource sets
    export EXP_CLUSTER_RESOURCE_SET=true
    clusterctl init --infrastructure docker    
  fi

  # Install resources for bootstrapping and CNI
  if [ ${EXP_CLUSTER_RESOURCE_SET} = true ]; then
    kubectl wait --for=condition=Ready --timeout=300s -n capi-system --all pod
    kubectl apply -f ${args[1]}/test/utils/data/calico-crs.yaml
    kubectl apply -f ${args[1]}/test/utils/data/calico-crs-configmap.yaml
  fi

  if [ ${GIT_PROVIDER} == "github" ]; then
    kubectl create secret generic my-pat --from-literal GITHUB_TOKEN=$GITHUB_TOKEN
  	cat ${args[1]}/test/utils/data/capi-gitops-cluster-bootstrap-config.yaml | sed s,{{GITOPS_REPO}},$GITOPS_REPO,g | sed s,{{GIT_PROVIDER_TOKEN}},GITHUB_TOKEN,g | kubectl apply -f -
  elif [ ${GIT_PROVIDER} == "gitlab" ]; then
    kubectl create secret generic my-pat --from-literal GITLAB_TOKEN=$GITLAB_TOKEN
  	cat ${args[1]}/test/utils/data/capi-gitops-cluster-bootstrap-config.yaml | sed s,{{GITOPS_REPO}},$GITOPS_REPO,g | sed s,{{GIT_PROVIDER_TOKEN}},GITLAB_TOKEN,g | kubectl apply -f -
  fi

  # Wait for cluster to settle
  kubectl wait --for=condition=Ready --timeout=300s -n wego-system --all pod --selector='app!=wego-app'
  kubectl get pods -A

  exit 0
}

function reset {
  # Delete postgres db 
  kubectl delete deployment postgres
  kubectl delete service postgres
  # Delete namespaces and their respective resources
  kubectl delete namespaces prom wkp-agent
  # Delete wego system from the management cluster
  $GITOPS_BIN_PATH flux uninstall --silent
  $GITOPS_BIN_PATH flux uninstall --namespace wego-system --silent
  # Delete any orphan resources
  kubectl delete CAPITemplate --all
  kubectl delete ClusterBootstrapConfig --all
  kubectl delete secret my-pat
  kubectl delete ClusterResourceSet --all
  kubectl delete configmap calico-crs-configmap
}

function reset_mccp {
    EVENT_WRITER_POD=$(kubectl get pods -n wego-system|grep event-writer|tr -s ' '|cut -f1 -d ' ')
    GITOPS_BROKER_POD=$(kubectl get pods -n wego-system|grep gitops-repo-broker|tr -s ' '|cut -f1 -d ' ')    

    # Sometime due to the test conditions the cluster service pod is in transition state i.e. one terminating and the new one is being created at the same time.
    # Under such state we have two cluster srvice pods momentarily 
    counter=10
    while [ $counter -gt 0 ]
    do
        CLUSTER_SERVICE_POD=$(kubectl get pods -n wego-system|grep cluster-service|tr -s ' '|cut -f1 -d ' ')
        pod_count=$(echo $CLUSTER_SERVICE_POD | wc -w |awk '{print $1}')
        if [ $pod_count -gt 1 ]
        then            
            sleep 2
            counter=$(( $counter - 1 ))
        else
            break
        fi        
    done    

    echo $EVENT_WRITER_POD
    echo $GITOPS_BROKER_POD
    echo $CLUSTER_SERVICE_POD
    kubectl exec -n wego-system $EVENT_WRITER_POD -- rm /var/database/mccp.db
    kubectl delete -n wego-system pod $EVENT_WRITER_POD
    kubectl delete -n wego-system pod $GITOPS_BROKER_POD
    kubectl delete -n wego-system pod $CLUSTER_SERVICE_POD
}

if [ ${args[0]} = 'setup' ]; then
    setup
elif [ ${args[0]} = 'reset' ]; then
    reset
elif [ ${args[0]} = 'reset_mccp' ]; then
    reset_mccp
fi

