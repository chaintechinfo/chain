// +build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func main() {
	ops := getOps()
	opinfoName := txvmFile("opinfo.go")
	out, err := os.Create(opinfoName)
	must(err)
	fmt.Fprint(out, "// Auto-generated from ops.go by gen.go\n\npackage txvm2\n\n")

	fmt.Fprint(out, "var opNames = [...]string{\n")
	for _, op := range ops {
		fmt.Fprintf(out, "\t%s: \"%s\",\n", op, strings.ToLower(op[2:]))
	}
	fmt.Fprint(out, "}\n\n")

	fmt.Fprint(out, "var opCodes = map[string]byte{\n")
	for _, op := range ops {
		fmt.Fprintf(out, "\t\"%s\": %s,\n", strings.ToLower(op[2:]), op)
	}
	fmt.Fprint(out, "}\n\n")

	fmt.Fprint(out, "var opFuncs = [...]func(*vm){\n")
	for _, op := range ops {
		switch op {
		case "OpCommand", "OpSatisfy":
			// do nothing - avoid initialization loop
		default:
			fmt.Fprintf(out, "\t%s: %c%s,\n", op, unicode.ToLower(rune(op[0])), op[1:])
		}
	}
	fmt.Fprint(out, "}\n\n")

	out.Close()

	cmd := exec.Command("gofmt", "-w", opinfoName)
	must(cmd.Run())
}

func getOps() []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, txvmFile("ops.go"), nil, 0)
	must(err)
	var constDecl *ast.GenDecl
	for _, d := range f.Decls {
		if gendecl, ok := d.(*ast.GenDecl); ok && gendecl.Tok == token.CONST {
			constDecl = gendecl
			break
		}
	}
	if constDecl == nil {
		panic("ops.go has no top-level const declaration")
	}
	var ops []string
	for _, spec := range constDecl.Specs {
		vspec, ok := spec.(*ast.ValueSpec)
		if !ok {
			panic("const decl contains non-const values?!")
		}
		if len(vspec.Names) != 1 {
			panic(fmt.Errorf("const spec contains %d names, want 1", len(vspec.Names)))
		}
		name := vspec.Names[0].Name
		if name == "Op0" {
			continue
		}
		if !strings.HasPrefix(name, "Op") {
			continue
		}
		ops = append(ops, name)
	}
	return ops
}

func txvmFile(name string) string {
	return os.Getenv("CHAIN") + "/protocol/txvm2/" + name
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
