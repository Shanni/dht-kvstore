#!/bin/sh
#set -x
set -e

node_count=$1

for i in `seq 8`; do
  echo "Starting node$i"

  source gcp/bin/gcp-set-zone.sh $i
  gcloud compute ssh node$i --command="./dht-server 333$i gcp-server.internal.txt > node$i.log 2>&1 & " --zone=$zone
done
