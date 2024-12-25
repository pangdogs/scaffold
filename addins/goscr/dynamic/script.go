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
	"go/ast"
	"golang.org/x/tools/go/packages"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"
)

// This This类型
type This struct {
	Name    string // 类型名称
	PkgName string // 包名
	PkgPath string // 包路径
}

// UniquePkgName 获取唯一包名
func (this *This) UniquePkgName() string {
	return strings.ReplaceAll(strings.ReplaceAll(this.PkgPath, "/", "_"), ".", "_")
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
	This         *This        // This类型，类型标识不为空时有效，表示扩展的基类型，为nil表示没有基类型
	Methods      []*Method    // 方法列表
	MethodBinder MethodBinder // 成员方法绑定器，类型标识不为空并且This类型不为nil时有效
}

// UniquePkgName 获取唯一包名
func (s *Script) UniquePkgName() string {
	return strings.ReplaceAll(strings.ReplaceAll(s.PkgPath, "/", "_"), ".", "_")
}

// Scripts 脚本集合
type Scripts map[string]*Script

// Ident 标识
func (scripts Scripts) Ident(ident string) *Script {
	if scripts == nil {
		return nil
	}
	return scripts[ident]
}

// Range 遍历
func (scripts Scripts) Range(fun generic.Func1[*Script, bool]) {
	for _, script := range scripts {
		if !fun.Exec(script) {
			return
		}
	}
}

// NewScriptLib 创建脚本库
func NewScriptLib() ScriptLib {
	return ScriptLib{}
}

// ScriptLib 脚本库
type ScriptLib map[string]Scripts

// Package 包
func (lib ScriptLib) Package(pkgPath string) Scripts {
	return lib[pkgPath]
}

// Range 遍历
func (lib ScriptLib) Range(fun generic.Func2[string, Scripts, bool]) {
	for pkgPath, scripts := range lib {
		if !fun.Exec(pkgPath, scripts) {
			return
		}
	}
}

// PushIdent 添加类型标识
func (lib ScriptLib) PushIdent(pkgPath, ident string, this *This) bool {
	scripts, ok := lib[pkgPath]
	if !ok {
		scripts = Scripts{}
		lib[pkgPath] = scripts
	}

	if _, ok := scripts[ident]; ok {
		return false
	}

	scripts[ident] = &Script{
		PkgName: path.Base(pkgPath),
		PkgPath: pkgPath,
		Ident:   ident,
		This:    this,
	}

	return true
}

// PushMethod 添加方法
func (lib ScriptLib) PushMethod(pkgPath, ident string, method string) bool {
	scripts, ok := lib[pkgPath]
	if !ok {
		scripts = Scripts{}
		lib[pkgPath] = scripts
	}

	s, ok := scripts[ident]
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
func (lib ScriptLib) Load(localPath string) error {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}

	pkgs, err := packages.Load(cfg, filepath.Join(localPath, "..."))
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if lib.Package(pkg.PkgPath) != nil {
			return fmt.Errorf("package %q conflicted", pkg.PkgPath)
		}
	}

	for _, pkg := range pkgs {
		lib.PushIdent(pkg.PkgPath, "", nil)

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok {
					return true
				}

				for _, spec := range genDecl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}

					firstField := pie.First(structType.Fields.List)
					if firstField == nil {
						lib.PushIdent(pkg.PkgPath, ts.Name.Name, nil)
						continue
					}

					firstFieldType, ok := firstField.Type.(*ast.StarExpr)
					if !ok {
						continue
					}

					firstFieldSelector, ok := firstFieldType.X.(*ast.SelectorExpr)
					if !ok || firstFieldSelector.Sel == nil {
						continue
					}

					firstFieldPkgIdent, ok := firstFieldSelector.X.(*ast.Ident)
					if !ok {
						continue
					}

					idx := slices.IndexFunc(file.Imports, func(spec *ast.ImportSpec) bool {
						if spec.Path == nil {
							return false
						}
						if spec.Name != nil {
							return spec.Name.Name == firstFieldPkgIdent.Name
						}
						return path.Base(strings.Trim(spec.Path.Value, `"`)) == firstFieldPkgIdent.Name
					})
					if idx < 0 {
						continue
					}
					imp := file.Imports[idx]

					lib.PushIdent(pkg.PkgPath, ts.Name.Name, &This{
						Name:    firstFieldSelector.Sel.Name,
						PkgName: firstFieldPkgIdent.Name,
						PkgPath: strings.Trim(imp.Path.Value, `"`),
					})
				}

				return true
			})
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				funcDecl, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				if funcDecl.Recv == nil {
					lib.PushMethod(pkg.PkgPath, "", funcDecl.Name.Name)
					return true
				}

				for _, field := range funcDecl.Recv.List {
					if starExpr, ok := field.Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok {
							lib.PushMethod(pkg.PkgPath, ident.Name, funcDecl.Name.Name)
						}
					} else if ident, ok := field.Type.(*ast.Ident); ok {
						lib.PushMethod(pkg.PkgPath, ident.Name, funcDecl.Name.Name)
					}
				}

				return true
			})
		}
	}

	return nil
}

// Compile 编译
func (lib ScriptLib) Compile(i *interp.Interpreter) error {
	buff := &bytes.Buffer{}

	for pkgPath, scripts := range lib {
		if _, err := i.EvalPath(pkgPath); err != nil {
			return err
		}

		for _, s := range scripts {
			if s.Ident != "" && s.This != nil {
				code := `
package {{.UniquePkgName}}_export

import (
	{{.UniquePkgName}} "{{.PkgPath}}"
	{{.This.UniquePkgName}} "{{.This.PkgPath}}"
)

func Bind_{{.Ident}}(this any, method string) any {
	switch method {
	{{range .Methods}}
	case "{{.Name}}":
		return {{$.UniquePkgName}}.{{$.Ident}}{this.(*{{$.This.UniquePkgName}}.{{$.This.Name}})}.{{.Name}}
	{{end}}
	}
	return nil
}
`
				tmpl, err := template.New("").Parse(code)
				if err != nil {
					return err
				}

				buff.Reset()
				if err := tmpl.Execute(buff, s); err != nil {
					return err
				}

				if _, err := i.Eval(buff.String()); err != nil {
					return err
				}

				binderRV, err := i.Eval(fmt.Sprintf(`%s_export.Bind_%s`, s.UniquePkgName(), s.Ident))
				if err != nil {
					return err
				}

				binder, ok := binderRV.Interface().(MethodBinder)
				if !ok {
					return fmt.Errorf("package %q ident %q incorrect method binder type", pkgPath, s.Ident)
				}

				s.MethodBinder = binder
			}

			if _, err := i.Eval(fmt.Sprintf(`import %s "%s"`, s.UniquePkgName(), s.PkgPath)); err != nil {
				return err
			}

			for _, method := range s.Methods {
				if s.Ident != "" {
					methodRV, err := i.Eval(fmt.Sprintf(`%s.%s.%s`, s.UniquePkgName(), s.Ident, method.Name))
					if err != nil {
						return err
					}
					method.Reflected = methodRV
				} else {
					methodRV, err := i.Eval(fmt.Sprintf(`%s.%s`, s.UniquePkgName(), method.Name))
					if err != nil {
						return err
					}
					method.Reflected = methodRV
				}
			}
		}
	}

	return nil
}
