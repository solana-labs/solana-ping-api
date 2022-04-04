# Solana Ping Service 

## Functions Provided
- High frequently send transactions and record results
- provide http API service 
- generate a report and submit to slack periodically
- actively check confirmation losses and send an warning to slack

## Server Setup
### PingService
This is similar to  "solana ping" tool in solana tool but can do concurrent rpc query.
It send transactions to rpc endpoint and wait for transactions is confirmed. 
Use `PingService: true` to turn on. The default is On. 

### RetensionService
Use `RetensionService: true` to turn on. Default is Off.
Clean database data periodically.

### SlackReportService
Use `SlackReportService: true` to turn on. Default is On.
send summary of ping result to a slack channel periodically.

### SlackAlertService
Use `SlackAlertService: true` to turn on. Default is On.
If confirmation loss is greater than a thredhold, send an alert to a channel

### Example:Run only API Query Server
In config.yaml ServerSetup: 
```
 PingService: true           
 RetensionService: false        
 SlackReportService: true
 SlackAlertService: true    
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
