package main

import (
	"errors"
	"strings"
)

var (
	InvalidCluster     = errors.New("invalid cluster")
	FindIndexNotFound  = errors.New("findIndex does not find pattern")
	ParseMessageError  = errors.New("parse message error")
	ConvertWrongType   = errors.New("parse result convert to type fail")
	ParseSplitError    = errors.New("split message fail")
	ResultInvalid      = errors.New("invalid Result")
	NoPingResultFound  = errors.New("no Ping Result")
	NoPingResultRecord = errors.New("no Ping Result Record")
)

var BlockhashNotFoundText = `rpc response error: {"code":-32002,"message":"Transaction simulation failed: Blockhash not found","data":{"accounts":null,"err":"BlockhashNotFound","logs":[],"unitsConsumed":0}}`
var RPCServerDeadlineExceededText = `rpc: call error, err: failed to do request, err: Post "https://api.internal.mainnet-beta.solana.com": context deadline exceeded`

type PingResultError string

var (
	BlockhashNotFound         = PingResultError("BlockhashNotFound")
	RPCServerDeadlineExceeded = PingResultError("context deadline exceeded")
)

func (e PingResultError) IsBlockhashNotFound() bool {
	if strings.Contains(string(e), string(BlockhashNotFound)) {
		return true
	}
	return false
}

func (e PingResultError) IsRPCServerDeadlineExceeded() bool {
	if strings.Contains(string(e), string(RPCServerDeadlineExceeded)) {
		return true
	}
	return false
}
func (e PingResultError) IsInErrorList(inErrs []PingResultError) bool {
	for _, e := range inErrs {
		switch e {
		case BlockhashNotFound:
			return e.IsBlockhashNotFound()
		case RPCServerDeadlineExceeded:
			return e.IsRPCServerDeadlineExceeded()
		default:
			return false
		}

	}
	return false
}
