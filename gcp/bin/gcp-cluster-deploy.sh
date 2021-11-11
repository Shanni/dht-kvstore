#!/bin/sh
#set -x
set -e

source gcp/bin/gcp-login.sh

echo "Cross-compiling go binary"
GOOS=linux GOARCH=amd64 go build -o dht-server src/server/agamemnon-server.go


for i in `seq $node_count`; do
  echo "Deploying to node$i"

  source gcp/bin/gcp-set-zone.sh $i
  gcloud compute ssh node$i --command="sudo pkill dht-server " --zone=$zone || true
  gcloud compute scp ./gcp/gcp-server.internal.txt node$i:gcp-server.internal.txt --zone=$zone
  gcloud compute scp dht-server node$i:~ --zone=$zone
done



