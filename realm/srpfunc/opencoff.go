package srpfunc

import "fmt"
import "github.com/opencoff/go-srp"

type myStruct1 struct {
}

func (m *myStruct1) test(param1 string, param2 string) string {
	bits := 1024
	pass := []byte("password string that's too long")
	i := []byte("foouser")

	s, err := srp.New(bits)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	fmt.Println(pass)
	fmt.Println(i)
	return ""
}
