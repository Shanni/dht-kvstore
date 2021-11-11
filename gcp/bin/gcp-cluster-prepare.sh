#!/bin/sh
#set -x
set -e

#source bin/gcp-login.sh
source gcp/login.sh

get_node_info() {
  gcloud compute instances list --filter=name:node$1 --format="value(name, networkInterfaces[].networkIP.notnull().list():label=INTERNAL_IP, networkInterfaces[].accessConfigs[0].natIP.notnull().list():label=EXTERNAL_IP)"
}

rm -f gcp-server.internal.txt
rm -f gcp-server.external.txt

for i in `seq $node_count`; do

  if [ $i -lt 31 ]; then
    echo "skip $i"
    continue
  fi

  echo "Setting up node$i"
  node_info=`get_node_info $i`

  source gcp/bin/gcp-set-zone.sh $i
  if [ -z "$node_info" ]; then
    gcloud compute instances create node$i \
          --zone=$ZONE \
          --machine-type="e2-micro" \
#          --boot-disk-size=200MB
    node_info=`get_node_info $i`
  fi

  internal_ip=`echo $node_info | awk '{print $2}'`
  external_ip=`echo $node_info | awk '{print $3}'`

  echo $internal_ip:$port >> gcp-server.internal.txt
  echo $external_ip:$port >> gcp-server.external.txt
done

echo "Updated gcp-server.internal.txt and gcp-server.external.txt"
