#!/bin/sh
#set -x
set -e

NODE_COUNT=$1

GOOS=linux GOARCH=amd64 go build -o dht-server src/server/agamemnon-server.go

for i in `seq $NODE_COUNT`; do
  echo "Starting to run server on node$i"

  if [[ -f gcp/bin/gcp-set-zone.sh ]]; then
    source gcp/bin/gcp-set-zone.sh $i
  else
    source bin/gcp-set-zone.sh $i
  fi

  j=$i
  if [ $i -lt 10 ]; then
    j=0$i
  fi

  gcloud compute ssh node$i --command="sudo pkill dht-server " --zone=$ZONE || true
  gcloud compute scp dht-server node$i:~ --zone=$ZONE

  gcloud compute ssh node$i --command="./dht-server 333$j gcp-servers.txt > node$i.log 2>&1 & " --zone=$ZONE
done

