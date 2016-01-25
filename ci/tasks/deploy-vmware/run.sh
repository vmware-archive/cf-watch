#!/bin/bash

set -e

echo "$EXECUTOR_PRIVATE_KEY" > executor.pem
chmod 0600 executor.pem

set -x

executor_address=$(cat executors/metadata)

ssh-keyscan $executor_address >> $HOME/.ssh/known_hosts
remote_path=$(ssh -i executor.pem vcap@$executor_address mktemp -td deploy-vmware.XXXXXXXX)
echo $remote_path > remote_path

rsync -a -e "ssh -i executor.pem" vagrantfile/Vagrantfile-* vcap@$executor_address:$remote_path/Vagrantfile

ssh -i executor.pem vcap@$executor_address <<EOF
  cd "$remote_path"
  vagrant up --provider=vmware_workstation
EOF

echo local.micropcf.io > domain
echo 192.168.11.11 > ip
