package reflectparam

import (
	"reflect"
)

type Args struct {
	S interface{} `csv:"name"`
	V int         `csv:"value"`
}

func f(s string, v any) Args {
	//split := strings.Split(s, " ")
	//ee := Args{}
	//ee.S = split[0]
	//atoi, _ := strconv.Atoi(split[1])
	//ee.V = atoi
	//elem := reflect.ValueOf(&e).Elem()

	rv := reflect.ValueOf(v)
	rv.Field(0).SetString("123")

	return rv.Elem().Interface().(Args)
}

func parse[t any](s string, e t) Args {
	return f(s, &Args{})

}
