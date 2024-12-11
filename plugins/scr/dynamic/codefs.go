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
	"git.golaxy.org/core/utils/generic"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// NewCodeFS 创建代码文件系统
func NewCodeFS() *CodeFS {
	return &CodeFS{}
}

// CodeFS 代码文件系统
type CodeFS struct {
	mappingPath generic.SliceMap[string, string]
}

// Mapping 映射包路径
func (cfs *CodeFS) Mapping(pkgRoot, localPath string) error {
	pkgRoot = path.Clean(filepath.ToSlash(pkgRoot))

	localPath, err := filepath.Abs(filepath.FromSlash(localPath))
	if err != nil {
		return err
	}
	localPath = filepath.Clean(localPath)

	if !cfs.mappingPath.TryAdd(pkgRoot, localPath) {
		return os.ErrExist
	}

	return nil
}

// Open implements fs.FS
func (cfs *CodeFS) Open(name string) (file fs.File, err error) {
	name = path.Clean(filepath.ToSlash(name))

	cfs.mappingPath.ReversedRange(func(pkgRoot string, localPath string) bool {
		if !strings.HasPrefix(name, pkgRoot) {
			return true
		}

		localName := strings.TrimPrefix(name, pkgRoot)
		if localName != "" {
			if localName[0] != '/' {
				return true
			}
			localName = filepath.FromSlash(localName)
		}
		realName := filepath.Join(localPath, localName)

		file, err = os.OpenFile(realName, os.O_RDONLY, 0)
		if err != nil {
			return true
		}

		return false
	})

	if file == nil && err == nil {
		return nil, os.ErrNotExist
	}

	return
}

// ReadFile implements fs.ReadFileFS
func (cfs *CodeFS) ReadFile(name string) ([]byte, error) {
	file, err := cfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// Stat implements fs.StatFS
func (cfs *CodeFS) Stat(name string) (fs.FileInfo, error) {
	file, err := cfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file.Stat()
}

// ReadDir implements fs.ReadDirFS
func (cfs *CodeFS) ReadDir(name string) (files []fs.DirEntry, err error) {
	name = path.Clean(filepath.ToSlash(name))

	cfs.mappingPath.ReversedRange(func(pkgRoot string, localPath string) bool {
		if !strings.HasPrefix(name, pkgRoot) {
			return true
		}

		localName := strings.TrimPrefix(name, pkgRoot)
		if localName != "" {
			if localName[0] != '/' {
				return true
			}
			localName = filepath.FromSlash(localName)
		}
		realName := filepath.Join(localPath, localName)

		files, err = os.ReadDir(realName)
		if err != nil {
			return true
		}

		return false
	})

	if files == nil && err == nil {
		return nil, os.ErrNotExist
	}

	return
}
