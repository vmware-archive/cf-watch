#!/bin/bash

set -e
export GO15VENDOREXPERIMENT=1
working_dir=$(cd `dirname $0` && cd ../ && pwd)
export GOPATH=$working_dir/go

mkdir -p $working_dir/go/src/github/pivotal-cf
cp -r cf-watch go/src/github/pivotal-cf/cf-watch
pushd go/src/github/pivotal-cf/cf-watch > /dev/null
  go install ./vendor/github.com/onsi/ginkgo/ginkgo
  go install ./vendor/github.com/onsi/gomega
  ginkgo -r ./*
popd > /dev/null
