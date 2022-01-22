package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/types"
)

var (
	txTimeoutDefault               = 10 * time.Second
	waitConfirmationTimeoutDefault = 50 * time.Second
	statusCheckTimeDefault         = 1 * time.Second
)

func Transfer(c *client.Client, sender types.Account, feePayer types.Account, receiverPubkey string, txTimeout time.Duration) (txHash string, err error) {
	// to fetch recent blockhash
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		log.Fatalf("get recent block hash error, err: %v\n", err)
	}
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
	if txTimeout <= 0 {
		txTimeout = time.Duration(txTimeoutDefault)
	}
	ctx, _ := context.WithTimeout(context.TODO(), txTimeout)
	txHash, err = c.SendTransaction(ctx, tx)

	if err != nil {
		log.Printf("Error: Failed to send tx, err: %v", err)
		return "", err
	}
	log.Println("tx:", txHash, " is sent")
	return txHash, nil
}

func waitConfirmation(c *client.Client, txHash string, timeout time.Duration, queryTime time.Duration) error {
	if timeout <= 0 {
		timeout = waitConfirmationTimeoutDefault
	}
	ctx, _ := context.WithTimeout(context.TODO(), timeout)
	for {
		resp, err := c.GetTransaction(ctx, txHash)
		if err != nil {
			log.Println("waitConfirmation Error:", err)
			return err
		}
		if resp != nil {
			log.Println("tx:", txHash, "is confirmed.", " Slot:", resp.Slot, " RecentBlockHash:", resp.Transaction.Message.RecentBlockHash)
			return nil
		}
		if queryTime <= 0 {
			queryTime = statusCheckTimeDefault
		}
		time.Sleep(queryTime)
	}
}

func getConfigKeyPair(cluster Cluster) (types.Account, error) {
	var c SolanaConfig
	switch cluster {
	case MainnetBeta:
		c = config.SolanaConfigInfo.ConfigMain
	case Testnet:
		c = config.SolanaConfigInfo.ConfigTestnet
	case Devnet:
		c = config.SolanaConfigInfo.ConfigDevnet
	default:
		log.Println("StatusNotFound Error:", cluster)
		return types.Account{}, errors.New("Invalid Cluster")
	}
	body, err := ioutil.ReadFile(c.KeypairPath)
	if err != nil {

	}
	key := []byte{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return types.Account{}, err
	}

	acct, err := types.AccountFromBytes(key)
	if err != nil {
		return types.Account{}, err
	}
	return acct, nil

}
