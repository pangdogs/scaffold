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
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// NewCodeFS 创建代码文件系统
func NewCodeFS(rootPath string) *CodeFS {
	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}
	return &CodeFS{
		rootPath: rootPath,
		fakeFs:   afero.NewMemMapFs(),
	}
}

type _FakeDirEntry struct {
	os.FileInfo
}

func (e *_FakeDirEntry) Type() os.FileMode {
	return e.Mode().Type()
}

func (e *_FakeDirEntry) Info() (os.FileInfo, error) {
	return e.FileInfo, nil
}

// CodeFS 代码文件系统
type CodeFS struct {
	rootPath    string
	mappingPath generic.SliceMap[string, string]
	fakeFs      afero.Fs
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

// Unmapping 移除包路径映射
func (cfs *CodeFS) Unmapping(pkgRoot string) {
	pkgRoot = path.Clean(filepath.ToSlash(pkgRoot))
	cfs.mappingPath.Delete(pkgRoot)
}

// AddFakeFile 添加虚拟文件
func (cfs *CodeFS) AddFakeFile(name string, data []byte) error {
	name = path.Clean(filepath.ToSlash(name))

	if b, err := afero.Exists(cfs.fakeFs, name); err != nil {
		return err
	} else if b {
		return os.ErrExist
	}

	return afero.WriteFile(cfs.fakeFs, name, data, os.ModePerm)
}

// RemoveFakeFile 移除虚拟文件
func (cfs *CodeFS) RemoveFakeFile(name string) {
	name = path.Clean(filepath.ToSlash(name))
	cfs.fakeFs.Remove(name)
}

// Open implements fs.FS
func (cfs *CodeFS) Open(name string) (file fs.File, err error) {
	name = path.Clean(filepath.ToSlash(name))

	if strings.HasPrefix(name, cfs.rootPath) {
		name = strings.TrimPrefix(name, cfs.rootPath)
	}

	if file, err = cfs.fakeFs.Open(name); err == nil {
		return
	}

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
func (cfs *CodeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = path.Clean(filepath.ToSlash(name))

	if strings.HasPrefix(name, cfs.rootPath) {
		name = strings.TrimPrefix(name, cfs.rootPath)
	}

	var files []fs.DirEntry
	var err error

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

	fakeFiles, _ := afero.ReadDir(cfs.fakeFs, name)
	if len(fakeFiles) > 0 {
		for _, fakeFile := range fakeFiles {
			idx := slices.IndexFunc(files, func(file fs.DirEntry) bool {
				return file.Name() == fakeFile.Name()
			})
			if idx >= 0 {
				files[idx] = &_FakeDirEntry{FileInfo: fakeFile}
				continue
			}
			files = append(files, &_FakeDirEntry{FileInfo: fakeFile})
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})
	}

	if len(files) <= 0 {
		return nil, os.ErrNotExist
	}

	return files, nil
}
