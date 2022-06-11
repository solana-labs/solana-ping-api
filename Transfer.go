package main

import (
	"context"
	"fmt"
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

func Transfer(c *client.Client, sender types.Account, feePayer types.Account, receiverPubkey string, txTimeout time.Duration) (txHash string, pingErr PingResultError) {
	// to fetch recent blockhash
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		return "", PingResultError(err.Error())
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
		return "", PingResultError(err.Error())
	}
	// send tx
	if txTimeout <= 0 {
		txTimeout = time.Duration(txTimeoutDefault)
	}
	ctx, _ := context.WithTimeout(context.TODO(), txTimeout)
	txHash, err = c.SendTransaction(ctx, tx)

	if err != nil {
		log.Printf("Error: Failed to send tx, err: %v", err)
		return "", PingResultError(err.Error())
	}
	return txHash, EmptyPingResultError
}

type SendPingTxParam struct {
	Client              *client.Client
	Ctx                 context.Context
	FeePayer            types.Account
	RequestComputeUnits uint32
	ComputeUnitPrice    uint32
}

func SendPingTx(param SendPingTxParam) (string, PingResultError) {
	latestBlockhashResponse, err := param.Client.GetLatestBlockhash(param.Ctx)
	if err != nil {
		return "", PingResultError(fmt.Sprintf("failed to get latest blockhash, err: %v", err))
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{param.FeePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        param.FeePayer.PublicKey,
			RecentBlockhash: latestBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				cmptbdgprog.RequestUnits(cmptbdgprog.RequestUnitsParam{
					Units:         param.RequestComputeUnits,
					AdditionalFee: (param.RequestComputeUnits * param.ComputeUnitPrice) / 1_000_000,
				}),
				memoprog.BuildMemo(memoprog.BuildMemoParam{
					Memo: []byte("ping"),
				}),
			},
		}),
	})
	if err != nil {
		return "", PingResultError(fmt.Sprintf("failed to new a tx, err: %v", err))
	}

	txhash, err := param.Client.SendTransaction(param.Ctx, tx)
	if err != nil {
		return "", PingResultError(fmt.Sprintf("failed to send a tx, err: %v", err))
	}

	return txhash, PingResultError("")
}

/*
	timeout: timeout for checking  a block with the assigned htxHash status
	requestTimeout: timeout for GetSignatureStatus
	checkInterval: interval to check for status

*/

func waitConfirmation(c *client.Client, txHash string, timeout time.Duration, requestTimeout time.Duration, checkInterval time.Duration) PingResultError {
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
				return PingResultError(err.Error())
			}
		}
		if resp != nil {
			if *resp.ConfirmationStatus == rpc.CommitmentConfirmed || *resp.ConfirmationStatus == rpc.CommitmentFinalized {
				return EmptyPingResultError
			}
		}
		if now.Sub(elapse).Seconds() > timeout.Seconds() {
			return PingResultError(ErrWaitForConfirmedTimeout.Error())
		}

		if checkInterval <= 0 {
			checkInterval = statusCheckTimeDefault
		}
		time.Sleep(checkInterval)
	}
}
