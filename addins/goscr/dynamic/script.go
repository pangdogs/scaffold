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

package dynamic

import (
	"bytes"
	"cmp"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
	"github.com/pangdogs/yaegi/interp"
	"github.com/spf13/afero"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"
)

// BindMode 绑定模式
type BindMode int32

const (
	None   BindMode = iota // 无绑定
	Func                   // 函数
	Struct                 // 结构体
)

// This This类型
type This struct {
	Name    string // 类型名称
	PkgName string // 包名
	PkgPath string // 包路径
}

// UniquePkgName 获取唯一包名
func (this *This) UniquePkgName() string {
	return strings.NewReplacer("/", "_", ".", "_").Replace(this.PkgPath)
}

// Method 方法
type Method struct {
	Name      string        // 方法名
	Reflected reflect.Value // 方法反射值
}

// MethodBinder 成员方法绑定器
type MethodBinder = func(this any, method string) any

// Script 脚本
type Script struct {
	PkgName      string       // 包名
	PkgPath      string       // 包路径
	Ident        string       // 类型标识，为空表示脚本均为全局方法
	BindMode     BindMode     // 绑定模式
	This         *This        // This类型
	Methods      []*Method    // 方法列表
	MethodBinder MethodBinder // 成员方法绑定器
}

// UniquePkgName 获取唯一包名
func (s *Script) UniquePkgName() string {
	return strings.NewReplacer("/", "_", ".", "_").Replace(s.PkgPath)
}

// ScriptBundle 脚本集合
type ScriptBundle map[string]*Script

// Ident 标识
func (scriptBundle ScriptBundle) Ident(ident string) *Script {
	if scriptBundle == nil {
		return nil
	}
	return scriptBundle[ident]
}

// Range 遍历
func (scriptBundle ScriptBundle) Range(fun generic.Func1[*Script, bool]) {
	for _, script := range scriptBundle {
		if !fun.UnsafeCall(script) {
			return
		}
	}
}

// NewScriptLib 创建脚本库
func NewScriptLib() ScriptLib {
	return ScriptLib{}
}

// ScriptLib 脚本库
type ScriptLib map[string]ScriptBundle

// Package 包
func (lib ScriptLib) Package(pkgPath string) ScriptBundle {
	return lib[pkgPath]
}

// Range 遍历
func (lib ScriptLib) Range(fun generic.Func2[string, ScriptBundle, bool]) {
	for pkgPath, scriptBundle := range lib {
		if !fun.UnsafeCall(pkgPath, scriptBundle) {
			return
		}
	}
}

// PushIdent 添加类型标识
func (lib ScriptLib) PushIdent(pkgPath, ident string, bindMode BindMode, this *This) bool {
	scriptBundle, ok := lib[pkgPath]
	if !ok {
		scriptBundle = ScriptBundle{}
		lib[pkgPath] = scriptBundle
	}

	if _, ok := scriptBundle[ident]; ok {
		return false
	}

	scriptBundle[ident] = &Script{
		PkgName:  path.Base(pkgPath),
		PkgPath:  pkgPath,
		Ident:    ident,
		BindMode: bindMode,
		This:     this,
	}

	return true
}

// PushMethod 添加方法
func (lib ScriptLib) PushMethod(pkgPath, ident string, method string) bool {
	scriptBundle, ok := lib[pkgPath]
	if !ok {
		scriptBundle = ScriptBundle{}
		lib[pkgPath] = scriptBundle
	}

	s, ok := scriptBundle[ident]
	if !ok {
		return false
	}

	if pie.Any(s.Methods, func(exists *Method) bool {
		return exists.Name == method
	}) {
		return false
	}

	s.Methods = append(s.Methods, &Method{
		Name: method,
	})

	slices.SortFunc(s.Methods, func(a, b *Method) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return true
}

// Load 加载
func (lib ScriptLib) Load(codeFs *CodeFs, scriptPath string) error {
	scriptPath = path.Clean(scriptPath)
	fset := token.NewFileSet()

	type _Code struct {
		PkgPath string
		File    *ast.File
	}

	var codes []*_Code

	err := afero.Walk(codeFs.AferoFs(), ".", func(filePath string, fileInfo fs.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() || filepath.Ext(fileInfo.Name()) != ".go" {
			return nil
		}

		pkgPath := path.Dir(filepath.ToSlash(filePath))
		if !strings.HasPrefix(pkgPath, scriptPath) {
			return nil
		}

		fileData, err := afero.ReadFile(codeFs.AferoFs(), filePath)
		if err != nil {
			return fmt.Errorf("read script file %q failed, %s", filePath, err)
		}

		file, err := parser.ParseFile(fset, filePath, fileData, parser.AllErrors|parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse script file %q failed, %s", filePath, err)
		}

		codes = append(codes, &_Code{PkgPath: pkgPath, File: file})

		lib.PushIdent(pkgPath, "", None, nil)

		return nil
	})
	if err != nil {
		return err
	}

	for _, code := range codes {
		ast.Inspect(code.File, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok {
				return true
			}

			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				var thisField *ast.Field
				var bindMode BindMode

				switch ty := ts.Type.(type) {
				case *ast.StructType:
					if ty.Fields == nil {
						continue
					}

					thisField = pie.First(ty.Fields.List)
					if thisField == nil {
						continue
					}

					bindMode = Struct

				case *ast.FuncType:
					if ty.TypeParams != nil && len(ty.TypeParams.List) > 0 {
						continue
					}

					if ty.Params != nil && len(ty.Params.List) > 0 {
						continue
					}

					if ty.Results == nil {
						continue
					}

					thisField = pie.First(ty.Results.List)
					if thisField == nil {
						continue
					}

					bindMode = Func

				default:
					continue
				}

				thisFieldType, ok := thisField.Type.(*ast.StarExpr)
				if !ok {
					continue
				}

				thisFieldSelector, ok := thisFieldType.X.(*ast.SelectorExpr)
				if !ok || thisFieldSelector.Sel == nil {
					continue
				}

				thisFieldPkgIdent, ok := thisFieldSelector.X.(*ast.Ident)
				if !ok {
					continue
				}

				idx := slices.IndexFunc(code.File.Imports, func(spec *ast.ImportSpec) bool {
					if spec.Path == nil {
						return false
					}
					if spec.Name != nil {
						return spec.Name.Name == thisFieldPkgIdent.Name
					}
					return path.Base(strings.Trim(spec.Path.Value, `"`)) == thisFieldPkgIdent.Name
				})
				if idx < 0 {
					continue
				}
				imp := code.File.Imports[idx]

				lib.PushIdent(code.PkgPath, ts.Name.Name, bindMode, &This{
					Name:    thisFieldSelector.Sel.Name,
					PkgName: thisFieldPkgIdent.Name,
					PkgPath: strings.Trim(imp.Path.Value, `"`),
				})
			}

			return true
		})
	}

	for _, code := range codes {
		ast.Inspect(code.File, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Recv == nil {
				lib.PushMethod(code.PkgPath, "", funcDecl.Name.Name)
				return true
			}

			for _, field := range funcDecl.Recv.List {
				if starExpr, ok := field.Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						lib.PushMethod(code.PkgPath, ident.Name, funcDecl.Name.Name)
					}
				} else if ident, ok := field.Type.(*ast.Ident); ok {
					lib.PushMethod(code.PkgPath, ident.Name, funcDecl.Name.Name)
				}
			}

			return true
		})
	}

	return nil
}

// Compile 编译
func (lib ScriptLib) Compile(i *interp.Interpreter, scriptPath string) error {
	scriptPath = path.Clean(scriptPath)
	buff := &bytes.Buffer{}

	for pkgPath, scriptBundle := range lib {
		if !strings.HasPrefix(pkgPath, scriptPath) {
			continue
		}

		if _, err := i.EvalPath(pkgPath); err != nil {
			return fmt.Errorf("eval script path %q failed, %s", pkgPath, err)
		}

		var no int

		for _, script := range scriptBundle {
			if script.BindMode != None {
				var code string

				switch script.BindMode {
				case Func:
					code = `
package {{.UniquePkgName}}_export

import (
	{{.UniquePkgName}}_{{.No}} "{{.PkgPath}}"
	{{.This.UniquePkgName}}_{{.No}} "{{.This.PkgPath}}"
)

func Bind_{{.Ident}}(this any, method string) any {
	switch method {
	{{range .Methods}}
	case "{{.Name}}":
		return {{$.UniquePkgName}}_{{$.No}}.{{$.Ident}}(this.(func() *{{$.This.UniquePkgName}}_{{$.No}}.{{$.This.Name}})).{{.Name}}
	{{end}}
	}
	return nil
}
`
				case Struct:
					code = `
package {{.UniquePkgName}}_export

import (
	{{.UniquePkgName}}_{{.No}} "{{.PkgPath}}"
	{{.This.UniquePkgName}}_{{.No}} "{{.This.PkgPath}}"
)

func Bind_{{.Ident}}(this any, method string) any {
	switch method {
	{{range .Methods}}
	case "{{.Name}}":
		return {{$.UniquePkgName}}_{{$.No}}.{{$.Ident}}{this.(*{{$.This.UniquePkgName}}_{{$.No}}.{{$.This.Name}})}.{{.Name}}
	{{end}}
	}
	return nil
}
`
				default:
					continue
				}

				tmpl, err := template.New("").Parse(code)
				if err != nil {
					return fmt.Errorf("script path %q new template failed, %s", pkgPath, err)
				}

				type _Args struct {
					*Script
					No int
				}

				args := _Args{
					Script: script,
					No:     no,
				}

				buff.Reset()
				if err := tmpl.Execute(buff, args); err != nil {
					return fmt.Errorf("script path %q execute template failed, %s", pkgPath, err)
				}

				if _, err := i.Eval(buff.String()); err != nil {
					return fmt.Errorf("script path %q eval export code failed, %s", pkgPath, err)
				}

				binderRV, err := i.Eval(fmt.Sprintf(`%s_export.Bind_%s`, script.UniquePkgName(), script.Ident))
				if err != nil {
					return fmt.Errorf("script path %q export ident %q failed, %s", pkgPath, script.Ident, err)
				}

				binder, ok := binderRV.Interface().(MethodBinder)
				if !ok {
					return fmt.Errorf("script path %q ident %q has incorrect method binder type", pkgPath, script.Ident)
				}

				script.MethodBinder = binder

				no++
			}

			if _, err := i.Eval(fmt.Sprintf(`import %s "%s"`, script.UniquePkgName(), script.PkgPath)); err != nil {
				return fmt.Errorf("import script path %q failed, %s", pkgPath, err)
			}

			for _, method := range script.Methods {
				if script.Ident != "" {
					methodRV, err := i.Eval(fmt.Sprintf(`%s.%s.%s`, script.UniquePkgName(), script.Ident, method.Name))
					if err != nil {
						return fmt.Errorf("script path %q export ident %q method %q failed, %s", pkgPath, script.Ident, method.Name, err)
					}
					method.Reflected = methodRV
				} else {
					methodRV, err := i.Eval(fmt.Sprintf(`%s.%s`, script.UniquePkgName(), method.Name))
					if err != nil {
						return fmt.Errorf("script path %q export method %q failed, %s", pkgPath, method.Name, err)
					}
					method.Reflected = methodRV
				}
			}
		}
	}

	return nil
}
