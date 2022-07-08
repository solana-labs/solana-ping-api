# Solana Ping Service 

## Functions Provided
- High frequently send transactions and record results
- provide http API service 
- generate a report and submit to slack periodically
- actively check confirmation losses and send an alert to slack
- Spam Filter of slack alert

## Server Setup
### API Service
API service for getting the results of ping service. 
Use `APIServer: Enabled: true` to turn on in in config-{cluster}.yaml. The default is true for mainnet but false for other clusters.

### PingService
This is similar to  "solana ping" tool in solana tool but can do concurrent rpc query.
It send transactions to rpc endpoint and wait for transactions is confirmed. 
Use `PingServiceEnabled: true` to turn on in config-{cluster}.yaml. The default is On. 

### RetensionService
Use `Retension: Enabled: true` in config.yaml to turn on. Default is Off.
Clean database data periodically.

### SlackReportService
Use `SlackReport: Enabled:true` in config-{cluster}.yaml to turn on. Default is On.
send summary of ping result to a slack channel periodically.

### SlackAlertService
Use `SlackReport: SlackAlert: Enabled: true`  in config-{cluster}.yaml to turn on. Default is On.
If confirmation loss is greater than a thredhold, send an alert to a channel

+ Example:Run only API Query Server
In config.yaml ServerSetup: 

```
(config.yaml)
Retension:
 Enabled: false
(config-{cluster}.yaml)
PingEnabled: true     
SlackReport:
 Enabled: true
 SlackAlert: 
  Enabled: true                        
```
## Installation
- download executable file 
- or build from source
    - Install golang 
    - clone from github.com/solana-labs/solana-ping-api
    - go mod tidy to download packages
    - go build 
- mkdir ~/.config/ping-api
- put config.yaml in ~/.config/ping-api/config.yaml

### Using GCP Database
- Install & Setup google cloud CLI
- download [Cloud SQL Auth proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy)
- chmod +x cloud_sql_proxy
- run cloud_sql_proxy

## setup recommendation
- mkdir ~/ping-api-server
- cp scripts in script to ~/ping-api-server
- make solana-ping-api system service 
    - create a /etc/systemd/system/solana-ping-api.service
    - remember to reload by ```sudo systemctl daemon-reload```

```
[Unit]
Description=Solana Ping API Service
After=network.target
StartLimitIntervalSec=1

[Service]
Type=simple
Restart=always
RestartSec=30
User=sol
LogRateLimitIntervalSec=0
ExecStart=/home/sol/ping-api-server/solana-ping-restart.sh

[Install]
WantedBy=multi-user.target

```

- put executable file in ~/ping-api-server
- cp config.yaml.samle to ~/ping-api-server/config.yaml and modify it 
- use cp-to-real-config.sh to copy config.yaml to ~/.config/ping-api/config.yaml
- start service by sudo sysmtemctl start solana-ping-api.service
- you can check log by ```sudo tail -f /var/log/syslog | grep ping-api```

## Alert Spam Filter

Alert Spam Filter could be changed frequently. The updte to date (4/18/2022) setting  is as below.
```
    Threshold increases when
    Loss > 20 % -> new threshold = 50% -> send alert
    Loss > 50 % -> new threshold = 75% -> send alert
    Loss > 75 % -> new threshold = 100% -> send alert
    Threshold decreases when
    Loss > 75 % to < 75%  -> new threshold = 75% -> send alert
    Loss > 50 % to < 50%  -> new threshold = 50% -> send alert
    Loss > 20 % to < 20%  -> new threshold = 20% -> NOT send alert
```