package internal

import (
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

type Acceptor func(visitor Visitor, where string, args ...string)
type Visitor func(dump.Datasource)

var acce Acceptor
var acci Acceptor

func VisitExport(db Visitor, where string, args ...string) {
	acce(db, where, args...)
}

func VisitImport(db Visitor, where string, args ...string) {
	acci(db, where, args...)
}

func Accept(dsProvider Acceptor) {
	acci = dsProvider
	acce = dsProvider
}

var TransImport func(string) []dump.Record
var TransExport func([]dump.Record, ...string) string
