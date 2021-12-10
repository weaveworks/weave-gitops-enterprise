#!/bin/bash

set -u

TEST_NAME="$1"
ARCHIVED_LOGS_PATH=$2

LOGS_PATH=/tmp/dumped-cluster-logs/
rm -rf "$LOGS_PATH"
mkdir "$LOGS_PATH"
mkdir -p "$ARCHIVED_LOGS_PATH"
kubectl cluster-info dump --all-namespaces --output-directory "$LOGS_PATH"
cd "$LOGS_PATH"
ARCHIVE_PATH="$ARCHIVED_LOGS_PATH/$TEST_NAME.tar.gz"
echo "archiving to $ARCHIVE_PATH"
tar -czf "$ARCHIVE_PATH" .
