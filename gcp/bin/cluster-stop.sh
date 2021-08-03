#!/bin/sh
#set -x
set -e

NODE_COUNT=$1

for i in `seq $NODE_COUNT`; do
  echo "Stopping server on node$i"

  source gcp/bin/gcp-set-zone.sh $i
  gcloud compute ssh node$i --command="sudo pkill dht-server " --zone=$ZONE || true
done
