package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/cmptbdgprog"
	"github.com/portto/solana-go-sdk/program/memoprog"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/rpc"
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
		//log.Println("get recent block hash error, err:", err)
		return "", err
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
		// log.Printf("Error: Failed to send tx, err: %v", err)
		return "", err
	}
	return txHash, nil
}

type SendPingTxParam struct {
	Client              *client.Client
	Ctx                 context.Context
	FeePayer            types.Account
	RequestComputeUnits uint32
	ComputeUnitPrice    uint64 // micro lamports
}

func SendPingTx(param SendPingTxParam) (string, error) {
	latestBlockhashResponse, err := param.Client.GetLatestBlockhash(param.Ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get latest blockhash, err: %v", err)
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{param.FeePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        param.FeePayer.PublicKey,
			RecentBlockhash: latestBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				cmptbdgprog.SetComputeUnitLimit(cmptbdgprog.SetComputeUnitLimitParam{
					Units: param.RequestComputeUnits,
				}),
				cmptbdgprog.SetComputeUnitPrice(cmptbdgprog.SetComputeUnitPriceParam{
					MicroLamports: param.ComputeUnitPrice,
				}),
				memoprog.BuildMemo(memoprog.BuildMemoParam{
					Memo: []byte("ping"),
				}),
			},
		}),
	})
	if err != nil {
		return "", fmt.Errorf("failed to new a tx, err: %v", err)
	}

	txhash, err := param.Client.SendTransaction(param.Ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to send a tx, err: %v", err)
	}

	return txhash, nil
}

func waitConfirmation(c *client.Client, txHash string, timeout time.Duration, requestTimeout time.Duration, queryTime time.Duration) error {
	if timeout <= 0 {
		timeout = waitConfirmationTimeoutDefault
		log.Println("timeout is not set! Use default timeout", timeout, " sec")
	}

	ctx, _ := context.WithTimeout(context.TODO(), requestTimeout)
	elapse := time.Now()
	for {
		resp, err := c.GetSignatureStatus(ctx, txHash)
		now := time.Now()
		if err != nil {
			if now.Sub(elapse).Seconds() < timeout.Seconds() {
				continue
			} else {
				return err
			}
		}
		if resp != nil {
			if *resp.ConfirmationStatus == rpc.CommitmentConfirmed || *resp.ConfirmationStatus == rpc.CommitmentFinalized {
				return nil
			}
		}
		if now.Sub(elapse).Seconds() > timeout.Seconds() {
			return err
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
