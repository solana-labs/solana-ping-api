package main

import (
	"errors"
	"strings"
)

// self defined Errors
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
	TransactionLoss    = errors.New("TransactionLoss")
)

// response errors
var (
	BlockhashNotFoundText         = `rpc response error: {"code":-32002,"message":"Transaction simulation failed: Blockhash not found","data":{"accounts":null,"err":"BlockhashNotFound","logs":[],"unitsConsumed":0}}`
	RPCServerDeadlineExceededText = `rpc: call error, err: failed to do request, err: Post "https://api.internal.mainnet-beta.solana.com": context deadline exceeded`
	ServiceUnavilable503Text      = `rpc: call error, err: get status code: 503, body: <html><body><h1>503 Service Unavailable</h1>
	No server is available to handle this request.
	</body></html>`
	NumSlotsBehindText = `{count:5 : rpc response error: {"code":-32005,"message":"Node is behind by 153 slots","data":{"numSlotsBehind":153}}`
	RPCEOFText         = `rpc: call error, err: failed to do request, err: Post "https://api.internal.mainnet-beta.solana.com": EOF, body: `
)

// ping response error type
type PingResultError string

// create ping response errors , identify keys and short-descriptions of responses
var (
	ErrBlockhashNotFound           = PingResultError(BlockhashNotFoundText)
	KeyBlockhashNotFound           = "BlockhashNotFound"
	ShortKeyBlockhashNotFound      = "BlockhashNotFound"
	ErrRPCServerDeadlineExceeded   = PingResultError(RPCServerDeadlineExceededText)
	KeyRPCServerDeadlineExceeded   = "context deadline exceeded"
	ShortRPCServerDeadlineExceeded = "post to api server context dealine exceeded"
	ErrServiceUnavilable503        = PingResultError(ServiceUnavilable503Text)
	KeyServiceUnavilable503        = "status code: 503"
	ShortServiceUnavilable503      = "code: 503, no server"
	ErrNumSlotsBehind              = PingResultError(NumSlotsBehindText)
	KeyNumSlotsBehind              = "numSlotsBehind"
	ShortNumSlotsBehind            = "numSlotsBehind"
	ErrRPCEOF                      = PingResultError(RPCEOFText)
	KeyRPCEOF                      = "EOF"
	ShortKeyRPCEOF                 = "rpc error EOF"
)

// Setup Statistic / Alert / Report Error Exception List
var (
	// Error which does not use in Statistic computation
	StatisticErrorExceptionList []PingResultError
	// Error not show in slack alert
	AlertErrorExceptionList []PingResultError
	// Error not show in the report Error List
	ReportErrorExceptionList []PingResultError
	// error that does not add in Take Time, Thus account as 0 , but other error count as WaitConfirmationTimeout
	PingTakeTimeErrExpectionList []PingResultError
)

func (e PingResultError) IsBlockhashNotFound() bool {
	if strings.Contains(string(e), KeyBlockhashNotFound) {
		return true
	}
	return false
}

func (e PingResultError) IsRPCServerDeadlineExceeded() bool {
	if strings.Contains(string(e), KeyRPCServerDeadlineExceeded) {
		return true
	}
	return false
}

func (e PingResultError) IsServiceUnavilable() bool {
	if strings.Contains(string(e), KeyServiceUnavilable503) {
		return true
	}
	return false
}

func (e PingResultError) IsNumSlotsBehind() bool {
	if strings.Contains(string(e), KeyNumSlotsBehind) {
		return true
	}
	return false
}

func (e PingResultError) IsErrRPCEOF() bool {
	if strings.Contains(string(e), KeyRPCEOF) {
		return true
	}
	return false
}

func (p PingResultError) IsInErrorList(inErrs []PingResultError) bool {
	for _, e := range inErrs {
		switch e {
		case ErrBlockhashNotFound:
			if strings.Contains(string(p), KeyBlockhashNotFound) {
				return true
			}
		case ErrRPCServerDeadlineExceeded:
			if strings.Contains(string(p), KeyRPCServerDeadlineExceeded) {
				return true
			}
		case ErrRPCServerDeadlineExceeded:
			if strings.Contains(string(p), KeyServiceUnavilable503) {
				return true
			}
		case ErrNumSlotsBehind:
			if strings.Contains(string(p), KeyNumSlotsBehind) {
				return true
			}
		case ErrRPCEOF:
			if strings.Contains(string(p), KeyRPCEOF) {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func StatisticErrExpectionInit() []PingResultError {
	StatisticErrorExceptionList := []PingResultError{}
	StatisticErrorExceptionList = append(StatisticErrorExceptionList, ErrBlockhashNotFound)
	return StatisticErrorExceptionList
}

func AlertErrExpectionInit() []PingResultError {
	AlertErrorExceptionList := []PingResultError{}
	AlertErrorExceptionList = append(AlertErrorExceptionList, ErrRPCServerDeadlineExceeded)
	AlertErrorExceptionList = append(AlertErrorExceptionList, ErrBlockhashNotFound)
	return AlertErrorExceptionList
}

func ReportErrExpectionInit() []PingResultError {
	ReportErrorExceptionList := []PingResultError{}
	return ReportErrorExceptionList
}

func PingTakeTimeErrExpectionInit() []PingResultError {
	PingTakeTimeErrExpectionList := []PingResultError{}
	PingTakeTimeErrExpectionList = append(PingTakeTimeErrExpectionList, ErrBlockhashNotFound)
	return PingTakeTimeErrExpectionList
}
