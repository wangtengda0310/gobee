package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	// if return 需要多一个测试用例
	//else 需要多一个测试用例
	//else 前面的数据需要服用(用个列表或者栈来存储?git commit链表)
	// 局部变量如何处理?
	t.Run("if不带return只需要准备一组测试", func(t *testing.T) {

		var sourceCode = `
package test
func test(p int) int {
	if true {
		p = p+1
	}
	return p
}
`
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", sourceCode, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		// 抓换node为*ast.FuncDecl
		var funcNode *ast.FuncDecl
		for _, decl := range node.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				funcNode = f
			}
		}
		cases := generateTestCasesForFuncOrMethod(funcNode)
		assert.Equal(t, 1, len(cases))
	})

	t.Run("if不带return要能进入body", func(t *testing.T) {

		var sourceCode = `
func test(p,q int) int {
	if true {
		p = p+1
	}
	return p
}
`
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", sourceCode, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		// 抓换node为*ast.FuncDecl
		var funcNode *ast.FuncDecl
		for _, decl := range node.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				funcNode = f
			}
		}
		cases := generateTestCasesForFuncOrMethod(funcNode)
		assert.Equal(t, 1, len(cases))
		assert.Fail(t, "prepare some about p without q")
	})

	t.Run("if带return需要准备两组测试", func(t *testing.T) {
		var sourceCode = `
func test() int {
	if true {
		return 1
	}
	return 2
}
`
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", sourceCode, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		// 抓换node为*ast.FuncDecl
		var funcNode *ast.FuncDecl
		for _, decl := range node.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				funcNode = f
			}
		}
		cases := generateTestCasesForFuncOrMethod(funcNode)
		assert.Equal(t, 2, len(cases))
		var branches int
		assert.Equal(t, 2, branches)

	})
	t.Run("if前的数据需要复用", func(t *testing.T) {
		// 收集够了if前的数据

		var sourceCode = `
func test(p,p int) int {
	if true {
		p = p+1
	} else {
		return q
	}
	return p
}
`
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", sourceCode, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		// 抓换node为*ast.FuncDecl
		var funcNode *ast.FuncDecl
		for _, decl := range node.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				funcNode = f
			}
		}
		cases := generateTestCasesForFuncOrMethod(funcNode)
		assert.Fail(t, "prepare some about p without q for case 1", cases[0])
		assert.Fail(t, "prepare some about p without q for case 2", cases[1])
	})
}
