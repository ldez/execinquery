package execinquery

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "execinquery is a linter about query string checker in Query function which reads your Go src files and warning it finds"

// Analyzer is checking database/sql pkg Query's function
var Analyzer = &analysis.Analyzer{
	Name: "execinquery",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	result.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.CallExpr:
			if len(n.Args) < 1 {
				return
			}

			selector, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return
			}

			if "database/sql" != pass.TypesInfo.Uses[selector.Sel].Pkg().Path() {
				return
			}

			if !strings.Contains(selector.Sel.Name, "Query") {
				return
			}

			var i int
			if strings.Contains(selector.Sel.Name, "Context") {
				i = 1
			}

			s := getQueryString(n.Args[i])
			if s == "" {
				return
			}

			s = strings.TrimSpace(cleanValue(s))

			if strings.HasPrefix(strings.ToLower(s), "select") {
				return
			}

			s = strings.ToTitle(strings.SplitN(s, " ", 2)[0])

			pass.Reportf(n.Fun.Pos(), "It's better to use Execute method instead of %s method to execute `%s` query", selector.Sel.Name, s)
		}
	})

	return nil, nil
}

func getQueryString(exp interface{}) string {
	switch e := exp.(type) {
	case *ast.AssignStmt:
		var v string
		for _, stmt := range e.Rhs {
			v += cleanValue(getQueryString(stmt))
		}
		return v

	case *ast.BasicLit:
		return e.Value

	case *ast.ValueSpec:
		var v string
		for _, value := range e.Values {
			v += cleanValue(getQueryString(value))
		}
		return v

	case *ast.Ident:
		return getQueryString(e.Obj.Decl)

	case *ast.BinaryExpr:
		v := cleanValue(getQueryString(e.X))
		v += cleanValue(getQueryString(e.Y))
		return v
	}

	return ""
}

func cleanValue(s string) string {
	v := strings.NewReplacer(`"`, "", "`", "").Replace(s)

	if !strings.HasPrefix(v, "-- ") {
		return v
	}

	// Remove SQL comments
	index := strings.Index(v, "\n")
	if index > 0 {
		return s[index+1:]
	}

	return ""
}
