package main

import (
	"context"
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/types"
)

var (
	txTimeout               = 20 * time.Second
	waitConfirmationTimeout = 10 * time.Second
	statusCheckTime         = 1 * time.Second
)

func useConfigTimeSetup() {
	if config.SolanaPing.PingSetup.TxTimeout > 0 {
		txTimeout = time.Duration(config.SolanaPing.PingSetup.TxTimeout)
	}
	if config.SolanaPing.PingSetup.WaitConfirmationTimeout > 0 {
		waitConfirmationTimeout = time.Duration(config.SolanaPing.PingSetup.WaitConfirmationTimeout)
	}
	if config.SolanaPing.PingSetup.StatusCheckTime > 0 {
		statusCheckTime = time.Duration(config.SolanaPing.PingSetup.StatusCheckTime)
	}
}

func Transfer(c *client.Client, sender types.Account, feePayer types.Account, receiverPubkey string) (txHash string, err error) {
	// to fetch recent blockhash
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}
	//78xstKJgEZCLu1QHQirKD3Jjtb9CbBDXZPiCoDxPNn1z to 9qT3WeLV5o3t3GVgCk9A3mpTRjSb9qBvnfrAsVKLhmU5
	// create a message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        feePayer.PublicKey,
		RecentBlockhash: res.Blockhash, // recent blockhash
		Instructions: []types.Instruction{
			sysprog.Transfer(sysprog.TransferParam{
				From:   sender.PublicKey,                           // from
				To:     common.PublicKeyFromString(receiverPubkey), // to
				Amount: 1,                                          //  SOL
			}),
		},
	})

	// create tx by message + signer
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, sender},
	})

	if err != nil {
		log.Printf("Error: Failed to create a new transaction, err: %v", err)
		return "", err
	}
	// send tx
	ctx, _ := context.WithTimeout(context.Background(), txTimeout)
	txHash, err = c.SendTransaction(ctx, tx)

	if err != nil {

		log.Printf("Error: Failed to send tx, err: %v", err)
		return "", err
	}
	log.Println("tx:", txHash, " is sent")
	return txHash, nil
}

func waitConfirmation(c *client.Client, txHash string) error {
	ctx, _ := context.WithTimeout(context.Background(), waitConfirmationTimeout)
	for {
		log.Println("next get !")
		resp, err := c.GetTransaction(ctx, txHash)
		log.Println("resp:", resp, " err:", err)
		if err != nil {
			log.Println("Error:", err)
			return err
		}
		if resp != nil {
			log.Println("tx:", txHash, "is confirmed.", " Slot:", resp.Slot, " RecentBlockHash:", resp.Transaction.Message.RecentBlockHash)
			return nil
		}
		time.Sleep(statusCheckTime)
	}
}
