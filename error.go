package main

import "errors"

var InvalidCluster = errors.New("invalid cluster")
var FindIndexNotFound = errors.New("findIndex does not find pattern")
var ParseMessageError = errors.New("parse message error")
var ConvertWrongType = errors.New("parse result convert to type fail")
var ParseSplitError = errors.New("split message fail")
var ResultInvalid = errors.New("invalid Result")
var NoPingResultFound = errors.New("no Ping Result")
var NoPingResultRecord = errors.New("no Ping Result Record")
