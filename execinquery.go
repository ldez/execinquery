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

			replacement := "Exec"
			var i int // the index of the query argument
			if strings.Contains(selector.Sel.Name, "Context") {
				replacement = "ExecContext"
				i = 1
			}

			if len(n.Args) <= i {
				return
			}

			query := getQueryString(n.Args[i])
			if query == "" {
				return
			}

			query = strings.TrimSpace(cleanValue(query))
			cmd, _, _ := strings.Cut(query, " ")
			cmd = strings.ToTitle(cmd)

			if strings.HasPrefix(cmd, "SELECT") {
				return
			}

			pass.Reportf(n.Fun.Pos(), "Use %s instead of %s to execute `%s` query", replacement, selector.Sel.Name, cmd)
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
