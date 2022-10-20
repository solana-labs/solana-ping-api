package main

import (
	"errors"
	"strings"
)

// ping response error type
type PingResultError string

const EmptyPingResultError = ""

// self defined Errors
var (
	ErrInvalidCluster          = errors.New("invalid cluster")
	ErrFindIndexNotFound       = errors.New("findIndex does not find pattern")
	ErrParseMessageError       = errors.New("parse message error")
	ErrConvertWrongType        = errors.New("parse result convert to type fail")
	ErrParseSplit              = errors.New("split message fail")
	ErrResultInvalid           = errors.New("invalid Result")
	ErrNoPingResult            = errors.New("no Ping Result")
	ErrNoPingResultRecord      = errors.New("no Ping Result Record")
	ErrNoPingResultShort       = errors.New("PingResultError has no shortname")
	ErrTransactionLoss         = errors.New("TransactionLoss")
	ErrWaitForConfirmedTimeout = errors.New("Wait for a confirmed block timeout")
	ErrGetKeyPair              = errors.New("No valid KeyPair")
	ErrKeyPairFile             = errors.New("Read KeyPair File Error")
)

// Setup Statistic / Alert / Report Error Exception List
var (
	ResponseErrIdentifierList []ErrRespIdentifier
	// Error which does not use in Statistic computation
	StatisticErrorExceptionList []ErrRespIdentifier
	// Error does not show in slack alert
	AlertErrorExceptionList []ErrRespIdentifier
	// Error does not show in the report Error List
	ReportErrorExceptionList []ErrRespIdentifier
	// error that does not be added into TakeTime
	PingTakeTimeErrExpectionList []ErrRespIdentifier
)

func (e PingResultError) IsBlockhashNotFound() bool {
	return BlockhashNotFound.IsIdentical(e)
}
func (e PingResultError) IsTransactionHasAlreadyBeenProcessed() bool {
	return TransactionHasAlreadyBeenProcessed.IsIdentical(e)
}

func (e PingResultError) IsRPCServerDeadlineExceeded() bool {
	return RPCServerDeadlineExceeded.IsIdentical(e)
}

func (e PingResultError) IsServiceUnavilable() bool {
	return ServiceUnavilable503.IsIdentical(e)
}

func (e PingResultError) IsTooManyRequest429() bool {
	return TooManyRequest429.IsIdentical(e)
}

func (e PingResultError) IsNumSlotsBehind() bool {
	return NumSlotsBehind.IsIdentical(e)
}

func (e PingResultError) IsErrRPCEOF() bool {
	return RPCEOF.IsIdentical(e)
}
func (e PingResultError) IsErrGatewayTimeout504() bool {
	return GatewayTimeout504.IsIdentical(e)
}
func (e PingResultError) IsNoSuchHost() bool {
	return NoSuchHost.IsIdentical(e)
}

func (p PingResultError) IsInErrorList(inErrs []ErrRespIdentifier) bool {
	for _, idf := range inErrs {
		if idf.IsIdentical(p) {
			return true
		}
	}
	return false
}
func (p PingResultError) Short() string {
	for _, idf := range ResponseErrIdentifierList {
		if idf.IsIdentical(p) {
			return idf.Short
		}
	}
	return string(p)
}

func (p PingResultError) Subsitute(old string, new string) string {
	return strings.ReplaceAll(string(p), old, new)
}

func (p PingResultError) NoError() bool {
	if string(p) == string(EmptyPingResultError) {
		return true
	}
	return false
}

func ResponseErrIdentifierInit() []ErrRespIdentifier {
	ResponseErrIdentifierList = []ErrRespIdentifier{}
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, BlockhashNotFound)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, TransactionHasAlreadyBeenProcessed)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, RPCServerDeadlineExceeded)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, ServiceUnavilable503)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, TooManyRequest429)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, NumSlotsBehind)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, RPCEOF)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, GatewayTimeout504)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, NoSuchHost)
	ResponseErrIdentifierList = append(ResponseErrIdentifierList, TxHasAlreadyProcess)
	return ResponseErrIdentifierList
}

func StatisticErrExpectionInit() []ErrRespIdentifier {
	StatisticErrorExceptionList = []ErrRespIdentifier{}
	StatisticErrorExceptionList = append(StatisticErrorExceptionList, BlockhashNotFound)
	StatisticErrorExceptionList = append(StatisticErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return StatisticErrorExceptionList
}

func AlertErrExpectionInit() []ErrRespIdentifier {
	AlertErrorExceptionList = []ErrRespIdentifier{}
	AlertErrorExceptionList = append(AlertErrorExceptionList, RPCServerDeadlineExceeded)
	AlertErrorExceptionList = append(AlertErrorExceptionList, BlockhashNotFound)
	AlertErrorExceptionList = append(AlertErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return AlertErrorExceptionList
}

func ReportErrExpectionInit() []ErrRespIdentifier {
	ReportErrorExceptionList = []ErrRespIdentifier{}
	ReportErrorExceptionList = append(ReportErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return ReportErrorExceptionList
}

func PingTakeTimeErrExpectionInit() []ErrRespIdentifier {
	PingTakeTimeErrExpectionList = []ErrRespIdentifier{}
	return PingTakeTimeErrExpectionList
}
