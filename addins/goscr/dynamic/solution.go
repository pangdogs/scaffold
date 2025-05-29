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
	"cmp"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/pangdogs/yaegi/interp"
	"github.com/spf13/afero"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
)

// Project 项目
type Project struct {
	ScriptRoot string           // 脚本根路径
	LocalPath  string           // 本地路径
	SymbolsTab []interp.Exports // 符号表
}

// NewSolution 创建解决方案
func NewSolution(pkgRoot string) *Solution {
	fs := NewCodeFs("src/main/vendor/")
	afero.WriteFile(fs.AferoFs(), path.Join(pkgRoot, "go.mod"), []byte(fmt.Sprintf("module %s", pkgRoot)), os.ModePerm)

	i := interp.New(interp.Options{
		SourcecodeFilesystem: fs,
		Unrestricted:         true,
	})

	return &Solution{
		pkgRoot: pkgRoot,
		fs:      fs,
		interp:  i,
		lib:     NewScriptLib(),
	}
}

// Solution 解决方案
type Solution struct {
	pkgRoot string
	fs      *CodeFs
	interp  *interp.Interpreter
	lib     ScriptLib
}

// Use 导入符号表
func (s *Solution) Use(symbols interp.Exports) error {
	return s.interp.Use(symbols)
}

// Eval 执行代码
func (s *Solution) Eval(code string) (reflect.Value, error) {
	return s.interp.Eval(code)
}

// Package 包
func (s *Solution) Package(pkgPath string) ScriptBundle {
	return s.lib.Package(pkgPath)
}

// Range 遍历
func (s *Solution) Range(fun generic.Func2[string, ScriptBundle, bool]) {
	s.lib.Range(fun)
}

// Load 加载项目
func (s *Solution) Load(project *Project) error {
	err := filepath.Walk(project.LocalPath, func(filePath string, fileInfo fs.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() {
			return nil
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		relFilePath, err := filepath.Rel(project.LocalPath, filePath)
		if err != nil {
			return err
		}

		return afero.WriteFile(s.fs.AferoFs(), path.Join(s.pkgRoot, project.ScriptRoot, relFilePath), fileData, os.ModePerm)
	})
	if err != nil {
		return err
	}

	for _, symbols := range project.SymbolsTab {
		if err := s.interp.Use(symbols); err != nil {
			return err
		}
	}

	if err := s.lib.Load(project.LocalPath); err != nil {
		return err
	}

	if err := s.lib.Compile(s.interp); err != nil {
		return err
	}

	return nil
}

// Method 方法
func (s *Solution) Method(pkgPath, method string) reflect.Value {
	script := s.lib.Package(pkgPath).Ident("")
	if script == nil {
		return reflect.Value{}
	}

	idx, ok := slices.BinarySearchFunc(script.Methods, method, func(method *Method, target string) int {
		return cmp.Compare(method.Name, target)
	})
	if !ok {
		return reflect.Value{}
	}

	return script.Methods[idx].Reflected
}

// BindMethod 绑定成员方法
func (s *Solution) BindMethod(this reflect.Value, pkgPath, ident string, method string) any {
	script := s.lib.Package(pkgPath).Ident(ident)
	if script == nil {
		return nil
	}

	if script.MethodBinder == nil {
		return nil
	}

	switch script.BindMode {
	case Func:
		getThis := this.MethodByName("This")
		if !getThis.IsValid() {
			return nil
		}
		ret := getThis.Call(nil)
		if len(ret) < 1 {
			return nil
		}
		this = ret[0]
	case Struct:
		break
	default:
		return nil
	}

	ret := script.MethodBinder(this.Interface(), method)
	if ret == nil {
		return nil
	}

	return ret
}
