#!/bin/bash

set -e

echo "$EXECUTOR_PRIVATE_KEY" > executor.pem
chmod 0600 executor.pem

set -x

executor_address=$(cat executors/metadata)

ssh-keyscan $executor_address >> $HOME/.ssh/known_hosts
remote_path=$(cat deploy/remote_path)

function cleanup { ssh -i executor.pem vcap@$executor_address rm -rf "$remote_path"; }
trap cleanup EXIT

ssh -i executor.pem vcap@$executor_address <<EOF
  cd "$remote_path"
  vagrant destroy -f
EOF
