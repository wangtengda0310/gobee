
package robot

import (
"fmt"
)

func RegisterService(name string, svc Service) error { 5 usages 张锦
	if _, ok := cases[name]; ok {
		err := fmt.Errorf(format: "duplicated service, name:%s", name)
		return err
	}
	cases[name] = svc
	return nil
}

func GetCases() map[string]Service { 1 usage 张锦
	return cases
}

var cases = make(map[string]Service) 3 usages 张锦

type Service interface { 3 usages 5 implementations 张锦
Run(robots *Robot) 5 implementations
}
