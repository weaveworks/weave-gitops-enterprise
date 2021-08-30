#!/bin/bash

set -u

NAMESPACES="$1"
TEST_NAME="$2"
ARCHIVED_LOGS_PATH=$3

LOGS_PATH=/tmp/dumped-cluster-logs/
rm -rf "$LOGS_PATH"
mkdir "$LOGS_PATH"
mkdir -p "$ARCHIVED_LOGS_PATH"
kubectl cluster-info dump --namespace "$NAMESPACES" --output-directory "$LOGS_PATH"
cd "$LOGS_PATH"
ARCHIVE_PATH="$ARCHIVED_LOGS_PATH/$TEST_NAME.tar.gz"
echo "archiving to $ARCHIVE_PATH"
tar -czf "$ARCHIVE_PATH" .
