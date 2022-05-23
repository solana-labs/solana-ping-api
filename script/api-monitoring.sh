#!/usr/bin/env bash

declare -A cluster_query_cmd
cluster_query_cmd[mainnet]="https://ping.solana.com/mainnet-beta/last6hours"
cluster_query_cmd[testnet]="https://ping.solana.com/testnet/last6hours"
cluster_query_cmd[devnet]="https://ping.solana.com/devnet/last6hours"

slack_webhook=""


slack_alert(){
	sdata=$(jq --null-input --arg val "$slacktext" '{"text":$val}')
	curl -X POST -H 'Content-type: application/json' --data "$sdata" $slack_webhook
}

# return is_alive, 1: alive 0:not alive (404/408 etc)
alive_check() {
	for retry in 0 1 2
	do
	if [[ $retry -gt 0 ]];then
		echo retry $retry after 30 sec ... 
		sleep 30
	fi
	alive_status_code=$(curl -o /dev/null -s -w "%{http_code}\n" --connect-timeout 10 ${cluster_query_cmd[$cluster]})
	if [[ $alive_status_code == 204 || $alive_status_code == 200 ]];then
		is_alive=1
		restart_times=0
            	echo $(date --rfc-3339=seconds) => $cluster  $alive_name is alive, status:$alive_status_code
            	break
	else
		is_alive=0
            	echo $(date --rfc-3339=seconds) => $cluster $alive_name is NOT alive, status:$alive_status_code
	fi
	done
}

echo $(date --rfc-3339=seconds)  ping-api-alive check ----

for cluster in "mainnet"
do
	alive_check
	if [[ $is_alive -eq 0 ]];then
		slacktext="{hostname: $HOSTNAME, cluster:$cluster, msg: api last6hours return $alive_status_code. Ping API Service has broken }"
		slack_alert
	fi
done
