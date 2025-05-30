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
	"archive/tar"
	"archive/zip"
	"bytes"
	"cmp"
	"compress/gzip"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/go-resty/resty/v2"
	"github.com/pangdogs/yaegi/interp"
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

// Project 项目
type Project struct {
	ScriptRoot string           // 脚本根路径
	LocalPath  string           // 本地路径
	RemoteURL  string           // 远程下载URL，支持打包格式：tar.gz、zip
	SymbolsTab []interp.Exports // 符号表
}

// NewSolution 创建解决方案
func NewSolution(pkgRoot string) *Solution {
	fs := NewCodeFs("src/main/vendor/")

	i := interp.New(interp.Options{
		SourcecodeFilesystem: fs,
		Unrestricted:         true,
	})

	return &Solution{
		pkgRoot:   pkgRoot,
		codeFs:    fs,
		interp:    i,
		scriptLib: NewScriptLib(),
	}
}

// Solution 解决方案
type Solution struct {
	pkgRoot   string
	codeFs    *CodeFs
	interp    *interp.Interpreter
	scriptLib ScriptLib
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
	return s.scriptLib.Package(pkgPath)
}

// Range 遍历
func (s *Solution) Range(fun generic.Func2[string, ScriptBundle, bool]) {
	s.scriptLib.Range(fun)
}

// Load 加载项目
func (s *Solution) Load(project *Project) error {
	scriptPath := path.Join(s.pkgRoot, project.ScriptRoot)

	b, err := afero.Exists(s.codeFs.AferoFs(), scriptPath)
	if err != nil {
		return fmt.Errorf("invalid script path %q, %s", scriptPath, err)
	}
	if b {
		return fmt.Errorf("script path %q conflicted", scriptPath)
	}

	if project.LocalPath != "" {
		err = filepath.Walk(project.LocalPath, func(filePath string, fileInfo fs.FileInfo, err error) error {
			if err != nil || fileInfo.IsDir() {
				return nil
			}

			fileData, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read local file %q failed, %s", filePath, err)
			}

			scriptFilePath, err := filepath.Rel(project.LocalPath, filePath)
			if err != nil {
				return fmt.Errorf("relative local file %q failed, %s", filePath, err)
			}
			scriptFilePath = path.Join(scriptPath, scriptFilePath)

			err = afero.WriteFile(s.codeFs.AferoFs(), scriptFilePath, fileData, os.ModePerm)
			if err != nil {
				return fmt.Errorf("write script file %q failed, %s", scriptFilePath, err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	if project.RemoteURL != "" {
		resp, err := resty.New().
			SetDoNotParseResponse(true).
			R().
			Get(project.RemoteURL)
		if err != nil {
			return fmt.Errorf("download remote file %q failed, %s", project.RemoteURL, err)
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("download remote file %q failed, status code %d", project.RemoteURL, resp.StatusCode())
		}

		switch strings.ToLower(path.Ext(project.RemoteURL)) {
		case ".tar.gz":
			if err := s.extractTarGzip(resp.RawBody()); err != nil {
				return fmt.Errorf("extract remote file %q failed, %s", project.RemoteURL, err)
			}
		case ".zip":
			if err := s.extractZip(resp.RawBody()); err != nil {
				return fmt.Errorf("extract remote file %q failed, %s", project.RemoteURL, err)
			}
		default:
			return fmt.Errorf("unsupported remote file %q", project.RemoteURL)
		}
	}

	for _, symbols := range project.SymbolsTab {
		if err := s.interp.Use(symbols); err != nil {
			return fmt.Errorf("script path %q use symbols failed, %s", scriptPath, err)
		}
	}

	if err := s.scriptLib.Load(s.codeFs, scriptPath); err != nil {
		return fmt.Errorf("load script path %q failed, %s", scriptPath, err)
	}

	if err := s.scriptLib.Compile(s.interp, scriptPath); err != nil {
		return fmt.Errorf("compile script path %q failed, %s", scriptPath, err)
	}

	return nil
}

// Method 方法
func (s *Solution) Method(pkgPath, method string) reflect.Value {
	script := s.scriptLib.Package(pkgPath).Ident("")
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
	script := s.scriptLib.Package(pkgPath).Ident(ident)
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

func (s *Solution) extractTarGzip(reader io.ReadCloser) error {
	defer reader.Close()

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := s.codeFs.AferoFs().MkdirAll(header.Name, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := s.codeFs.AferoFs().MkdirAll(filepath.Dir(header.Name), os.ModePerm); err != nil {
				return err
			}

			err := func() error {
				dstFile, err := s.codeFs.AferoFs().Create(header.Name)
				if err != nil {
					return err
				}
				defer dstFile.Close()

				if _, err := io.Copy(dstFile, tarReader); err != nil {
					return err
				}

				return nil
			}()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Solution) extractZip(reader io.ReadCloser) error {
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, zipFile := range zipReader.File {
		filePath := filepath.Clean(zipFile.Name)
		if strings.Contains(filePath, "..") {
			continue
		}

		if zipFile.FileInfo().IsDir() {
			if err := s.codeFs.AferoFs().MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := s.codeFs.AferoFs().MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		err := func() error {
			zipFileReader, err := zipFile.Open()
			if err != nil {
				return err
			}
			defer zipFileReader.Close()

			dstFile, err := s.codeFs.AferoFs().OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, zipFileReader); err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
