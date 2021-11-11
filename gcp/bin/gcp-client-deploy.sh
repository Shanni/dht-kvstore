#!/bin/sh
#set -x
set -e

ZONE="us-west2-b"

#gcloud compute instances create client \
#          --zone=$ZONE \
#          --machine-type="e2-micro"

FILE=dht-client
if [ -f $FILE ]; then
  echo "$FILE exist, now remove old version"
  rm -f dht-cleint
fi

GOOS=linux GOARCH=amd64 go build -o dht-client test-client/gcp-client.go
gcloud compute scp dht-client client:~ --zone=$ZONE

gcloud compute scp gcp/gcp-server.internal.txt client:servers.txt --zone=$ZONE
gcloud compute scp gcp/tester/CPEN533_MP_Tests-1.0-SNAPSHOT-all.jar client:CPEN533_MP_Tests-1.0-SNAPSHOT-all.jar --zone=$ZONE
gcloud compute scp gcp/tester/mpTests-1.0-SNAPSHOT-all.jar client:mpTests-1.0-SNAPSHOT-all.jar --zone=$ZONE

