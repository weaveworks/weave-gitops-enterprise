#!/bin/bash


docker run --env-file <(env|grep -E 'CIRCLE|SECRET|SRCDIR') -v /root:/root -v /tmp:/tmp -v /home/circleci:/home/circleci --entrypoint=$SRCDIR/test/integration/bin/provision_test_vms.sh quay.io/wks/build:latest

