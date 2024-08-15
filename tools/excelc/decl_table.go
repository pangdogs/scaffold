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
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/xuri/excelize/v2"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	SheetTableHeader     = 1
	SheetTableHeaderSize = 4

	SheetTableColumnName    = 1
	SheetTableColumnType    = 2
	SheetTableColumnMeta    = 3
	SheetTableColumnComment = 4
)

func parseTableDecls(file *excelize.File, globalDecls *generic.SliceMap[Type, *Decl]) *generic.SliceMap[Type, *Decl] {
	var decls generic.SliceMap[Type, *Decl]

	sheets := slices.DeleteFunc(file.GetSheetList(), func(sheet string) bool {
		return sheet == "" || !unicode.IsLetter(rune(sheet[0]))
	})
	if len(sheets) <= 0 {
		return &decls
	}

	sheet := sheets[0]

	rows, err := file.Rows(sheet)
	if err != nil {
		panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，%s", file.Path, sheet, err))
	}
	defer rows.Close()

	type ColumnDesc struct {
		Name    string
		Type    string
		Meta    string
		Comment string
	}

	var tableDesc []*ColumnDesc

	for i := 1; rows.Next(); i++ {
		if i >= SheetTableHeader+SheetTableHeaderSize {
			break
		}

		row, err := rows.Columns()
		if err != nil {
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，%s", file.Path, SheetTypes, i, err))
		}

		for j, cell := range row {
			row[j] = strings.NewReplacer("\r", "", "\n", "\\n").Replace(strings.TrimSpace(cell))
		}

	loop:
		for j, cell := range row {
			switch i {
			case SheetTableColumnName:
				if tableDesc == nil {
					for k, c := range row {
						if c == "" {
							row = row[:k]
							break
						}
					}
					tableDesc = make([]*ColumnDesc, len(row))
				}

				if j < 0 || j >= len(tableDesc) {
					break loop
				}

				tableDesc[j] = &ColumnDesc{
					Name: snake2Camel(cell),
				}

			case SheetTableColumnType:
				if j < 0 || j >= len(tableDesc) {
					break loop
				}

				tableDesc[j].Type = cell

			case SheetTableColumnMeta:
				if j < 0 || j >= len(tableDesc) {
					break loop
				}

				tableDesc[j].Meta = cell

			case SheetTableColumnComment:
				if j < 0 || j >= len(tableDesc) {
					break loop
				}

				tableDesc[j].Comment = cell
			}
		}
	}

	tableDesc = slices.DeleteFunc(tableDesc, func(field *ColumnDesc) bool {
		return field == nil || field.Name == "" || !unicode.IsLetter(rune(field.Name[0]))
	})
	if len(tableDesc) <= 0 {
		return &decls
	}

	tableDecl := &Decl{
		File:    file.Path,
		Sheet:   sheet,
		Line:    SheetTableHeader,
		Type:    Type(snake2Camel(strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))) + "Columns"),
		IsTable: true,
	}
	decls.Add(tableDecl.Type, tableDecl)

	for _, columnDesc := range tableDesc {
		if columnDesc.Type == "" {
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 未配置类型", file.Path, sheet, columnDesc.Name))
		}

		columnType := Type(columnDesc.Type)

		mapped := columnType.IsMap()
		if mapped {
			k, v := columnType.KV()

			var kDecl, vDecl *Decl

			if !k.CanK() {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 类型 %q，错误的Key类型", file.Path, sheet, columnDesc.Name, columnType))
			}

			kDecl = &Decl{
				Type:      k,
				IsBuiltin: true,
			}

			if v.IsBuiltin() {
				vDecl = &Decl{
					Type:      v,
					IsBuiltin: true,
				}
			} else {
				var ok bool
				vDecl, ok = globalDecls.Get(v)
				if !ok {
					panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 类型 %q，未定义的Value类型", file.Path, sheet, columnDesc.Name, columnType))
				}
			}

			columnDecl := &Decl{
				Type:  columnType,
				IsMap: true,
				Mapping: &Mapping{
					K: kDecl,
					V: vDecl,
				},
			}

			meta, err := parseMeta(columnDesc.Meta)
			if err != nil {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 解析Meta %q 失败，%s", file.Path, SheetTypes, columnDesc.Name, columnDesc.Meta, err))
			}

			columnField := &Field{
				Decl:     columnDecl,
				IsColumn: true,
				Name:     columnDesc.Name,
				Meta:     meta,
				Comment:  columnDesc.Comment,
			}

			tableDecl.Fields.Add(columnField.Name, columnField)
			continue
		}

		repeated := columnType.IsRepeated()
		if repeated {
			columnType = columnType.Child()
		}

		var columnDecl *Decl

		if columnType.IsBuiltin() {
			columnDecl = &Decl{
				Type:      columnType,
				IsBuiltin: true,
			}
		} else {
			var ok bool
			columnDecl, ok = globalDecls.Get(columnType)
			if !ok {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 类型 %q 未定义", file.Path, sheet, columnDesc.Name, columnType))
			}
		}

		meta, err := parseMeta(columnDesc.Meta)
		if err != nil {
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，列 %q 解析Meta %q 失败，%s", file.Path, SheetTypes, columnDesc.Name, columnDesc.Meta, err))
		}

		columnField := &Field{
			Decl: columnDecl,
			Meta: defaultMeta,
		}

		if repeated {
			parent := &Field{
				Decl: &Decl{
					Type:       columnType.Repeated(),
					IsRepeated: true,
					Child:      columnField,
				},
				IsColumn: true,
				Name:     columnDesc.Name,
				Meta:     meta,
				Comment:  columnDesc.Comment,
			}

			tableDecl.Fields.Add(parent.Name, parent)

		} else {
			columnField.IsColumn = true
			columnField.Name = columnDesc.Name
			columnField.Meta = meta
			columnField.Comment = columnDesc.Comment

			tableDecl.Fields.Add(columnField.Name, columnField)
		}
	}

	return &decls
}
