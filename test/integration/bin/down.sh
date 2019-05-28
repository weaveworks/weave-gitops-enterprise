#!/bin/bash

docker run --env-file <(env|grep -E 'CIRCLE|SECRET|SRCDIR') -v /root:/root -v /tmp:/tmp -v /home/circleci:/home/circleci --entrypoint=$SRCDIR/test/integration/bin/circle-destroy-vms quay.io/wks/build:latest
