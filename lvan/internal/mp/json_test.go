package mp

import (
	"testing"

	"github.com/k0kubun/pp"
)

func TestName(t *testing.T) {
	pp.Printf("Hello, %s %v!\n", "world", struct {
		Name string
		Age  int
	}{"Alice", 30})
	panic(pp.Errorf("%v", struct {
		Name string
		Age  int
	}{"Bob", 25}))
}
