#!/bin/sh
#set -x
set -e

if [ "$#" -ne 1 ]; then
  echo 2>&1 "Usage: $0 NODE_COUNT"
  exit 1
fi

NODE_COUNT=$1

for i in `seq $NODE_COUNT`; do
  source gcp/bin/gcp-set-zone.sh $i
  gcloud compute instances start node$i --zone=$ZONE
done

#for i in `seq $NODE_COUNT`; do
#  source gcp/bin/gcp-set-zone.sh $i
#  gcloud compute scp dht-server node$i:~ --zone=$ZONE
#done

#gcloud compute instances start client --zone="us-west2-b"

# generate IPs for clients testing
source gcp/bin/generate-ips.sh $NODE_COUNT

