package srpfunc

import "github.com/voynic/srp"

type myStruct struct {
}

func (m *myStruct) test(param1 int, param2 string) string {
	srp.Hash()
	return ""
}
