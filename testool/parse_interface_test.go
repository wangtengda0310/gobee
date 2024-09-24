package main

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/Xuanwo/gg"
)

func Test_parseTemplate(t *testing.T) {
	(&methodSignature{
		Name:    "test",
		Params:  [][]string{{"t", "*testing.T"}, {"attributeOwner", "*MockIUserData"}},
		Returns: []string{"func(t *testing.T, attributeOwner *MockIUserData)"},
	}).parseTemplate("withMockerMethod")
}

func TestSample(t *testing.T) {
	sourceCode := `package main

import (
	"gforge/common/ctx"
	"gforge/pb"
)

type Owner interface {
	GetModel(string) Model
	SetModel(string, Model)
}

`

	gen("", sourceCode)
}
func TestGGInterface(t *testing.T) {
	is := &interfaceSignature{Name: "Test", Methods: []*methodSignature{{
		Name:    "test",
		Params:  [][]string{{"t", "*testing.T"}, {"attributeOwner", "*MockIUserData"}},
		Returns: []string{"func(t *testing.T, attributeOwner *MockIUserData)"},
	}}}
	generator := New()
	ng := generator.NewGroup()
	newStruct := ng.NewStruct(fmt.Sprint("Mock", is.Name, "Impl"))
	for _, method := range is.Methods {
		newStruct.
			AddField(method.Name+"Func", fmt.Sprint(method.Name, "Func"))

	}
	fmt.Println(ng.String())

}
func TestGGFuncs(t *testing.T) {
	var InterfaceName = "IUserData"
	var ms = &methodSignature{
		Name:    "test",
		Params:  [][]string{{"t", "*testing.T"}, {"attributeOwner", "*MockIUserData"}},
		Returns: []string{"func(t *testing.T, attributeOwner *MockIUserData)"},
	}

	genBygg(InterfaceName, ms)
}

func genBygg(InterfaceName string, ms *methodSignature) {
	generator := New()

	f := generator.NewGroup()
	f.AddPackage("test")
	f.NewStruct(fmt.Sprint("Mock", InterfaceName)).AddField("impl", "Mock"+InterfaceName+"Impl")

	f.AddType(fmt.Sprint(ms.Name, "Func"), fmt.Sprint("func(", ms.ParamsWithType(), ")(", strings.Join(ms.Returns, ""), ")"))
	body := f.NewFunction(ms.Name).
		WithReceiver("m", fmt.Sprint("*Mock", InterfaceName)).
		AddParameter("t", "*testing.T").
		AddBody(
			"return m.impl." + ms.Name + "Func()",
		)

	addBody := f.NewFunction(fmt.Sprint("with", ms.Name)).
		AddParameter("t", "*testing.T").
		AddParameter("", strings.Join(ms.ReturnsWithVarName(), ",")).
		AddBody(
			"t.Helper()",
			"return func(m *Mock"+InterfaceName+") {",
			"m.impl."+ms.Name+"Func = func("+ms.ParamsWithType()+")("+strings.Join(ms.Returns, "")+") {",
			"return "+strings.Join(ms.ReturnsWithoutVarName(), ","),
			"}",
			"}",
		)
	for _, s := range ms.ReturnsWithVarName() {
		addBody.AddParameter(s, "")
		body.AddResult("", s)
	}

	err := generator.WriteFile(fmt.Sprint("mock_", InterfaceName, ".go"))
	if err != nil {
		fmt.Println(f.String())
	}
}
