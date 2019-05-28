#!/bin/bash


docker login -u="$DOCKER_USER" -p="$DOCKER_PASSWORD" quay.io
export GOROOT=/home/circleci/go-${GOVERSION}
export PATH=~/go-${GOVERSION}/bin:$PATH
# Work around for broken docker package in RHEL
# See https://github.com/weaveworks/wks/issues/235
export DOCKER_VERSION='1.13.1-75*'
go test -failfast -v -timeout 1h ./test/integration -args -run.interactive -cmd /tmp/workspace/cmd/wksctl/wksctl -tags.wks-k8s-krb5-server=$(./tools/image-tag) -tags.wks-mock-authz-server=$(./tools/image-tag)
