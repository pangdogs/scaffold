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
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// NewCodeFs 创建代码文件系统
func NewCodeFs(fakePaths ...string) *CodeFs {
	cfs := &CodeFs{
		aferoFs: afero.NewMemMapFs(),
	}

	for _, fakePath := range fakePaths {
		if fakePath == "" {
			continue
		}
		if !strings.HasSuffix(fakePath, "/") {
			fakePath += "/"
		}
		cfs.fakePaths = append(cfs.fakePaths, fakePath)
	}

	return cfs
}

type _CodeDirEntry struct {
	fs.FileInfo
}

func (e *_CodeDirEntry) Type() fs.FileMode {
	return e.FileInfo.Mode()
}

func (e *_CodeDirEntry) Info() (fs.FileInfo, error) {
	return e.FileInfo, nil
}

type _CodeFsDirFile struct {
	afero.File
}

// ReadDir implements fs.ReadDirFile
func (d *_CodeFsDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	fileInfos, err := d.File.Readdir(n)
	if err != nil {
		return nil, err
	}

	fileEntries := make([]fs.DirEntry, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		fileEntries = append(fileEntries, &_CodeDirEntry{FileInfo: fileInfo})
	}

	return fileEntries, nil
}

// CodeFs 代码文件系统
type CodeFs struct {
	fakePaths []string
	aferoFs   afero.Fs
}

// AferoFs 获取Afero文件系统
func (cfs *CodeFs) AferoFs() afero.Fs {
	return cfs.aferoFs
}

// Open implements fs.FS
func (cfs *CodeFs) Open(name string) (fs.File, error) {
	file, err := cfs.aferoFs.Open(cfs.normalizedPath(name))
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return &_CodeFsDirFile{File: file}, nil
	}

	return file, nil
}

// ReadFile implements fs.ReadFileFS
func (cfs *CodeFs) ReadFile(name string) ([]byte, error) {
	file, err := cfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// Stat implements fs.StatFS
func (cfs *CodeFs) Stat(name string) (fs.FileInfo, error) {
	file, err := cfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file.Stat()
}

// ReadDir implements fs.ReadDirFS
func (cfs *CodeFs) ReadDir(name string) ([]fs.DirEntry, error) {
	name = cfs.normalizedPath(name)

	fileInfos, err := afero.ReadDir(cfs.aferoFs, name)
	if err != nil {
		return nil, err
	}

	fileEntries := make([]fs.DirEntry, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		fileEntries = append(fileEntries, &_CodeDirEntry{FileInfo: fileInfo})
	}

	return fileEntries, nil
}

func (cfs *CodeFs) normalizedPath(name string) string {
	name = path.Clean(filepath.ToSlash(name))

	for _, fakePath := range cfs.fakePaths {
		if fakePath == "" {
			continue
		}
		if strings.HasPrefix(name, fakePath) {
			return strings.TrimPrefix(name, fakePath)
		}
		if name == fakePath[:len(fakePath)-1] {
			return "/"
		}
	}

	return name
}
