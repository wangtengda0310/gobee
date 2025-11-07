package exec

import (
	"os/exec"
)

const (
	In        = "in"
	TpZip     = "zip"
	TpSql     = "sql-tpl"
	TpDir     = "dir"
	TpConsole = "-"
)

type Acceptor func(visitor Visitor, where string, args ...string)
type Visitor func(*exec.Cmd)

var acci Acceptor

func VisitImport(db Visitor, where string, args ...string) {
	acci(db, where, args...)
}

func Accept(dsProvider Acceptor) {
	acci = dsProvider
}
