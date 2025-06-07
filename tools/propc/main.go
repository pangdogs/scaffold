/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package main

import (
	"bytes"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

func main() {
	cmd := &cobra.Command{
		Short: "属性同步代码生成工具。",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
			{
				declFile := viper.GetString("decl_file")
				if declFile == "" {
					panic("[--decl_file]值不能为空")
				}
				if _, err := os.Stat(declFile); err != nil {
					panic(fmt.Errorf("[--decl_file]文件错误，%s", err))
				}
			}
		},
		Run: run,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
	}
	cmd.Flags().String("decl_file", os.Getenv("GOFILE"), "属性定义文件（.go）。")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func run(*cobra.Command, []string) {
	fset, fast, fdata := loadDeclFile()

	type Op struct {
		Name          string
		Args          string
		Decl          string
		Call          string
		CallResults   string
		ReturnResults string
		Types         map[string]any
	}

	type Prop struct {
		Name string
		Ops  []Op
	}

	props := generic.UnorderedSliceMap[string, *Prop]{}

	ast.Inspect(fast, func(node ast.Node) bool {
		ts, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}

		atti := getAtti(fset, fast, node)

		if !atti.Has("sync") {
			return true
		}

		if b, err := strconv.ParseBool(atti.Get("sync")); err != nil || !b {
			return true
		}

		props.Add(ts.Name.String(), &Prop{Name: ts.Name.String()})
		return true
	})

	ast.Inspect(fast, func(node ast.Node) bool {
		fd, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}

		atti := getAtti(fset, fast, node)

		if !atti.Has("sync") {
			return true
		}

		if b, err := strconv.ParseBool(atti.Get("sync")); err != nil || !b {
			return true
		}

		if fd.Recv == nil || len(fd.Recv.List) <= 0 {
			return true
		}

		starExpr, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			return true
		}

		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		prop, ok := props.Get(ident.Name)
		if !ok {
			return true
		}

		op := Op{}

		op.Name = fd.Name.String()

		{
			for i, field := range fd.Type.Params.List {
				if i > 0 {
					op.Args += ", "
				}
				op.Args += field.Names[0].String()
			}
		}

		{
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, fd.Type)

			op.Decl = fd.Name.String() + strings.TrimPrefix(buf.String(), "func")
		}

		{
			op.Call = fd.Name.String() + "("

			for i, field := range fd.Type.Params.List {
				if i > 0 {
					op.Call += ", "
				}
				op.Call += field.Names[0].String()
			}

			op.Call += ")"
		}

		op.ReturnResults = "return"

		if fd.Type.Results != nil && len(fd.Type.Results.List) > 0 {
			op.ReturnResults += " "

			for i, field := range fd.Type.Results.List {
				if i > 0 {
					op.ReturnResults += ", "
					op.CallResults += ", "
				}

				if len(field.Names) > 0 {
					op.ReturnResults += field.Names[0].String()
					op.CallResults += field.Names[0].String()
				} else {
					op.ReturnResults += fmt.Sprintf("r%d", i+1)
					op.CallResults += fmt.Sprintf("r%d", i+1)
				}
			}

			op.CallResults += " = "
		}

		{
			op.Types = make(map[string]any)

			for _, field := range fd.Type.Params.List {
				start := fset.Position(field.Type.Pos()).Offset
				end := fset.Position(field.Type.End()).Offset
				name := string(fdata[start:end])
				op.Types[name] = struct{}{}
			}
		}

		prop.Ops = append(prop.Ops, op)

		return true
	})

	var imports []string

	for _, is := range fast.Imports {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, is)

		imports = append(imports, buf.String())
	}

	if !slices.ContainsFunc(imports, func(i string) bool {
		switch i {
		case `"git.golaxy.org/scaffold/addins/propview"`, `propview "git.golaxy.org/scaffold/addins/propview"`:
			return true
		default:
			return false
		}
	}) {
		imports = append(imports, `propview "git.golaxy.org/scaffold/addins/propview"`)
	}

	opImports := map[string]struct{}{
		"propview": {},
	}

	props.Each(func(_ string, prop *Prop) {
		for _, op := range prop.Ops {
			for t := range op.Types {
				if strings.Contains(t, ".") {
					pkg := strings.Split(strings.Trim(t, "*"), ".")
					if len(pkg) > 0 {
						opImports[pkg[0]] = struct{}{}
					}
				}
			}
		}
	})

	imports = slices.DeleteFunc(imports, func(i string) bool {
		alias := strings.Split(i, " ")
		if len(alias) >= 2 {
			if _, ok := opImports[alias[0]]; ok {
				return false
			}
			return true
		}

		if _, ok := opImports[path.Base(strings.Trim(i, `"`))]; ok {
			return false
		}

		return true
	})

	if props.Len() <= 0 {
		return
	}

	const tmpl = `
{{- .Comment}}

package {{.Package}}

import (
	{{- range .Imports}}
	{{.}}
	{{- end}}
)

{{range .Props -}}
type {{.Name}}Sync struct {
	propview.PropSyncer
	{{.Name}}
}

func (ps *{{.Name}}Sync) Load(service string) error {
	data, revision, err := propview.UnsafePropSync(ps).Load(service)
	if err != nil {
		return err
	}
	return ps.Unmarshal(data, revision)
}

func (ps *{{.Name}}Sync) Save(service string) error {
	data, revision, err := ps.Marshal()
	if err != nil {
		return err
	}
	return propview.UnsafePropSync(ps).Save(service, data, revision)
}

func (ps *{{.Name}}Sync) Managed() propview.IProp {
	return &ps.{{.Name}}
}

{{- $propName := .Name}}
{{range .Ops}}
func (ps *{{$propName}}Sync) {{.Decl}} {
	{{.CallResults}}ps.{{$propName}}.{{.Call}}
	propview.UnsafePropSync(ps).Sync(propview.UnsafeProp(&ps.{{$propName}}).IncrRevision(), "{{.Name}}", {{.Args}})
	{{.ReturnResults}}
}
{{end}}
{{end}}
`

	type TmplArgs struct {
		Comment string
		Package string
		Imports []string
		Props   []*Prop
	}

	args := &TmplArgs{
		Comment: fmt.Sprintf("// Code generated by %s %s; DO NOT EDIT.", strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), strings.Join(os.Args[1:], " ")),
		Package: fast.Name.Name,
		Imports: imports,
		Props:   props.Values(),
	}

	t := template.Must(template.New("code").Parse(tmpl))

	declFile := viper.GetString("decl_file")

	os.MkdirAll(filepath.Dir(declFile), os.ModePerm)

	file, err := os.OpenFile(filepath.Join(filepath.Dir(declFile), filepath.Base(strings.TrimSuffix(declFile, ".go"))+".sync.gen.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = t.Execute(file, args)
	if err != nil {
		panic(err)
	}

	props.Each(func(_ string, prop *Prop) {
		log.Printf("Prop: %s", prop.Name)
		for _, op := range prop.Ops {
			log.Printf("\t- Op: %s", op.Decl)
		}
	})
}

func loadDeclFile() (*token.FileSet, *ast.File, []byte) {
	declFile := viper.GetString("decl_file")

	fileData, err := ioutil.ReadFile(declFile)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()

	fast, err := parser.ParseFile(fset, declFile, fileData, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	return fset, fast, fileData
}

func parseGenAtti(str, atti string) url.Values {
	idx := strings.Index(str, atti)
	if idx < 0 {
		return url.Values{}
	}

	str = str[idx+len(atti):]

	end := strings.IndexAny(str, "\r\n")
	if end >= 0 {
		str = str[:end]
	}

	values, _ := url.ParseQuery(str)

	for k, vs := range values {
		for i, v := range vs {
			vs[i] = strings.TrimSpace(v)
		}
		values[k] = vs
	}

	return values
}

func getAtti(fset *token.FileSet, fast *ast.File, node ast.Node) url.Values {
	for _, comment := range fast.Comments {
		if fset.Position(comment.End()).Line+1 == fset.Position(node.Pos()).Line {
			return parseGenAtti(comment.Text(), "+prop-sync-gen:")
		}
	}
	return url.Values{}
}
