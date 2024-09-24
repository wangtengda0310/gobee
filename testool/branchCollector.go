package main

import "go/ast"

// 这个方法为每个函数或方法生成测试用例,调用者可以ast.Inspect回调中返回false以免重复遍历
func generateTestCasesForFuncOrMethod(funcNode *ast.FuncDecl) []any {
	// if not function body or method body return nil
	if funcNode == nil {
		return nil
	}
	cases := []any{}
	ast.Inspect(funcNode, func(n ast.Node) bool {
		// 检查节点是否为if语句
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		// 检查if语句是否有return语句
		hasReturn := false
		ast.Inspect(ifStmt, func(n ast.Node) bool {
			if _, ok := n.(*ast.ReturnStmt); ok {
				hasReturn = true
				return false
			}
			return true
		})

		if hasReturn {
			// create a test case for if with return
			cases = append(cases, "if with return")
		}

		return true
	})
	return cases
}
