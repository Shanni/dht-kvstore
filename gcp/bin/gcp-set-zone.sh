#!/bin/sh
#set -x
set -e

ZONE="us-west1-b"
if [[ $1 -gt 7 && $1 -lt 24 ]]; then
  ZONE="us-west2-b"
elif [[ $1 -ge 24 && $1 -lt 40 ]]; then
  ZONE="us-west1-a"
fi
