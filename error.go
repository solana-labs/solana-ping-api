package main

import (
	"errors"
)

// ping response error type
type PingResultError string

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

// create ping response errors , identify keys and short-descriptions of responses
// var (
// 	ErrGatewayTimeout504   = PingResultError(GatewayTimeout504Text)
// 	KeyGatewayTimeout504   = "code: 504"
// 	ShortGatewayTimeout504 = "504-gateway-timeout"
// )

// Setup Statistic / Alert / Report Error Exception List
var (
	// Error which does not use in Statistic computation
	StatisticErrorExceptionList []ErrorIdentifier
	// Error does not show in slack alert
	AlertErrorExceptionList []ErrorIdentifier
	// Error does not show in the report Error List
	ReportErrorExceptionList []ErrorIdentifier
	// error that does not be added into TakeTime
	PingTakeTimeErrExpectionList []ErrorIdentifier
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

func (p PingResultError) IsInErrorList(inErrs []ErrorIdentifier) bool {
	for _, idf := range inErrs {
		if idf.IsIdentical(p) {
			return true
		}
	}
	return false
}

func StatisticErrExpectionInit() []ErrorIdentifier {
	StatisticErrorExceptionList = []ErrorIdentifier{}
	StatisticErrorExceptionList = append(StatisticErrorExceptionList, BlockhashNotFound)
	StatisticErrorExceptionList = append(StatisticErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return StatisticErrorExceptionList
}

func AlertErrExpectionInit() []ErrorIdentifier {
	AlertErrorExceptionList = []ErrorIdentifier{}
	AlertErrorExceptionList = append(AlertErrorExceptionList, RPCServerDeadlineExceeded)
	AlertErrorExceptionList = append(AlertErrorExceptionList, BlockhashNotFound)
	AlertErrorExceptionList = append(AlertErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return AlertErrorExceptionList
}

func ReportErrExpectionInit() []ErrorIdentifier {
	ReportErrorExceptionList = []ErrorIdentifier{}
	ReportErrorExceptionList = append(ReportErrorExceptionList, TransactionHasAlreadyBeenProcessed)
	return ReportErrorExceptionList
}

func PingTakeTimeErrExpectionInit() []ErrorIdentifier {
	PingTakeTimeErrExpectionList = []ErrorIdentifier{}
	return PingTakeTimeErrExpectionList
}
