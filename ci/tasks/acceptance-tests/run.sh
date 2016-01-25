#!/bin/bash

set -e

echo "$EXECUTOR_PRIVATE_KEY" > executor.pem
chmod 0600 executor.pem

set -x

executor_address=$(cat executors/metadata)

ssh-keyscan $executor_address >> $HOME/.ssh/known_hosts
remote_path=$(ssh -i executor.pem vcap@$executor_address mktemp -td cf-watch.XXXXXXXX)

function cleanup { ssh -i executor.pem vcap@$executor_address rm -rf "$remote_path"; }
trap cleanup EXIT

cf_watch_path=go/src/github.com/pivotal-cf/cf-watch
ssh -A -i executor.pem vcap@$executor_address mkdir -p $remote_path/$cf_watch_path

rsync -a -e "ssh -i executor.pem" cf-watch vcap@$executor_address:$remote_path/$cf_watch_path
rm -rf micropcf || true

domain=$(cat deploy/domain)

ssh -A -i executor.pem vcap@$executor_address <<EOF
  export GO15VENDOREXPERIMENT=1
  export GOPATH=$remote_path/go
  export PATH=$remote_path/go/bin:\$PATH
  export CF_DOMAIN=local.micropcf.io
  export CF_USERNAME=admin
  export CF_PASSWORD=admin

  cd $remote_path/$cf_watch_path
  go install ./vendor/github.com/onsi/ginkgo/ginkgo
  go install ./vendor/github.com/onsi/gomega
  ginkgo .
EOF
