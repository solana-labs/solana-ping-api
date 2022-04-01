package main

import (
	"errors"
	"log"
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
	NoPingResultShort  = errors.New("PingResultError has no shortname")
)
var (
	BlockhashNotFoundText         = `rpc response error: {"code":-32002,"message":"Transaction simulation failed: Blockhash not found","data":{"accounts":null,"err":"BlockhashNotFound","logs":[],"unitsConsumed":0}}`
	RPCServerDeadlineExceededText = `rpc: call error, err: failed to do request, err: Post "https://api.internal.mainnet-beta.solana.com": context deadline exceeded`
	ServiceUnavilable503Text      = `rpc: call error, err: get status code: 503, body: <html><body><h1>503 Service Unavailable</h1>
	No server is available to handle this request.
	</body></html>`
	NumSlotsBehindText = `{count:5 : rpc response error: {"code":-32005,"message":"Node is behind by 153 slots","data":{"numSlotsBehind":153}}`
)

type PingResultError string

var (
	BlockhashNotFound            = PingResultError(BlockhashNotFoundText)
	BlockhashNotFoundKey         = PingResultError("BlockhashNotFound")
	RPCServerDeadlineExceeded    = PingResultError(RPCServerDeadlineExceededText)
	RPCServerDeadlineExceededKey = PingResultError("context deadline exceeded")
	ServiceUnavilable503         = PingResultError(ServiceUnavilable503Text)
	ServiceUnavilable503Key      = PingResultError("status code: 503")
	NumSlotsBehind               = PingResultError(NumSlotsBehindText)
	NumSlotsBehindKey            = PingResultError("numSlotsBehind")
)

func (e PingResultError) IsBlockhashNotFound() bool {
	if strings.Contains(string(e), string(BlockhashNotFoundKey)) {
		return true
	}
	return false
}

func (e PingResultError) IsRPCServerDeadlineExceeded() bool {
	if strings.Contains(string(e), string(RPCServerDeadlineExceededKey)) {
		return true
	}
	return false
}
func (p PingResultError) IsInErrorList(inErrs []PingResultError) bool {
	for _, e := range inErrs {
		switch e {
		case BlockhashNotFound:
			if strings.Contains(string(p), string(BlockhashNotFoundKey)) {
				return true
			}
		case RPCServerDeadlineExceeded:
			if strings.Contains(string(p), string(RPCServerDeadlineExceededKey)) {
				return true
			}
		default:
			log.Println("--->Not in the List: ", p)
			return false
		}
	}
	return false
}
