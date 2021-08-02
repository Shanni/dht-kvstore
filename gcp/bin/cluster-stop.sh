#!/bin/sh
#set -x
set -e

node_count=$1

for i in `seq 8`; do
  echo "Stopping node$i"

  source gcp/bin/gcp-set-zone.sh $i
  gcloud compute ssh node$i --command="sudo pkill dht-server " --zone=$zone || true
done
