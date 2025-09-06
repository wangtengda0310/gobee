package dump

type BeforeImportRecords func(records []Record)
type BeforeLoadEachData func(record Record)
type AfterImportEachData func(record Record)
type AfterImportAllData func(records []Record)

var beforeImportRecordsCallback []BeforeImportRecords

func RegisterBeforeImportRecordsCallBack(cb BeforeImportRecords) {
	beforeImportRecordsCallback = append(beforeImportRecordsCallback, cb)
}

var beforeImportEachRecordsCallback []BeforeLoadEachData

func RegisterBeforeLoadEachDataCallBack(cb BeforeLoadEachData) {
	beforeImportEachRecordsCallback = append(beforeImportEachRecordsCallback, cb)
}

var afterImportEachRecordsCallback []AfterImportEachData

func RegisterAfterImportEachDataCallBack(cb AfterImportEachData) {
	afterImportEachRecordsCallback = append(afterImportEachRecordsCallback, cb)
}

var afterImportAllRecordsCallback []AfterImportAllData

func RegisterAfterImportAllDataCallBack(cb AfterImportAllData) {
	afterImportAllRecordsCallback = append(afterImportAllRecordsCallback, cb)
}
