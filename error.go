package main

import "errors"

var FindIndexNotFound = errors.New("FindIndex does not find pattern")
var ParseMessageError = errors.New("Parse message error")
var ConvertWrongType = errors.New("Parse result convert to type fail")
var ParseSplitError = errors.New("Split message fail")
var ResultInvalid = errors.New("Invalid Result")
var NoPingResultFound = errors.New("No Ping Result")
var NoPingResultRecord = errors.New("No Ping Result Record")
