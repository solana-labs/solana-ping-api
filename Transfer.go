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
	FeePayer            types.Account
	RequestComputeUnits uint32
	ComputeUnitPrice    uint64 // micro lamports
	ReceiverPubkey      string
}

func SendPingTx(param SendPingTxParam) (string, string, PingResultError) {
	retry := 5
	errRecords := make([]string, 0, retry)

	for retry > 0 {
		retry -= 1
		time.Sleep(10 * time.Millisecond)

		// get a recnet blockhash
		latestBlockhashResponse, err := param.Client.GetLatestBlockhashWithConfig(
			context.Background(),
			client.GetLatestBlockhashConfig{
				Commitment: rpc.CommitmentConfirmed,
			},
		)
		if err != nil {
			errRecords = append(errRecords, fmt.Sprintf("failed to get the latest blockhash, err: %v", err))
			continue
		}
		blockhash := latestBlockhashResponse.Blockhash

		// generate a random amount for trasferring
		rand.Seed(time.Now().UnixNano())
		amount := uint64(rand.Intn(1000)) + 1

		// constructing a tx
		tx, err := types.NewTransaction(types.NewTransactionParam{
			Signers: []types.Account{param.FeePayer},
			Message: types.NewMessage(types.NewMessageParam{
				FeePayer:        param.FeePayer.PublicKey,
				RecentBlockhash: blockhash,
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
			errRecords = append(errRecords, fmt.Sprintf("failed to create a ping tx, err: %v", err))
			continue
		}

		// send the tx
		txhash, err := param.Client.SendTransactionWithConfig(
			context.Background(),
			tx,
			client.SendTransactionConfig{
				PreflightCommitment: rpc.CommitmentConfirmed,
			},
		)
		if err != nil {
			errRecords = append(errRecords, fmt.Sprintf("failed to send the ping tx, err: %v", err))
			continue
		}
		return txhash, blockhash, PingResultError("")
	}

	return "", "", PingResultError(fmt.Sprintf("failed to send a ping tx, errs: %v", errRecords))
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

func waitConfirmation2(c *client.Client, txHash, blockhash string) PingResultError {
	startTime := time.Now()
	endTime := startTime.Add(3 * time.Minute)

	for time.Now().Before(endTime) {
		time.Sleep(1 * time.Second)

		// check if blockhash is valid
		isBlockhashValid, err := c.IsBlockhashValidWithConfig(
			context.Background(),
			blockhash,
			client.IsBlockhashValidConfig{
				Commitment: rpc.CommitmentConfirmed,
			},
		)
		if err != nil {
			continue
		}
		if !isBlockhashValid {
			return PingResultError(fmt.Sprintf("blockhash is not valid, txHash: %v, blockhash: %v, err: %v", txHash, blockhash, err))
		}

		// check txhash status
		getSignatureStatus, err := c.GetSignatureStatus(context.Background(), txHash)
		if err != nil {
			continue
		}
		if getSignatureStatus == nil {
			continue
		}
		commitment := *getSignatureStatus.ConfirmationStatus
		if commitment == rpc.CommitmentConfirmed || commitment == rpc.CommitmentFinalized {
			return EmptyPingResultError
		}
	}

	return PingResultError(fmt.Sprintf("the confirmation process exceeds 3 mins, txHash: %v, blockhash: %v", txHash, blockhash))
}
