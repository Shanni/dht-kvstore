#!/bin/sh
#set -x
set -e

if [ "$#" -ne 1 ]; then
  echo 2>&1 "Usage: $0 NODE_COUNT "
  exit 1
fi

NODE_COUNT=$1

get_node_info() {
  gcloud compute instances list --filter="name=node"$i --format="value(name, networkInterfaces[].networkIP.notnull().list():label=INTERNAL_IP, networkInterfaces[].accessConfigs[0].natIP.notnull().list():label=EXTERNAL_IP)"
}

rm -f gcp/tester/servers.txt

for i in `seq $NODE_COUNT`; do

  node_info=`get_node_info $i`

  if [ $i -lt 10 ]; then
    i=0$i
  fi
#  internal_ip=`echo $node_info | awk '{print $2}'`
  external_ip=`echo $node_info | awk '{print $3}'`

#  echo $internal_ip:333$i >> servers.txt
  echo $external_ip:333$i >> gcp/tester/servers.txt
done