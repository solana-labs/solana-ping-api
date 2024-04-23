package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/cmptbdgprog"
	"github.com/blocto/solana-go-sdk/program/sysprog"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
)

var (
	txTimeoutDefault               = 10 * time.Second
	waitConfirmationTimeoutDefault = 50 * time.Second
	statusCheckTimeDefault         = 1 * time.Second
)

func Transfer(c *client.Client, sender types.Account, feePayer types.Account, receiverPubkey string, txTimeout time.Duration) (txHash string, pingErr PingResultError) {
	// to fetch recent blockhash
	res, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Println("Failed to get latest blockhash, err: ", err)
		return "", PingResultError(fmt.Sprintf("Failed to get latest blockhash, err: %v", err))
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
		return "", PingResultError(fmt.Sprintf("Failed to new a tx, err: %v", err))
	}
	// send tx
	if txTimeout <= 0 {
		txTimeout = time.Duration(txTimeoutDefault)
	}
	ctx, _ := context.WithTimeout(context.TODO(), txTimeout)
	txHash, err = c.SendTransaction(ctx, tx)

	if err != nil {
		log.Printf("Error: Failed to send tx, err: %v", err)
		return "", PingResultError(fmt.Sprintf("Failed to send a tx, err: %v", err))
	}
	return txHash, EmptyPingResultError
}

type SendPingTxParam struct {
	Client              *client.Client
	Ctx                 context.Context
	FeePayer            types.Account
	RequestComputeUnits uint32
	ComputeUnitPrice    uint64 // micro lamports
	ReceiverPubkey      string
}

func SendPingTx(param SendPingTxParam) (string, PingResultError) {
	latestBlockhashResponse, err := param.Client.GetLatestBlockhashWithConfig(
		param.Ctx,
		client.GetLatestBlockhashConfig{
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return "", PingResultError(fmt.Sprintf("failed to get latest blockhash, err: %v", err))
	}

	rand.Seed(time.Now().UnixNano())
	amount := uint64(rand.Intn(1000)) + 1

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
				sysprog.Transfer(sysprog.TransferParam{
					From:   param.FeePayer.PublicKey,
					To:     common.PublicKeyFromString(param.ReceiverPubkey),
					Amount: amount,
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
				return PingResultError(fmt.Sprintf("failed to get signatureStatus, err: %v", err))
			}
		}
		if resp != nil {
			if *resp.ConfirmationStatus == rpc.CommitmentConfirmed || *resp.ConfirmationStatus == rpc.CommitmentFinalized {
				return EmptyPingResultError
			}
		}
		if now.Sub(elapse).Seconds() > timeout.Seconds() {
			if resp != nil && *resp.ConfirmationStatus == rpc.CommitmentProcessed {
				return PingResultError(ErrInProcessedStateTimeout.Error())
			}
			return PingResultError(ErrWaitForConfirmedTimeout.Error())
		}

		if checkInterval <= 0 {
			checkInterval = statusCheckTimeDefault
		}
		time.Sleep(checkInterval)
	}
}
