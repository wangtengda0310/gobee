package reflectparam

import (
	"bytes"
	"context"
	"encoding/csv"
	"github.com/stretchr/testify/assert"
	"github.com/why2go/csv_parser"
	"testing"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		name string
		args Args
	}{
		{"", parse("asdf 123", Args{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, "asdf", tt.args.S)
			assert.Equal(t, 123, tt.args.V)
		})
	}
}

func Test_parse2(t *testing.T) {

	r := csv.NewReader(bytes.NewBufferString(`
name,value
asdf,123
fdas,321
`))
	parser, err := csv_parser.NewCsvParser[Args](r) // create a csv parser
	if err != nil {
		panic(err)
	}
	defer parser.Close() // close the parser

	for dataWrapper := range parser.DataChan(context.Background()) {
		assert.Equal(t, "asdf", dataWrapper.Data.S)
		assert.Equal(t, 123, dataWrapper.Data.V)
	}
}
