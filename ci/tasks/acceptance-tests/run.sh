#!/bin/bash

set -ex

nanocf local.nanocf

export GO15VENDOREXPERIMENT=1
export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH
export CF_DOMAIN=local.nanocf
export CF_USERNAME=admin
export CF_PASSWORD=admin

mkdir -p go/src/github.com/pivotal-cf
cp -r cf-watch go/src/github.com/pivotal-cf/

pushd go/src/github.com/pivotal-cf/cf-watch > /dev/null
  go install ./vendor/github.com/onsi/ginkgo/ginkgo
  go install ./vendor/github.com/onsi/gomega
  ginkgo .
popd > /dev/null
