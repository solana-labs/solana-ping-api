#!/usr/bin/env bash
set -ex
logdir="$HOME/ping-api-server/log"

if ! [[ -d $logdir ]];then
    mkdir $logdir
fi
log="$logdir/solana-ping-api.log"
## Your executable location
exec /home/sol/ping-api-server/solana-ping-api-service 2 >> $log
ret=$?

echo $ret