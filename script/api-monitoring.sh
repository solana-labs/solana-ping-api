#!/usr/bin/env bash

declare -A cluster_query_cmd
cluster_query_cmd[mainnet]="https://ping.solana.com/mainnet-beta/last6hour"
cluster_query_cmd[testnet]="https://ping.solana.com/testnet/last6hours"
cluster_query_cmd[devnet]="https://ping.solana.com/devnet/last6hours"
slack_webhook="https://hooks.slack.com/services/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"
restart_time=$PING_SERVER_RESTART_TIME
restart_cmd=""

if [[ -z "$restart_time"  ]];then
	export PING_SERVER_RESTART_TIME=0 # initialize restart time
fi

slack_alert(){
	sdata=$(jq --null-input --arg val "$slacktext" '{"text":$val}')
	curl -X POST -H 'Content-type: application/json' --data "$sdata" $slack_webhook
}
# return is_alive, 1: alive 0:not alive (404/408 etc)
alive_check() {
    for retry in 0 1 2
	do
        echo retry=$retry
		if [[ $retry -gt 0 ]];then
			sleep 5
		fi
		alive_status_code=$(curl -o /dev/null -s -w "%{http_code}\n" --connect-timeout 10 ${cluster_query_cmd[$cluster]})
		if [[ $alive_status_code == 204 || $alive_status_code == 200 ]];then
			is_alive=1
            echo $cluster  $alive_name is alive, status:$alive_status_code
			export PING_SERVER_RESTART_TIME=0 #reset restart time
            break
		else
			is_alive=0
            echo $cluster $alive_name is NOT alive, status:$alive_status_code
		fi
   
	done
}

for cluster in "mainnet" "testnet" "devnet"
do
	alive_check
	if [[ $is_alive -eq 0 ]];
		# restart solana-ping-api.service
		#exec /etc/init.d/solana-ping-api.service restart
		if [[ $restart_time -lt 3 ]];then
			restart_time=$restart_time+1
			export PING_SERVER_RESTART_TIME=$restart_time
			echo PING_SERVER_RESTART_TIME=$PING_SERVER_RESTART_TIME
			exec /home/sol/ping-api-server/restart-api.sh
		fi
	fi
done