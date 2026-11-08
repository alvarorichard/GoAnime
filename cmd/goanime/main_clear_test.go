package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// main ends in os.Exit and cannot be called from a test, so what it does with
// the signals FlagParser returns is checked in its source: the request to
// clear history has to reach handlers.HandleClearRequest, report its error and
// stop, instead of falling through to a search.

// clearBranchProblems says what is wrong with how src handles
// util.ErrClearRequested; empty when it is handled as it should be.
func clearBranchProblems(src string) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.SkipObjectResolution)
	if err != nil {
		return []string{"main.go does not parse: " + err.Error()}
	}

	var branch *ast.IfStmt
	ast.Inspect(file, func(n ast.Node) bool {
		ifs, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		cond, ok := ifs.Cond.(*ast.BinaryExpr)
		if ok && cond.Op == token.EQL && isIdent(cond.X, "err") && isPkgSelector(cond.Y, "util", "ErrClearRequested") {
			branch = ifs
		}
		return true
	})
	if branch == nil {
		return []string{"no branch for `err == util.ErrClearRequested`: --clear-* would fall through to a search"}
	}

	var calls, fails, returns int
	ast.Inspect(branch.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if isPkgSelector(node.Fun, "handlers", "HandleClearRequest") {
				calls++
			}
			if isIdent(node.Fun, "fail") {
				fails++
			}
		case *ast.ReturnStmt:
			returns++
		}
		return true
	})

	var problems []string
	if calls != 1 {
		problems = append(problems, "the clear branch must call handlers.HandleClearRequest once")
	}
	if fails != 1 {
		problems = append(problems, "the clear branch must report a failure through fail() once")
	}
	if returns != 1 {
		problems = append(problems, "the clear branch must return once: it has to stop after clearing")
	}
	return problems
}

func isIdent(expr ast.Expr, name string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == name
}

func isPkgSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg && sel.Sel.Name == name
}

func TestMainHandsTheClearRequestToItsHandlerAndStops(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if problems := clearBranchProblems(string(src)); len(problems) > 0 {
		t.Errorf("main.go: %s", strings.Join(problems, "; "))
	}
}

// The check above is only worth its verdict if it can fail: these are the ways
// main.go could get the branch wrong.
func TestTheClearBranchCheckSeesABrokenMain(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	good := string(raw)

	const branch = "if err == util.ErrClearRequested {\n\t\t\tif clearErr := handlers.HandleClearRequest(); clearErr != nil {\n\t\t\t\tfail(clearErr)\n\t\t\t}\n\t\t\treturn\n\t\t}"
	if !strings.Contains(good, branch) {
		t.Fatalf("main.go no longer has the branch this test breaks on purpose; update the test with it:\n%s", branch)
	}
	for name, broken := range map[string]string{
		"no branch at all":            strings.Replace(good, branch, "", 1),
		"keeps going after clearing":  strings.Replace(good, branch, strings.Replace(branch, "\t\t\treturn\n", "", 1), 1),
		"drops the error":             strings.Replace(good, branch, strings.Replace(branch, "if clearErr := handlers.HandleClearRequest(); clearErr != nil {\n\t\t\t\tfail(clearErr)\n\t\t\t}", "_ = handlers.HandleClearRequest()", 1), 1),
		"never calls the handler":     strings.Replace(good, branch, strings.Replace(branch, "handlers.HandleClearRequest()", "handlers.HandleUpdateRequest()", 1), 1),
		"waits on another error":      strings.Replace(good, "err == util.ErrClearRequested", "err == util.ErrUpdateRequested", 1),
		"main.go that does not parse": "package main\nfunc (",
	} {
		if len(clearBranchProblems(broken)) == 0 {
			t.Errorf("%s: the check found nothing wrong", name)
		}
	}
	if problems := clearBranchProblems(good); len(problems) != 0 {
		t.Errorf("the real main.go is reported as broken: %v", problems)
	}
}
