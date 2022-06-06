#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'setup' ] && [ ${args[0]} != 'reset' ] && [ ${args[0]} != 'reset_controllers' ])
then 
    echo "Invalid option, valid values => [ setup, reset, reset_controllers ]"
    exit 1
fi

set -x 

function get-localhost-ip {
  local  __resultvar=$1
  local $interface
  local locahost_ip
  for i in {0..10}
  do
    if [ "$(uname -s)" == "Linux" ]; then
      interface=eth$i
    elif [ "$(uname -s)" == "Darwin" ]; then
      interface=en$i
    fi
    locahost_ip=$(ifconfig $interface | grep -i MASK | awk '{print $2}' | cut -f2 -d: | grep -E '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}')
    if [ -z $locahost_ip ]; then
      continue
    else          
      break
    fi
  done  
  eval $__resultvar="'$locahost_ip'"
}

function setup {
  if [ ${#args[@]} -ne 2 ]
  then
    echo "Workspace path is a required argument"
    exit 1
  fi

  get-localhost-ip LOCALHOST_IP

  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    WORKER_NAME=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1 | cut -d '/' -f2-)
    WORKER_NODE_EXTERNAL_IP=$(kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='${WORKER_NAME}')].status.addresses[?(@.type=='ExternalIP')].address}")

    # Configure inbound UI node ports
    if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ]; then
      INSTANCE_SECURITY_GROUP=$(aws ec2 describe-instances --filter "Name=ip-address,Values=${WORKER_NODE_EXTERNAL_IP}" --query 'Reservations[*].Instances[*].NetworkInterfaces[0].Groups[0].{sg:GroupId}' --output text)
      aws ec2 authorize-security-group-ingress --group-id ${INSTANCE_SECURITY_GROUP}  --ip-permissions FromPort=${UI_NODEPORT},ToPort=${UI_NODEPORT},IpProtocol=tcp,IpRanges='[{CidrIp=0.0.0.0/0}]',Ipv6Ranges='[{CidrIpv6=::/0}]'
    else
      gcloud compute firewall-rules create ui-node-port --allow tcp:${UI_NODEPORT}
    fi
  elif [ -z ${WORKER_NODE_EXTERNAL_IP} ]; then
    # MANAGEMENT_CLUSTER_KIND is a KIND cluster
    WORKER_NODE_EXTERNAL_IP=${LOCALHOST_IP}
  fi

  # Set enterprise cluster CNAME host entry in the hosts file
  hostEntry=$(cat /etc/hosts | grep "${WORKER_NODE_EXTERNAL_IP} ${MANAGEMENT_CLUSTER_CNAME}")
  upgradeHostEntry=$(cat /etc/hosts | grep "${LOCALHOST_IP} ${UPGRADE_MANAGEMENT_CLUSTER_CNAME}")
  if [ -z "${hostEntry}" ] || [ -z "${upgradeHostEntry}" ]; then
    echo "${WORKER_NODE_EXTERNAL_IP} ${MANAGEMENT_CLUSTER_CNAME}" | sudo tee -a /etc/hosts
    echo "${LOCALHOST_IP} ${UPGRADE_MANAGEMENT_CLUSTER_CNAME}" | sudo tee -a /etc/hosts
  fi

  if [ "$GITHUB_EVENT_NAME" == "schedule" ]; then
    helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/nightly/charts-v3/
  else
    helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3/
  fi
  helm repo update  
  
  kubectl create namespace flux-system

  # Create secrete for git provider authentication
  if [ ${GIT_PROVIDER} == "github" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY.git

    kubectl create secret generic git-provider-credentials --namespace=flux-system \
    --from-literal="GIT_PROVIDER_TOKEN=${GITHUB_TOKEN}"

    flux bootstrap github \
      --owner=${GITHUB_ORG} \
      --repository=${CLUSTER_REPOSITORY} \
      --branch=main \
      --path=./clusters/my-cluster

  elif [ ${GIT_PROVIDER} == "gitlab" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY.git

    if [ -z ${WEAVE_GITOPS_GIT_HOST_TYPES} ]; then
      kubectl create secret generic git-provider-credentials --namespace=flux-system \
      --from-literal="GIT_PROVIDER_TOKEN=$GITLAB_TOKEN" \
      --from-literal="GITLAB_CLIENT_ID=$GITLAB_CLIENT_ID" \
      --from-literal="GITLAB_CLIENT_SECRET=$GITLAB_CLIENT_SECRET"
    else
      kubectl create secret generic git-provider-credentials --namespace=flux-system \
      --from-literal="GIT_PROVIDER_TOKEN=$GITLAB_TOKEN" \
      --from-literal="GITLAB_CLIENT_ID=$GITLAB_CLIENT_ID" \
      --from-literal="GITLAB_CLIENT_SECRET=$GITLAB_CLIENT_SECRET" \
      --from-literal="GITLAB_HOSTNAME=$GIT_PROVIDER_HOSTNAME" \
      --from-literal="GIT_HOST_TYPES=$WEAVE_GITOPS_GIT_HOST_TYPES"
    fi

    flux bootstrap gitlab \
      --owner=${GITLAB_ORG} \
      --repository=${CLUSTER_REPOSITORY} \
      --branch=main \
      --hostname=${GIT_PROVIDER_HOSTNAME} \
      --path=./clusters/my-cluster
  fi  

  # Create admin cluster user secret
  kubectl create secret generic cluster-user-auth \
  --namespace flux-system \
  --from-literal=username=wego-admin \
  --from-literal=password=${CLUSTER_ADMIN_PASSWORD_HASH}
  
  #  Create client credential secret for OIDC (dex)
  kubectl create secret generic client-credentials \
  --namespace flux-system \
  --from-literal=clientID=${DEX_CLIENT_ID} \
  --from-literal=clientSecret=${DEX_CLIENT_SECRET}

  kubectl apply -f ${args[1]}/test/utils/scripts/entitlement-secret.yaml 

  # Choosing weave-gitops-enterprise chart version to install
  if [ -z ${ENTERPRISE_CHART_VERSION} ]; then
    CHART_VERSION=$(git describe --always --abbrev=7 | sed 's/^[^0-9]*//')
  else
    CHART_VERSION=${ENTERPRISE_CHART_VERSION}
  fi

  # Install weave gitops enterprise controllers
  helmArgs=()
  helmArgs+=( --set "service.ports.https=8000" )
  helmArgs+=( --set "service.targetPorts.https=8000" )
  helmArgs+=( --set "config.git.type=${GIT_PROVIDER}" )
  helmArgs+=( --set "config.git.hostname=${GIT_PROVIDER_HOSTNAME}" )
  helmArgs+=( --set "config.capi.repositoryURL=${GIT_REPOSITORY_URL}" )
  helmArgs+=( --set "config.capi.repositoryPath=./clusters/my-cluster/clusters" )
  helmArgs+=( --set "config.capi.repositoryClustersPath=./clusters" )
  helmArgs+=( --set "config.cluster.name=$(kubectl config current-context)" )
  helmArgs+=( --set "config.capi.baseBranch=main" )
   helmArgs+=( --set "tls.enabled=false" )
  helmArgs+=( --set "config.oidc.enabled=true" )
  helmArgs+=( --set "config.oidc.clientCredentialsSecret=client-credentials" )
  helmArgs+=( --set "config.oidc.issuerURL=${OIDC_ISSUER_URL}" )
  helmArgs+=( --set "config.oidc.redirectURL=https://${MANAGEMENT_CLUSTER_CNAME}:${UI_NODEPORT}/oauth2/callback" )

  if [ ! -z $WEAVE_GITOPS_GIT_HOST_TYPES ]; then
    helmArgs+=( --set "config.extraVolumes[0].name=ssh-config" )
    helmArgs+=( --set "config.extraVolumes[0].configMap.name=ssh-config" )
    helmArgs+=( --set "config.extraVolumeMounts[0].name=ssh-config" )
    helmArgs+=( --set "config.extraVolumeMounts[0].mountPath=/root/.ssh" )

    ssh-keyscan ${GIT_PROVIDER_HOSTNAME} > known_hosts
    kubectl create configmap ssh-config --namespace flux-system --from-file=./known_hosts
  fi

  helm install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace flux-system ${helmArgs[@]}

  helm repo add profiles-catalog https://raw.githubusercontent.com/weaveworks/weave-gitops-profile-examples/gh-pages
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm repo add cert-manager https://charts.jetstack.io
  helm repo update 

  # Install cert-manager for tls certificate creation
  helm upgrade --install \
    cert-manager cert-manager/cert-manager \
    --namespace cert-manager --create-namespace \
    --version v1.8.0 \
    --set installCRDs=true
  kubectl wait --for=condition=Ready --timeout=120s -n cert-manager --all pod

  # Install ingress-nginx for tls termination 
  helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-nginx --create-namespace \
    --version 4.0.18 \
    --set controller.service.type=NodePort \
    --set controller.service.nodePorts.https=${UI_NODEPORT}
  kubectl wait --for=condition=Ready --timeout=120s -n ingress-nginx --all pod
  
  cat ${args[1]}/test/utils/data/certificate-issuer.yaml | \
      sed s,{{HOST_NAME}},${MANAGEMENT_CLUSTER_CNAME},g | \
      kubectl apply -f -
  kubectl wait --for=condition=Ready --timeout=60s -n flux-system --all certificate

  cat ${args[1]}/test/utils/data/ingress.yaml | \
      sed s,{{HOST_NAME}},${MANAGEMENT_CLUSTER_CNAME},g | \
      kubectl apply -f -

  # Install RBAC for user authentication
   kubectl apply -f ${args[1]}/test/utils/data/rbac-auth.yaml

  # enable cluster resource sets
  export EXP_CLUSTER_RESOURCE_SET=true
  # Install capi infrastructure provider
  if [ "$CAPI_PROVIDER" == "capa" ]; then
    aws cloudformation describe-stacks --stack-name wge-capi-cluster-api-provider-aws-sigs-k8s-io --region us-east-1
    if [ $? -ne 0 ]; then
      clusterawsadm bootstrap iam create-cloudformation-stack --config aws_bootstrap_config.yaml --region=us-east-1
    fi
    export AWS_B64ENCODED_CREDENTIALS=$(clusterawsadm bootstrap credentials encode-as-profile --region=us-east-1)
    aws ec2 describe-key-pairs --key-name weave-gitops-pesto --region=us-east-1
    if [ $? -ne 0 ]; then
      aws ec2 create-key-pair --key-name weave-gitops-pesto --region us-east-1 --output text > ~/.ssh/weave-gitops-pesto.pem
    fi
    clusterctl init --infrastructure aws
  elif [ "$CAPI_PROVIDER" == "capg" ]; then
    export GCP_B64ENCODED_CREDENTIALS=$( echo ${GCP_SA_KEY} | base64 | tr -d '\n' )
    clusterctl init --infrastructure gcp
  else
    clusterctl init --infrastructure docker    
  fi

  # Install policy agent to enforce rego policies - (Installing policy agent after capi because capi violates some of thge policies and failed to install)
  helm upgrade --install weave-policy-agent profiles-catalog/weave-policy-agent \
    --version 0.3.x \
    --set accountId=weaveworks \
    --set clusterId=${MANAGEMENT_CLUSTER_CNAME}
  kubectl wait --for=condition=Ready --timeout=120s -n policy-system --all pod

  # Install resources for bootstrapping and CNI
  kubectl apply -f ${args[1]}/test/utils/data/profile-repo.yaml
  
  if [ ${EXP_CLUSTER_RESOURCE_SET} = true ]; then
    kubectl wait --for=condition=Ready --timeout=300s -n capi-system --all pod
    kubectl apply -f ${args[1]}/test/utils/data/calico-crs.yaml
    kubectl apply -f ${args[1]}/test/utils/data/calico-crs-configmap.yaml
  fi

  if [ ${GIT_PROVIDER} == "github" ]; then
    kubectl create secret generic my-pat --from-literal GITHUB_TOKEN=$GITHUB_TOKEN
    cat ${args[1]}/test/utils/data/gitops-cluster-bootstrap-config.yaml | \
      sed s,{{GIT_PROVIDER}},github,g | \
      sed s,{{GITOPS_REPO_NAME}},$CLUSTER_REPOSITORY,g | \
      sed s,{{GITOPS_REPO_OWNER}},$GITHUB_ORG,g | \
      sed s,{{GIT_PROVIDER_HOSTNAME}},$GIT_PROVIDER_HOSTNAME,g | \
      kubectl apply -f -
  elif [ ${GIT_PROVIDER} == "gitlab" ]; then
    kubectl create secret generic my-pat --from-literal GITLAB_TOKEN=$GITLAB_TOKEN
    cat ${args[1]}/test/utils/data/gitops-cluster-bootstrap-config.yaml | \
      sed s,{{GIT_PROVIDER}},gitlab,g | \
      sed s,{{GITOPS_REPO_NAME}},$CLUSTER_REPOSITORY,g | \
      sed s,{{GITOPS_REPO_OWNER}},$GITLAB_ORG,g | \
      sed s,{{GIT_PROVIDER_HOSTNAME}},$GIT_PROVIDER_HOSTNAME,g | \
      kubectl apply -f -
  fi

  # Wait for cluster to settle
  kubectl wait --for=condition=Ready --timeout=300s -n flux-system --all pod --selector='app!=wego-app'
  kubectl get pods -A

  exit 0
}

function reset {
  # Delete flux system from the management cluster
  flux uninstall --silent
  # Delete any orphan resources
  kubectl delete CAPITemplate --all
  kubectl delete ClusterBootstrapConfig --all
  kubectl delete secret my-pat
  kubectl delete ClusterResourceSet --all
  kubectl delete configmap calico-crs-configmap
  kubectl delete ClusterRoleBinding clusters-service-impersonator
  kubectl delete ClusterRole clusters-service-impersonator-role 
  # Delete policy agent
  kubectl delete ValidatingWebhookConfiguration policy-agent
  kubectl delete namespaces policy-system  
  # Delete capi provider
  if [ "$CAPI_PROVIDER" == "capa" ]; then
    clusterctl delete --infrastructure aws
  elif [ "$CAPI_PROVIDER" == "capg" ]; then
    clusterctl delete --infrastructure gcp
  else
    clusterctl delete --infrastructure docker    
  fi
}

function reset_controllers {
    if [ ${#args[@]} -ne 2 ]; then
      echo "Cotroller's type is a required argument, valid values => [ enterprise, core, all ]"
      exit 1
    fi

    
    controllerNames=()
    if [ ${args[1]} == "enterprise" ] || [ ${args[1]} == "all" ]; then
      # Sometime due to the test conditions the cluster service pod is in transition state i.e. one terminating and the new one is being created at the same time.
      # Under such state we have two cluster srvice pods momentarily 
      counter=10
      while [ $counter -gt 0 ]
      do
          CLUSTER_SERVICE_POD=$(kubectl get pods -n flux-system|grep cluster-service|tr -s ' '|cut -f1 -d ' ')
          pod_count=$(echo $CLUSTER_SERVICE_POD | wc -w |awk '{print $1}')
          if [ $pod_count -gt 1 ]
          then            
              sleep 2
              counter=$(( $counter - 1 ))
          else
              break
          fi        
      done
      controllerNames+=" ${CLUSTER_SERVICE_POD}"
    fi

    if [ ${args[1]} == "core" ] || [ ${args[1]} == "all" ]; then
      KUSTOMIZE_POD=$(kubectl get pods -n flux-system|grep kustomize-controller|tr -s ' '|cut -f1 -d ' ')
      controllerNames+=" ${KUSTOMIZE_POD}"
    fi

    kubectl delete -n flux-system pod $controllerNames
}

if [ ${args[0]} = 'setup' ]; then
    setup
elif [ ${args[0]} = 'reset' ]; then
    reset
elif [ ${args[0]} = 'reset_controllers' ]; then
    reset_controllers
fi

