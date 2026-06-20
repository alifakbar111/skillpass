package niljson

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "niljson",
	Doc:      "detects nullable struct fields without omitempty in json tags",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// isNullableType returns true for types that can be nil at runtime.
func isNullableType(t types.Type) bool {
	switch t := t.(type) {
	case *types.Slice:
		return true
	case *types.Pointer:
		return true
	case *types.Map:
		return true
	case *types.Named:
		// json.RawMessage is []byte underneath — also nullable
		if t.Obj().Name() == "RawMessage" && t.Obj().Pkg() != nil &&
			strings.HasSuffix(t.Obj().Pkg().Path(), "encoding/json") {
			return true
		}
		return isNullableType(t.Underlying())
	case *types.Interface:
		return true
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		ts := n.(*ast.TypeSpec)
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			return
		}

		for _, field := range st.Fields.List {
			if len(field.Names) == 0 || field.Tag == nil {
				continue
			}
			tagValue := field.Tag.Value
			if !strings.Contains(tagValue, "json:") {
				continue
			}
			if strings.Contains(tagValue, ",omitempty") {
				continue
			}

			obj := pass.TypesInfo.Defs[field.Names[0]]
			if obj == nil {
				continue
			}
			if !isNullableType(obj.Type()) {
				continue
			}

			pass.Reportf(field.Pos(), "struct %s: field %s (%s) missing omitempty in json tag",
				ts.Name.Name, field.Names[0].Name, obj.Type().String())
		}
	})

	return nil, nil
}
