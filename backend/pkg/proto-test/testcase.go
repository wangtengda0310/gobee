package prototest

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/cases"
)

// testcase 相关函数已移至 cases 包，此处保留别名供包内引用
var (
	SaveTestCaseToFile         = cases.SaveTestCaseToFile
	AppendTestCaseToFile       = cases.AppendTestCaseToFile
	LoadTestCaseFromFile       = cases.LoadTestCaseFromFile
	NormalizeTestCaseRecording = cases.NormalizeTestCaseRecording
	IsReqDirection             = cases.IsReqDirection
)

type TestCaseEntry = cases.TestCaseEntry
