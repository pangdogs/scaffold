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
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
	"github.com/go-playground/form/v4"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

const (
	Double   Type = "double"
	Float    Type = "float"
	Int32    Type = "int32"
	Int64    Type = "int64"
	Uint32   Type = "uint32"
	Uint64   Type = "uint64"
	Sint32   Type = "sint32"
	Sint64   Type = "sint64"
	Fixed32  Type = "fixed32"
	Fixed64  Type = "fixed64"
	Sfixed32 Type = "sfixed32"
	Sfixed64 Type = "sfixed64"
	Bool     Type = "bool"
	String   Type = "string"
	Bytes    Type = "bytes"
)

type Type string

func (ty Type) IsBuiltin() bool {
	switch ty {
	case Double, Float, Int32, Int64, Uint32, Uint64, Sint32, Sint64, Fixed32, Fixed64, Sfixed32, Sfixed64, Bool, String, Bytes:
		return true
	default:
		return false
	}
}

func (ty Type) IsRepeated() bool {
	return strings.HasSuffix(string(ty), "[]") || strings.HasPrefix(string(ty), "[]") || strings.HasPrefix(string(ty), "repeated ")
}

func (ty Type) IsMap() bool {
	return strings.HasPrefix(string(ty), "map<") && strings.HasSuffix(string(ty), ">")
}

func (ty Type) CanK() bool {
	switch ty {
	case Int32, Int64, Uint32, Uint64, Sint32, Sint64, Fixed32, Fixed64, Sfixed32, Sfixed64, Bool, String:
		return true
	default:
		return false
	}
}

func (ty Type) Child() Type {
	if ty.IsRepeated() {
		if strings.HasPrefix(string(ty), "[]") {
			return Type(strings.TrimSpace(strings.TrimPrefix(string(ty), "[]")))
		}
		if strings.HasSuffix(string(ty), "[]") {
			return Type(strings.TrimSpace(strings.TrimSuffix(string(ty), "[]")))
		}
		if strings.HasPrefix(string(ty), "repeated ") {
			return Type(strings.TrimSpace(strings.TrimPrefix(string(ty), "repeated ")))
		}
	}
	log.Panicf("not repeated: %s", ty)
	panic("unreachable")
}

func (ty Type) KV() (k Type, v Type) {
	if ty.IsMap() {
		t := strings.TrimSuffix(strings.TrimPrefix(string(ty), "map<"), ">")
		s := strings.Split(t, ",")
		if len(s) != 2 {
			log.Panicf("invalid map: %s", ty)
		}
		return Type(strings.TrimSpace(s[0])), Type(strings.TrimSpace(s[1]))
	}
	log.Panicf("not map: %s", ty)
	panic("unreachable")
}

func (ty Type) Repeated() Type {
	return Type(string(ty) + "[]")
}

type Meta struct {
	Separator         string   `form:"separator"`
	Scope             []string `form:"scope"`
	UniqueIndex       []int32  `form:"unique_index"`
	HashUniqueIndex   []int32  `form:"hash_unique_index"`
	SortedUniqueIndex []int32  `form:"sorted_unique_index"`
}

var defaultMeta = &Meta{
	Separator:         ",",
	Scope:             nil,
	UniqueIndex:       nil,
	HashUniqueIndex:   nil,
	SortedUniqueIndex: nil,
}

func parseMeta(str string) (*Meta, error) {
	if str == "" {
		return defaultMeta, nil
	}

	meta := *defaultMeta

	values, err := url.ParseQuery(str)
	if err != nil {
		return nil, err
	}

	err = form.NewDecoder().Decode(&meta, values)
	if err != nil {
		return nil, err
	}

	meta.Separator = strings.TrimSpace(meta.Separator)

	meta.Scope = pie.Of(meta.Scope).Map(func(s string) string {
		return strings.TrimSpace(s)
	}).Filter(func(s string) bool {
		return s != ""
	}).Result

	meta.UniqueIndex = normalizeIndexTags(meta.UniqueIndex)
	if len(meta.UniqueIndex) > 0 {
		switch viper.GetString("pb_unique_index_as") {
		case "hash_unique_index":
			meta.HashUniqueIndex = append(meta.HashUniqueIndex, meta.UniqueIndex...)
		case "sorted_unique_index":
			meta.SortedUniqueIndex = append(meta.SortedUniqueIndex, meta.UniqueIndex...)
		}
	}

	meta.HashUniqueIndex = normalizeIndexTags(meta.HashUniqueIndex)
	meta.SortedUniqueIndex = normalizeIndexTags(meta.SortedUniqueIndex)

	if conflicted := pie.Of(meta.HashUniqueIndex).Filter(func(tag int32) bool {
		return pie.Contains(meta.SortedUniqueIndex, tag)
	}).Result; len(conflicted) > 0 {
		return nil, fmt.Errorf("hash_unique_index tags %v conflict with sorted_unique_index", conflicted)
	}

	return &meta, nil
}

func normalizeIndexTags(tags []int32) []int32 {
	if len(tags) == 0 {
		return nil
	}

	seen := make(map[int32]struct{}, len(tags))
	out := make([]int32, 0, len(tags))

	for _, tag := range tags {
		if tag < 0 {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})

	if len(out) == 0 {
		return nil
	}

	return out
}

type Field struct {
	*Decl
	IsColumn bool
	Number   int
	Name     string
	Alias    string
	Default  string
	Meta     *Meta
	Comment  string
}

func (f *Field) MatchTargets() bool {
	targets := pie.Of(viper.GetStringSlice("targets")).Map(func(target string) string {
		return strings.TrimSpace(target)
	}).Filter(func(target string) bool {
		return target != ""
	}).Result

	if len(targets) <= 0 || len(f.Meta.Scope) <= 0 {
		return true
	}
	if len(f.Meta.HashUniqueIndex) > 0 || len(f.Meta.SortedUniqueIndex) > 0 {
		return true
	}

	return pie.Any(targets, func(target string) bool {
		return pie.Any(f.Meta.Scope, func(scope string) bool {
			return strings.EqualFold(scope, target)
		})
	})
}

func (f *Field) ProtobufMeta() string {
	sb := strings.Builder{}

	if f.IsRepeated {
		sb.WriteString(fmt.Sprintf("(%s.Separator) = '%s'", viper.GetString("pb_package"), f.Meta.Separator))
	}

	if f.IsColumn {
		for _, tag := range f.Meta.HashUniqueIndex {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("(%s.HashUniqueIndex) = %d", viper.GetString("pb_package"), tag))
		}
		for _, tag := range f.Meta.SortedUniqueIndex {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("(%s.SortedUniqueIndex_) = %d", viper.GetString("pb_package"), tag))
		}
	}

	for _, scope := range f.Meta.Scope {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("(%s.Scope) = '%s'", viper.GetString("pb_package"), scope))
	}

	if f.Alias != "" {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		if f.IsEnumValue {
			sb.WriteString(fmt.Sprintf("(%s.EnumValueAlias) = '%s'", viper.GetString("pb_package"), f.Alias))
		} else {
			sb.WriteString(fmt.Sprintf("(%s.FieldAlias) = '%s'", viper.GetString("pb_package"), f.Alias))
		}
	}

	if sb.Len() <= 0 {
		return ""
	}

	return fmt.Sprintf(" [%s]", sb.String())
}

type Mapping struct {
	K, V *Decl
}

type Decl struct {
	File        string
	Sheet       string
	Line        int
	Type        Type
	IsBuiltin   bool
	IsTable     bool
	IsStruct    bool
	IsEnum      bool
	IsEnumValue bool
	IsRepeated  bool
	IsMap       bool
	Fields      generic.UnorderedSliceMap[string, *Field]
	FieldsAlias generic.UnorderedSliceMap[string, *Field]
	Child       *Field
	Mapping     *Mapping
	EnumValue   string
}

func (d *Decl) EnumFields() generic.UnorderedSliceMap[string, *Field] {
	fields := make(generic.UnorderedSliceMap[string, *Field], 0, d.Fields.Len())

	d.Fields.Each(func(name string, decl *Field) {
		if !decl.MatchTargets() {
			return
		}
		fields.Add(name, decl)
	})

	sort.SliceStable(fields, func(i, j int) bool {
		a, _ := strconv.Atoi(fields[i].V.EnumValue)
		b, _ := strconv.Atoi(fields[j].V.EnumValue)
		return a < b
	})

	return fields
}

func (d *Decl) StructFields() generic.UnorderedSliceMap[string, *Field] {
	fields := make(generic.UnorderedSliceMap[string, *Field], 0, d.Fields.Len())

	d.Fields.Each(func(name string, decl *Field) {
		if !decl.MatchTargets() {
			return
		}
		fields.Add(name, decl)
	})

	return fields
}

func (d *Decl) StructHashUniqueIndexes() generic.UnorderedSliceMap[string, string] {
	return d.structUniqueIndexes(func(field *Field) []int32 {
		return field.Meta.HashUniqueIndex
	})
}

func (d *Decl) StructSortedUniqueIndexes() generic.UnorderedSliceMap[string, string] {
	return d.structUniqueIndexes(func(field *Field) []int32 {
		return field.Meta.SortedUniqueIndex
	})
}

func (d *Decl) structUniqueIndexes(tagsOf func(field *Field) []int32) generic.UnorderedSliceMap[string, string] {
	var tagFields generic.SliceMap[int32, []*Field]

	d.Fields.Each(func(name string, decl *Field) {
		for _, tag := range tagsOf(decl) {
			fields, ok := tagFields.Get(tag)
			if ok {
				tagFields.Add(tag, append(fields, decl))
			} else {
				tagFields.Add(tag, []*Field{decl})
			}
		}
	})

	indexes := make(generic.UnorderedSliceMap[string, string], 0, tagFields.Len())

	tagFields.Each(func(tag int32, fields []*Field) {
		if pie.Any(fields, func(field *Field) bool { return !field.MatchTargets() }) {
			return
		}

		name := strings.Builder{}
		indexFields := strings.Builder{}

		for _, f := range fields {
			name.WriteString(f.Name)

			if indexFields.Len() > 0 {
				indexFields.WriteString(",")
			}
			indexFields.WriteString(f.Name)
		}

		indexes.TryAdd(name.String(), indexFields.String())
	})

	return indexes
}

func (d *Decl) ProtoType() string {
	if d.IsEnum {
		return string(d.Type) + ".Enum"
	}
	if d.IsRepeated {
		return "repeated " + d.Child.ProtoType()
	}
	return string(d.Type)
}

type Row []string

func (row Row) Get(idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.NewReplacer("\r", "", "\n", "\\n").Replace(strings.TrimSpace(row[idx]))
}

func (row Row) Empty() bool {
	for i := range row {
		if row.Get(i) != "" {
			return false
		}
	}
	return true
}

const (
	SheetTypes       = "@types"
	SheetTypesHeader = 1
)

func parseTypeDecls(file *excelize.File, globalDecls *generic.SliceMap[Type, *Decl]) *generic.SliceMap[Type, *Decl] {
	var decls generic.SliceMap[Type, *Decl]

	rows, err := file.Rows(SheetTypes)
	if err != nil {
		if _, ok := err.(excelize.ErrSheetNotExist); ok {
			return &decls
		}
		log.Panicf("read excel file %q sheet %q failed, %s", file.Path, SheetTypes, err)
	}
	defer rows.Close()

	type Columns struct {
		Type      int
		FieldName int
		FieldType int
		EnumValue int
		Alias     int
		Default   int
		Meta      int
		Comment   int
	}

	type FieldDesc struct {
		Type      string
		IsStruct  bool
		IsEnum    bool
		FieldName string
		FieldType string
		EnumValue string
		Alias     string
		Default   string
		Meta      string
		Comment   string
	}

	columns := Columns{
		Type:      -1,
		FieldName: -1,
		FieldType: -1,
		EnumValue: -1,
		Alias:     -1,
		Default:   -1,
		Meta:      -1,
		Comment:   -1,
	}

	for i := 1; rows.Next(); i++ {
		if i <= SheetTypesHeader {
			switch i {
			case SheetTypesHeader:
				row, err := rows.Columns()
				if err != nil {
					log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, SheetTypes, i, err)
				}

				for j, cell := range row {
					row[j] = strings.NewReplacer("\r", "", "\n", "\\n").Replace(strings.TrimSpace(cell))
				}

				for j, cell := range row {
					switch cell {
					case "类型", "对象类型", "ObjectType", "Type":
						columns.Type = j
					case "字段名", "FieldName":
						columns.FieldName = j
					case "字段类型", "FieldType":
						columns.FieldType = j
					case "值", "枚举值", "Value", "EnumValue":
						columns.EnumValue = j
					case "别名", "Alias":
						columns.Alias = j
					case "默认值", "Default":
						columns.Default = j
					case "元数据", "特性", "Meta":
						columns.Meta = j
					case "注释", "Comment":
						columns.Comment = j
					}
				}
			}
			continue
		}

		_row, err := rows.Columns()
		if err != nil {
			log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, SheetTypes, i, err)
		}
		row := Row(_row)

		fieldDesc := &FieldDesc{
			Type:      row.Get(columns.Type),
			IsStruct:  row.Get(columns.EnumValue) == "",
			IsEnum:    row.Get(columns.EnumValue) != "",
			FieldName: snake2Camel(row.Get(columns.FieldName)),
			FieldType: row.Get(columns.FieldType),
			EnumValue: row.Get(columns.EnumValue),
			Alias:     row.Get(columns.Alias),
			Default:   row.Get(columns.Default),
			Meta:      row.Get(columns.Meta),
			Comment:   row.Get(columns.Comment),
		}

		if fieldDesc.Type == "" {
			continue
		}

		ty := Type(fieldDesc.Type)

		if ty.IsBuiltin() {
			log.Panicf("read excel file %q sheet %q row %d failed: built-in types cannot be defined", file.Path, SheetTypes, i)
		}

		ty = Type(snake2Camel(string(ty)))

		if ty.IsRepeated() {
			log.Panicf("read excel file %q sheet %q row %d failed: array types cannot be defined", file.Path, SheetTypes, i)
		}

		typeDecl, ok := decls.Get(ty)
		if !ok {
			typeDecl = &Decl{
				File:     file.Path,
				Sheet:    SheetTypes,
				Line:     i,
				Type:     ty,
				IsStruct: fieldDesc.IsStruct,
				IsEnum:   fieldDesc.IsEnum,
			}
			decls.Add(ty, typeDecl)
		}

		if typeDecl.Type != ty || typeDecl.IsStruct != fieldDesc.IsStruct || typeDecl.IsEnum != fieldDesc.IsEnum {
			log.Panicf("read excel file %q sheet %q row %d failed: does not match previously defined type %q", file.Path, SheetTypes, i, typeDecl.Type)
		}

		meta, err := parseMeta(fieldDesc.Meta)
		if err != nil {
			log.Panicf("read excel file %q sheet %q row %d failed: parse meta %q failed, %s", file.Path, SheetTypes, i, fieldDesc.Meta, err)
		}

		fieldName := fieldDesc.FieldName
		fieldType := Type(fieldDesc.FieldType)
		fieldAlias := fieldDesc.Alias

		repeated := fieldType.IsRepeated()
		if repeated {
			fieldType = fieldType.Child()
		}

		if !fieldType.IsBuiltin() {
			fieldType = Type(snake2Camel(string(fieldType)))
		}

		if typeDecl.IsEnum {
			if repeated {
				log.Panicf("read excel file %q sheet %q row %d failed: enum field types cannot be arrays", file.Path, SheetTypes, i)
			}

			ev, err := strconv.Atoi(fieldDesc.EnumValue)
			if err != nil {
				log.Panicf("read excel file %q sheet %q row %d failed: parse enum field value failed, %s", file.Path, SheetTypes, i, err)
			}

			if ev < 0 {
				log.Panicf("read excel file %q sheet %q row %d failed: enum field value cannot be negative", file.Path, SheetTypes, i)
			}

			fieldType = Int32
		}

		field := &Field{
			Meta: defaultMeta,
		}

		if fieldType.IsBuiltin() {
			field.Decl = &Decl{
				Type:        fieldType,
				IsBuiltin:   true,
				IsEnumValue: typeDecl.IsEnum,
			}

			if typeDecl.IsEnum {
				field.EnumValue = fieldDesc.EnumValue
			} else {
				field.Default = fieldDesc.Default
			}

		} else {
			fieldDecl, ok := decls.Get(fieldType)
			if !ok {
				fieldDecl, ok = globalDecls.Get(fieldType)
				if !ok {
					log.Panicf("read excel file %q sheet %q row %d failed: field type %q is undefined", file.Path, SheetTypes, i, fieldType)
				}
			}
			field.Decl = fieldDecl
			field.Default = fieldDesc.Default
		}

		if repeated {
			fieldNumber := typeDecl.Fields.Len() + 1

			parent := &Field{
				Decl: &Decl{
					Type:       fieldType.Repeated(),
					IsRepeated: true,
					Child:      field,
				},
				Number:  fieldNumber,
				Name:    fieldName,
				Alias:   fieldAlias,
				Meta:    meta,
				Comment: fieldDesc.Comment,
			}

			typeDecl.Fields.Add(parent.Name, parent)
			if parent.Alias != "" {
				typeDecl.FieldsAlias.Add(parent.Alias, parent)
			}

		} else {
			field.Number = typeDecl.Fields.Len() + 1
			field.Name = fieldName
			field.Alias = fieldAlias
			field.Meta = meta
			field.Comment = fieldDesc.Comment

			typeDecl.Fields.Add(field.Name, field)
			if field.Alias != "" {
				typeDecl.FieldsAlias.Add(field.Alias, field)
			}
		}
	}

	return &decls
}
