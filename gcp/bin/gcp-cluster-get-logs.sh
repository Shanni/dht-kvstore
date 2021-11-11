#!/bin/sh
#set -x
set -e

source gcp/login.sh

mkdir "gcp/logs-gcp" || true
for i in `seq $node_count`; do
  source gcp/bin/gcp-set-zone.sh $i
  touch gcp/logs-gcp/node$i.log
  gcloud compute scp node$i:node$i.log gcp/logs-gcp/ --zone=$ZONE
done

wait
