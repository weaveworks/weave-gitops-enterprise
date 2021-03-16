#!/bin/bash

set -e

if [ -z $1 ] || ([ $1 != 'label' ] && [ $1 != 'reset' ])
then 
    echo "Invalid option, valid values => [ label, reset ]"
    exit 1
fi

function label_worker_node {
    WORKER_NODE=$(kubectl get nodes|tr -s ' ' |cut -d ' ' -f1,3 |grep '<none>' |cut -d ' ' -f1 |head -1)
    echo $WORKER_NODE
    kubectl label nodes $WORKER_NODE wkp-database-volume-node=true
}

function reset_mccp {
    EVENT_WRITER_POD=$(kubectl get pods -n wkp-gitops-repo-broker|grep wkp-event-writer|tr -s ' '|cut -f1 -d ' ')
    GITOPS_BROKER_POD=$(kubectl get pods -n wkp-gitops-repo-broker|grep gitops-repo-broker|tr -s ' '|cut -f1 -d ' ')
    echo $EVENT_WRITER_POD
    echo $GITOPS_BROKER_POD
    kubectl exec -n wkp-gitops-repo-broker $EVENT_WRITER_POD -- rm /var/database/mccp.db
    kubectl delete -n wkp-gitops-repo-brokerpod $EVENT_WRITER_POD
    kubectl delete -n wkp-gitops-repo-broker pod $GITOPS_BROKER_POD
}

echo "Selected Option: "$1

if [ $1 = 'label' ]
then
    label_worker_node
fi


if [ $1 = 'reset' ]
then
    reset_mccp
fi
