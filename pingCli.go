package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

const (
	ArgCount      = "3"
	ArgInterval   = "0"
	ConfigMainnet = "config-mainnet-beta.yml"
	ConfigTestnet = "config-testnet.yml"
	ConfigDevnet  = "config-devnet.yml"
	PingExePath   = "/home/sol/.local/share/solana/install/active_release/bin"
)

var ConfigPath = ""

type Cluster string

const (
	MainnetBeta Cluster = "mainnet-beta"
	Testnet             = "testnet"
	Devnet              = "devnet"
)

func solanaPing(c Cluster) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var config string
	switch c {
	case MainnetBeta:
		config = ConfigPath + ConfigMainnet
	case Testnet:
		config = ConfigPath + ConfigTestnet
	case Devnet:
		config = ConfigPath + ConfigDevnet
	default:
		config = ConfigPath + ConfigDevnet
	}

	cmd := exec.CommandContext(ctx, "solana", "ping", "-c", ArgCount, "-i", ArgInterval, "-C", config)
	cmd.Env = append(os.Environ(), ":"+PingExePath)
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
