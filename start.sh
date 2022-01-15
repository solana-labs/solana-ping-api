#!/usr/bin/env bash

log="$HOME/wks_go/log/ping-api.dev.log"

if ![[ -d $log ]];then
    mkdir $log
fi

set -ex
## Your executable location
exec $HOME/wks_go/solana-ping-api/solana-ping-api-service 2 >> $log
ret=$?

echo r$ret > $HOME/wks_go/solana-ping-api/execute.out