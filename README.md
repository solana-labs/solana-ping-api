# Solana Ping API 

## Purpose
- execute solana ping command
- provide http API service
- generate a report and send to slack
- store result of solana ping
- provide different kind of datapoints

## Installation
- download executable file 
- or build from source
    - Install golang 
    - clone from git@github.com:pieceofr/solana-ping-api.git
    - go mod tidy to download packages
    - go build 
- mkdir ~/.config/ping-api
- put config.yaml in ~/.config/ping-api/config.yaml

## sugguest setup 
- mkdir ~/ping-api-server
- cp scripts in script to ~/ping-api-server
- make solana-ping-api system service 
    - create a /etc/systemd/system/solana-ping-api.service
    - remember to di ```sudo systemctl daemon-reload```

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
