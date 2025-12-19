package csv

import (
	"github.com/go-spring/spring-core/gs"
)

type Loader interface {
	Load()
}
type DataHolder interface {
	CsvStructMapping() map[string]any
}

func init() {
	gs.Object(&Gocsvimpl{}).Export(gs.As[Loader]())
}
