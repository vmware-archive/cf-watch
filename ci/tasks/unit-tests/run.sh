#!/bin/bash

set -e
export GO15VENDOREXPERIMENT=1
export GOPATH=$PWD/go

mkdir -p go/src/github.com/pivotal-cf
cp -r cf-watch go/src/github.com/pivotal-cf/
pushd go/src/github.com/pivotal-cf/cf-watch > /dev/null
  go install ./vendor/github.com/onsi/ginkgo/ginkgo
  go install ./vendor/github.com/onsi/gomega
  ginkgo -r ./*
popd > /dev/null
