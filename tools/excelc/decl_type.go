package main

import (
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/go-playground/form/v4"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

func (ty Type) GetChild() Type {
	if strings.HasPrefix(string(ty), "[]") {
		return Type(strings.TrimPrefix(string(ty), "[]"))
	}
	if strings.HasSuffix(string(ty), "[]") {
		return Type(strings.TrimSuffix(string(ty), "[]"))
	}
	if strings.HasPrefix(string(ty), "repeated ") {
		return Type(strings.TrimPrefix(string(ty), "repeated "))
	}
	panic(fmt.Errorf("not repeated: %s", ty))
}

func (ty Type) GetParent() Type {
	return Type(string(ty) + "[]")
}

type Meta struct {
	UniqueIndex       int32  `form:"unique_index"`
	UniqueSortedIndex int32  `form:"unique_sorted_index"`
	Separator         string `form:"separator"`
}

var defaultMeta = &Meta{
	UniqueIndex:       0,
	UniqueSortedIndex: 0,
	Separator:         ",",
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

	if meta == *defaultMeta {
		return defaultMeta, nil
	}

	return &meta, nil
}

type Field struct {
	*Decl
	IsColumn bool
	Name     string
	Alias    string
	Default  string
	Meta     *Meta
	Comment  string
}

func (f *Field) ProtobufMeta() string {
	sb := strings.Builder{}

	if f.IsRepeated {
		sb.WriteString(fmt.Sprintf("(%s.Separator) = '%s'", viper.GetString("pb_package"), f.Meta.Separator))
	}

	if f.IsColumn {
		if f.Meta.UniqueIndex > 0 {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("(%s.UniqueIndex) = %d", viper.GetString("pb_package"), f.Meta.UniqueIndex))
		}
		if f.Meta.UniqueSortedIndex > 0 {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("(%s.UniqueSortedIndex) = %d", viper.GetString("pb_package"), f.Meta.UniqueSortedIndex))
		}
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
	Fields      generic.UnorderedSliceMap[string, *Field]
	FieldsAlias generic.UnorderedSliceMap[string, *Field]
	Child       *Field
	Value       string
}

func (d *Decl) EnumFields() generic.UnorderedSliceMap[string, *Field] {
	fields := make(generic.UnorderedSliceMap[string, *Field], 0, d.Fields.Len())

	d.Fields.Each(func(name string, decl *Field) {
		fields.Add(name, decl)
	})

	sort.SliceStable(fields, func(i, j int) bool {
		a, _ := strconv.Atoi(fields[i].V.Value)
		b, _ := strconv.Atoi(fields[j].V.Value)
		return a < b
	})

	return fields
}

func (d *Decl) StructFields() generic.UnorderedSliceMap[string, *Field] {
	fields := make(generic.UnorderedSliceMap[string, *Field], 0, d.Fields.Len())

	d.Fields.Each(func(name string, decl *Field) {
		fields.Add(name, decl)
	})

	return fields
}

func (d *Decl) StructUniqueIndexes() generic.UnorderedSliceMap[string, string] {
	var tagFields generic.SliceMap[int32, []*Field]

	d.Fields.Each(func(name string, decl *Field) {
		if decl.Meta.UniqueIndex > 0 {
			fields, ok := tagFields.Get(decl.Meta.UniqueIndex)
			if ok {
				tagFields.Add(decl.Meta.UniqueIndex, append(fields, decl))
			} else {
				tagFields.Add(decl.Meta.UniqueIndex, []*Field{decl})
			}
		}
	})

	indexes := make(generic.UnorderedSliceMap[string, string], 0, tagFields.Len())

	tagFields.Each(func(tag int32, fields []*Field) {
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

func (d *Decl) StructUniqueSortedIndexes() generic.UnorderedSliceMap[string, string] {
	var tagFields generic.SliceMap[int32, []*Field]

	d.Fields.Each(func(name string, decl *Field) {
		if decl.Meta.UniqueSortedIndex > 0 {
			fields, ok := tagFields.Get(decl.Meta.UniqueSortedIndex)
			if ok {
				tagFields.Add(decl.Meta.UniqueSortedIndex, append(fields, decl))
			} else {
				tagFields.Add(decl.Meta.UniqueSortedIndex, []*Field{decl})
			}
		}
	})

	indexes := make(generic.UnorderedSliceMap[string, string], 0, tagFields.Len())

	tagFields.Each(func(tag int32, fields []*Field) {
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

func (d *Decl) ProtobufType() string {
	if d.IsEnum {
		return string(d.Type) + ".Enum"
	}
	if d.IsRepeated {
		return "repeated " + d.Child.ProtobufType()
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
		panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，%s", file.Path, SheetTypes, err))
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
					panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，%s", file.Path, SheetTypes, i, err))
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
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，%s", file.Path, SheetTypes, i, err))
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
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，不能定义内置类型", file.Path, SheetTypes, i))
		}

		ty = Type(snake2Camel(string(ty)))

		if ty.IsRepeated() {
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，不能定义数组类型", file.Path, SheetTypes, i))
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
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，与已定义类型 %q 不符", file.Path, SheetTypes, i, typeDecl.Type))
		}

		meta, err := parseMeta(fieldDesc.Meta)
		if err != nil {
			panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，解析Meta %q 失败，%s", file.Path, SheetTypes, i, fieldDesc.Meta, err))
		}

		fieldName := fieldDesc.FieldName
		fieldType := Type(fieldDesc.FieldType)
		fieldAlias := fieldDesc.Alias

		repeated := fieldType.IsRepeated()
		if repeated {
			fieldType = fieldType.GetChild()
		}

		if !fieldType.IsBuiltin() {
			fieldType = Type(snake2Camel(string(fieldType)))
		}

		if typeDecl.IsEnum {
			if repeated {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，枚举字段类型不能定义为数组", file.Path, SheetTypes, i))
			}

			ev, err := strconv.Atoi(fieldDesc.EnumValue)
			if err != nil {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，解析枚举字段值错误，%s", file.Path, SheetTypes, i, err))
			}

			if ev < 0 {
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，枚举字段值不能为负数", file.Path, SheetTypes, i))
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
				field.Value = fieldDesc.EnumValue
			} else {
				field.Default = fieldDesc.Default
			}

		} else {
			fieldDecl, ok := decls.Get(fieldType)
			if !ok {
				fieldDecl, ok = globalDecls.Get(fieldType)
				if !ok {
					panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，字段类型 %q 未定义", file.Path, SheetTypes, i, fieldType))
				}
			}
			field.Decl = fieldDecl
			field.Default = fieldDesc.Default
		}

		if repeated {
			parent := &Field{
				Decl: &Decl{
					Type:       fieldType.GetParent(),
					IsRepeated: true,
					Child:      field,
				},
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
