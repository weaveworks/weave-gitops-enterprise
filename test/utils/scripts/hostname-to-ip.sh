#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ]; then 
    echo "Cluster hostname/CNAME argument is required"
    exit 1
fi

function get-localhost-ip {
  local  __resultvar=$1
  local interface
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

function get-external-ip {
  local  __resultvar=$1
  local worker_name
  local external_ip
  
  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    worker_name=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1 | cut -d '/' -f2-)
    external_ip=$(kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='${worker_name}')].status.addresses[?(@.type=='ExternalIP')].address}")
  fi
  eval $__resultvar="'$external_ip'"
}

get-localhost-ip LOCALHOST_IP
get-external-ip WORKER_NODE_EXTERNAL_IP

if [ -z ${WORKER_NODE_EXTERNAL_IP} ]; then
    # MANAGEMENT_CLUSTER_KIND is a KIND cluster
    WORKER_NODE_EXTERNAL_IP=${LOCALHOST_IP}
fi

 # Set cluster CNAME host entry in the hosts file
  hostEntry=$(cat /etc/hosts | grep "${WORKER_NODE_EXTERNAL_IP} ${args[0]}")
  if [ -z "${hostEntry}" ]; then
    echo "Setting hostname entry to /etc/hosts file: ${WORKER_NODE_EXTERNAL_IP} ${args[0]}"
    echo "${WORKER_NODE_EXTERNAL_IP} ${args[0]}" | sudo tee -a /etc/hosts
  fi
