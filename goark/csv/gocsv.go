package csv

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"
)

type Gocsvimpl struct {
	file    string       `value:"${csvfile}"`
	Holders []DataHolder `autowire:"*"`
}

func (i Gocsvimpl) Load() {
	for _, holder := range i.Holders {
		m := holder.CsvStructMapping()
		for file, record := range m {
			open, err2 := os.Open(filepath.Join(i.file, file))
			if err2 != nil {
				log.Panic(err2)
			}

			err := gocsv.UnmarshalFile(open, record)
			if err != nil {
				println(err.Error())
			}
			open.Close()
		}
	}

	println("load ok")
}
