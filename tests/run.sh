#!/bin/bash

set -e # -x

cpi_path=$PWD/cpi

rm -f creds.yml

echo "-----> `date`: Create dev release"
bosh create-release --force --dir ./../ --tarball $cpi_path

echo "-----> `date`: Create env"
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  --state ./state.json \
  --vars-store ./creds.yml \
  -o ~/workspace/bosh-deployment/virtualbox/cpi.yml \
  -o ~/workspace/bosh-deployment/virtualbox/outbound-network.yml \
  -o ~/workspace/bosh-deployment/bosh-lite.yml \
  -o ~/workspace/bosh-deployment/bosh-lite-runc.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  -v director_name="Bosh Lite Director" \
  -v virtualbox_cpi_path=$cpi_path \
  -v internal_ip=192.168.50.6 \
  -v internal_gw=192.168.50.1 \
  -v internal_cidr=192.168.50.0/24 \
  -v outbound_network_name=NatNetwork

export BOSH_ENVIRONMENT=192.168.50.6
export BOSH_CA_CERT="$(bosh int creds.yml --path /director_ssl/ca)"
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET="$(bosh int creds.yml --path /admin_password)"

echo "-----> `date`: Update cloud config"
bosh -n update-cloud-config ~/workspace/bosh-deployment/warden/cloud-config.yml

echo "-----> `date`: Upload stemcell"
bosh -n upload-stemcell "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3421.9" \
  --sha1 1396d7877204e630b9e77ae680f492d26607461d

echo "-----> `date`: Create env second time to test disk attachment"
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  --state ./state.json \
  --vars-store ./creds.yml \
  -o ~/workspace/bosh-deployment/virtualbox/cpi.yml \
  -o ~/workspace/bosh-deployment/virtualbox/outbound-network.yml \
  -o ~/workspace/bosh-deployment/bosh-lite.yml \
  -o ~/workspace/bosh-deployment/bosh-lite-runc.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  -v director_name="Bosh Lite Director" \
  -v virtualbox_cpi_path=$cpi_path \
  -v internal_ip=192.168.50.6 \
  -v internal_gw=192.168.50.1 \
  -v internal_cidr=192.168.50.0/24 \
  -v outbound_network_name=NatNetwork \
  --recreate

echo "-----> `date`: Delete previous deployment"
bosh -n -d zookeeper delete-deployment --force

echo "-----> `date`: Deploy"
bosh -n -d zookeeper deploy zookeeper.yml

echo "-----> `date`: Recreate all VMs"
bosh -n -d zookeeper recreate

echo "-----> `date`: Exercise deployment"
bosh -n -d zookeeper run-errand smoke_tests

echo "-----> `date`: Restart deployment"
bosh -n -d zookeeper restart

echo "-----> `date`: Report any problems"
bosh -n -d zookeeper cck --report

echo "-----> `date`: Delete random VM"
bosh -n -d zookeeper delete-vm `bosh -d zookeeper vms|sort|cut -f5|head -1`

echo "-----> `date`: Fix deleted VM"
bosh -n -d zookeeper cck --auto

echo "-----> `date`: Delete deployment"
bosh -n -d zookeeper delete-deployment

echo "-----> `date`: Clean up disks, etc."
bosh -n -d zookeeper clean-up --all

echo "-----> `date`: Deleting env"
bosh delete-env ~/workspace/bosh-deployment/bosh.yml \
  --state ./state.json \
  --vars-store ./creds.yml \
  -o ~/workspace/bosh-deployment/virtualbox/cpi.yml \
  -o ~/workspace/bosh-deployment/virtualbox/outbound-network.yml \
  -o ~/workspace/bosh-deployment/bosh-lite.yml \
  -o ~/workspace/bosh-deployment/bosh-lite-runc.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  -v director_name="Bosh Lite Director" \
  -v virtualbox_cpi_path=$cpi_path \
  -v internal_ip=192.168.50.6 \
  -v internal_gw=192.168.50.1 \
  -v internal_cidr=192.168.50.0/24 \
  -v outbound_network_name=NatNetwork

echo "-----> `date`: Done"
