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
	ConfigMainnet = "/home/pieceofr/.config/solana/cli/config.yml"
	ConfigTestnet = "/home/pieceofr/.config/solana/cli/config.yml"
	ConfigDevnet  = "/home/pieceofr/.config/solana/cli/config.yml"
	PingExePath   = "/home/sol/.local/share/solana/install/active_release/bin"
)

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
		config = ConfigMainnet
	case Testnet:
		config = ConfigTestnet
	case Devnet:
		config = ConfigDevnet
	default:
		config = ConfigDevnet
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
