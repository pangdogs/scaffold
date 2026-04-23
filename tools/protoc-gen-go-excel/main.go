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
	"errors"
	"fmt"
	"slices"
	"strings"

	"git.golaxy.org/core/utils/generic"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	excelutilsPackage = protogen.GoImportPath("git.golaxy.org/scaffold/tools/excelc/excelutils")
	bytesPackage      = protogen.GoImportPath("bytes")
	slicesPackage     = protogen.GoImportPath("slices")
	mathPackage       = protogen.GoImportPath("math")
)

const (
	indexTypeHashUnique   = "HashUniqueIndex"
	indexTypeSortedUnique = "SortedUniqueIndex"
)

type FieldDecl struct {
	Field  *protogen.Field
	GOType string
}

type ProtoDescriptors interface {
	Enums() protoreflect.EnumDescriptors
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}

type Extensions struct {
	IsTable,
	IsColumns,
	IndexType,
	IndexFields protoreflect.ExtensionType
}

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if err := registerProtoTypes(protoregistry.GlobalTypes, f.Desc); err != nil {
				return fmt.Errorf("register proto types for %q: %w", f.Desc.Path(), err)
			}
		}
		for _, f := range gen.Files {
			if f.Generate {
				if err := generateFile(gen, f); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) error {
	fileName := file.GeneratedFilenamePrefix + ".excel.go"
	g := gen.NewGeneratedFile(fileName, file.GoImportPath)

	genGeneratedHeader(gen, file, g)

	g.P("package ", file.GoPackageName)
	g.P()

	ext, err := parseExtensions(file)
	if err != nil {
		return err
	}

	indexTypeName := protoFullName(file, "IndexType.Enum")
	indexType, err := protoregistry.GlobalTypes.FindEnumByName(indexTypeName)
	if err != nil {
		return fmt.Errorf("parse proto type %q failed, %s", indexTypeName, err)
	}

	for i, m := range file.Messages {
		pbMsg := file.Proto.MessageType[i]

		isTable := proto.GetExtension(pbMsg.Options, ext.IsTable).(bool)
		if !isTable {
			continue
		}

		fieldRowsIdx := slices.IndexFunc(m.Fields, func(field *protogen.Field) bool {
			return string(field.Desc.Name()) == "Rows"
		})
		if fieldRowsIdx < 0 {
			continue
		}
		fieldRows := m.Fields[fieldRowsIdx]
		defaultMethodsEmitted := false

		for j, f := range m.Fields {
			indexTypeValue, ok := proto.GetExtension(pbMsg.Field[j].Options, ext.IndexType).(protoreflect.EnumNumber)
			if !ok || indexTypeValue <= 0 {
				continue
			}

			indexTypeValueDesc := indexType.Descriptor().Values().ByNumber(indexTypeValue)
			if indexTypeValueDesc == nil {
				continue
			}

			indexFields := proto.GetExtension(pbMsg.Field[j].Options, ext.IndexFields).(string)
			if indexFields == "" {
				continue
			}

			indexFieldDecls := generic.UnorderedSliceMap[string, *FieldDecl]{}

			for _, indexFieldName := range strings.Split(indexFields, ",") {
				idx := slices.IndexFunc(fieldRows.Message.Fields, func(f *protogen.Field) bool {
					return string(f.Desc.Name()) == indexFieldName
				})
				if idx < 0 {
					return fmt.Errorf("parse proto type %q failed, index field %q not found", fieldRows.Message.Desc.Name(), indexFieldName)
				}

				fd := &FieldDecl{}
				fd.Field = fieldRows.Message.Fields[idx]

				ty, p := fieldGoType(g, fd.Field)
				if p {
					ty = "*" + ty
				}
				fd.GOType = ty

				indexFieldDecls.Add(indexFieldName, fd)
			}

			var indexArgs strings.Builder

			indexFieldDecls.Each(func(name string, fd *FieldDecl) {
				if indexArgs.Len() > 0 {
					indexArgs.WriteString(", ")
				}
				indexArgs.WriteString(name)
				indexArgs.WriteString(" ")
				indexArgs.WriteString(fd.GOType)
			})

			var notFoundArgs strings.Builder

			indexFieldDecls.Each(func(name string, fd *FieldDecl) {
				if notFoundArgs.Len() > 0 {
					notFoundArgs.WriteString(", ")
				}
				notFoundArgs.WriteString(`"`)
				notFoundArgs.WriteString(name)
				notFoundArgs.WriteString(`", `)
				notFoundArgs.WriteString(name)
			})

			indexShorten := f.GoName
			switch indexTypeValueDesc.Name() {
			case indexTypeHashUnique:
				indexShorten = indexTypeHashUnique + strings.TrimPrefix(indexShorten, indexTypeHashUnique)
			case indexTypeSortedUnique:
				indexShorten = indexTypeSortedUnique + strings.TrimPrefix(indexShorten, indexTypeSortedUnique)
			}

			if !defaultMethodsEmitted {
				g.P("func (x *", m.GoIdent, ") Lookup(", indexArgs.String(), ") (*", fieldRows.Message.GoIdent, ", bool) {")
				g.P("\treturn x.LookupBy", indexShorten, "(", indexFieldNames(indexFieldDecls), ")")
				g.P("}")
				g.P()

				g.P("func (x *", m.GoIdent, ") Get(", indexArgs.String(), ") *", fieldRows.Message.GoIdent, " {")
				g.P("\treturn x.GetBy", indexShorten, "(", indexFieldNames(indexFieldDecls), ")")
				g.P("}")
				g.P()

				defaultMethodsEmitted = true
			}

			g.P("func (x *", m.GoIdent, ") LookupBy", indexShorten, "(", indexArgs.String(), ") (*", fieldRows.Message.GoIdent, ", bool) {")
			g.P("if x.", fieldRows.GoName, " == nil {")
			g.P("\treturn nil, false")
			g.P("}")
			g.P()

			switch indexTypeValueDesc.Name() {
			case indexTypeHashUnique:
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls, false, notFoundArgs.String())

				g.P("offset, ok := x.", f.GoName, "[idx]")
				g.P("if !ok {")
				g.P("\treturn nil, false")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[offset]")
				g.P()

				if hv {
					emitRowMatchesFunc(g, fieldRows.Message.GoIdent, indexFieldDecls)
					g.P("if matchesRow(row) {")
					g.P("\treturn row, true")
					g.P("}")
					g.P()
					g.P("bucket, ok := x.", f.GoName, "Conflict[idx]")
					g.P("if !ok {")
					g.P("\treturn nil, false")
					g.P("}")
					g.P()
					g.P("for _, conflictOffset := range bucket.Offsets {")
					g.P("\trow = x.", fieldRows.GoName, "[conflictOffset]")
					g.P("\tif matchesRow(row) {")
					g.P("\t\treturn row, true")
					g.P("\t}")
					g.P("}")
					g.P()
					g.P("return nil, false")
				} else {
					g.P("return row, true")
				}

			case indexTypeSortedUnique:
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls, false, notFoundArgs.String())

				g.P("if x.", f.GoName, " == nil {")
				g.P("\treturn nil, false")
				g.P("}")
				g.P()
				g.P("itemOffset, ok := ", slicesPackage.Ident("BinarySearch"), "(x.", f.GoName, ".Values, idx)")
				g.P("if !ok {")
				g.P("\treturn nil, false")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[x.", f.GoName, ".Offsets[itemOffset]]")
				g.P()

				if hv {
					emitRowMatchesFunc(g, fieldRows.Message.GoIdent, indexFieldDecls)
					g.P("if matchesRow(row) {")
					g.P("\treturn row, true")
					g.P("}")
					g.P()
					g.P("bucket, ok := x.", f.GoName, "Conflict[idx]")
					g.P("if !ok {")
					g.P("\treturn nil, false")
					g.P("}")
					g.P()
					g.P("for _, conflictOffset := range bucket.Offsets {")
					g.P("\trow = x.", fieldRows.GoName, "[conflictOffset]")
					g.P("\tif matchesRow(row) {")
					g.P("\t\treturn row, true")
					g.P("\t}")
					g.P("}")
					g.P()
					g.P("return nil, false")
				} else {
					g.P("return row, true")
				}
			}

			g.P("}")
			g.P()

			g.P("func (x *", m.GoIdent, ") GetBy", indexShorten, "(", indexArgs.String(), ") *", fieldRows.Message.GoIdent, " {")
			g.P("if x.", fieldRows.GoName, " == nil {")
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
			g.P("}")
			g.P()

			switch indexTypeValueDesc.Name() {
			case indexTypeHashUnique:
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls, true, notFoundArgs.String())

				g.P("offset, ok := x.", f.GoName, "[idx]")
				g.P("if !ok {")
				g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[offset]")
				g.P()

				if hv {
					emitRowMatchesFunc(g, fieldRows.Message.GoIdent, indexFieldDecls)
					g.P("if matchesRow(row) {")
					g.P("\treturn row")
					g.P("}")
					g.P()
					g.P("bucket, ok := x.", f.GoName, "Conflict[idx]")
					g.P("if !ok {")
					g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
					g.P("}")
					g.P()
					g.P("for _, conflictOffset := range bucket.Offsets {")
					g.P("\trow = x.", fieldRows.GoName, "[conflictOffset]")
					g.P("\tif matchesRow(row) {")
					g.P("\t\treturn row")
					g.P("\t}")
					g.P("}")
					g.P()
					g.P("panic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
				} else {
					g.P("return row")
				}

			case indexTypeSortedUnique:
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls, true, notFoundArgs.String())

				g.P("if x.", f.GoName, " == nil {")
				g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
				g.P("}")
				g.P()
				g.P("itemOffset, ok := ", slicesPackage.Ident("BinarySearch"), "(x.", f.GoName, ".Values, idx)")
				g.P("if !ok {")
				g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[x.", f.GoName, ".Offsets[itemOffset]]")
				g.P()

				if hv {
					emitRowMatchesFunc(g, fieldRows.Message.GoIdent, indexFieldDecls)
					g.P("if matchesRow(row) {")
					g.P("\treturn row")
					g.P("}")
					g.P()
					g.P("bucket, ok := x.", f.GoName, "Conflict[idx]")
					g.P("if !ok {")
					g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
					g.P("}")
					g.P()
					g.P("for _, conflictOffset := range bucket.Offsets {")
					g.P("\trow = x.", fieldRows.GoName, "[conflictOffset]")
					g.P("\tif matchesRow(row) {")
					g.P("\t\treturn row")
					g.P("\t}")
					g.P("}")
					g.P()
					g.P("panic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs.String(), "))")
				} else {
					g.P("return row")
				}
			}

			g.P("}")
			g.P()
		}
	}

	return nil
}

func indexFieldNames(fieldDecls generic.UnorderedSliceMap[string, *FieldDecl]) string {
	var names strings.Builder
	fieldDecls.Each(func(name string, fd *FieldDecl) {
		if names.Len() > 0 {
			names.WriteString(", ")
		}
		names.WriteString(name)
	})
	return names.String()
}

func genGeneratedHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("// Code generated by protoc-gen-go-excel. DO NOT EDIT.")

	g.P("// versions:")
	protocVersion := "(unknown)"
	if v := gen.Request.GetCompilerVersion(); v != nil {
		protocVersion = fmt.Sprintf("v%v.%v.%v", v.GetMajor(), v.GetMinor(), v.GetPatch())
		if s := v.GetSuffix(); s != "" {
			protocVersion += "-" + s
		}
	}
	g.P("// \tprotoc        ", protocVersion)

	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
}

func registerProtoTypes(pbTypes *protoregistry.Types, desc ProtoDescriptors) error {
	for i := range desc.Extensions().Len() {
		ext := desc.Extensions().Get(i)

		_, err := pbTypes.FindExtensionByName(ext.FullName())
		if err == nil {
			continue
		}
		if !errors.Is(err, protoregistry.NotFound) {
			return err
		}

		err = pbTypes.RegisterExtension(dynamicpb.NewExtensionType(ext))
		if err != nil {
			return err
		}
	}

	for i := range desc.Enums().Len() {
		enum := desc.Enums().Get(i)

		_, err := pbTypes.FindEnumByName(enum.FullName())
		if err == nil {
			continue
		}
		if !errors.Is(err, protoregistry.NotFound) {
			return err
		}

		err = pbTypes.RegisterEnum(dynamicpb.NewEnumType(enum))
		if err != nil {
			return err
		}
	}

	for i := range desc.Messages().Len() {
		msg := desc.Messages().Get(i)

		_, err := pbTypes.FindMessageByName(msg.FullName())
		if err == nil {
			continue
		}
		if !errors.Is(err, protoregistry.NotFound) {
			return err
		}

		err = pbTypes.RegisterMessage(dynamicpb.NewMessageType(msg))
		if err != nil {
			return err
		}

		err = registerProtoTypes(pbTypes, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseExtensions(file *protogen.File) (*Extensions, error) {
	extensions := &Extensions{}
	var err error

	extName := protoFullName(file, "IsTable")
	extensions.IsTable, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoFullName(file, "IsColumns")
	extensions.IsColumns, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoFullName(file, "IndexType_")
	extensions.IndexType, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoFullName(file, "IndexFields")
	extensions.IndexFields, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	return extensions, nil
}

func protoFullName(file *protogen.File, suffix string) protoreflect.FullName {
	if pkg := string(file.Desc.Package()); pkg != "" {
		return protoreflect.FullName(pkg + "." + suffix)
	}
	return protoreflect.FullName(suffix)
}

func fieldGoType(g *protogen.GeneratedFile, field *protogen.Field) (goType string, pointer bool) {
	if field.Desc.IsWeak() {
		return "struct{}", false
	}

	pointer = field.Desc.HasPresence()
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
		pointer = false // rely on nullability of slices for presence
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = "*" + g.QualifiedGoIdent(field.Message.GoIdent)
		pointer = false // pointer captured as part of the type
	}
	switch {
	case field.Desc.IsList():
		return "[]" + goType, false
	case field.Desc.IsMap():
		keyType, _ := fieldGoType(g, field.Message.Fields[0])
		valType, _ := fieldGoType(g, field.Message.Fields[1])
		return fmt.Sprintf("map[%v]%v", keyType, valType), false
	}
	return goType, pointer
}

func singleFieldToIndex(g *protogen.GeneratedFile, name string, decl *FieldDecl, panicForNotFound bool, notFoundArgs string) (hashVerification bool) {
	defer g.P()

	if decl.Field.Desc.IsMap() {
		g.P("idx, err := ", excelutilsPackage.Ident("MapToIndex"), "(", name, ")")
		g.P("if err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
		return true
	}

	if decl.Field.Desc.IsList() {
		g.P("idx, err := ", excelutilsPackage.Ident("ListToIndex"), "(", name, ")")
		g.P("if err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
		return true
	}

	switch decl.Field.Desc.Kind() {
	case protoreflect.BoolKind:
		g.P("idx := ", excelutilsPackage.Ident("BooleanToIndex"), "(", name, ")")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		g.P("idx := ", excelutilsPackage.Ident("IntegerToIndex"), "(", name, ")")
	case protoreflect.FloatKind:
		g.P("idx := ", excelutilsPackage.Ident("FloatToIndex"), "(", name, ")")
	case protoreflect.DoubleKind:
		g.P("idx := ", excelutilsPackage.Ident("DoubleToIndex"), "(", name, ")")
	case protoreflect.StringKind:
		g.P("idx, err := ", excelutilsPackage.Ident("StringToIndex"), "(", name, ")")
		g.P("if err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
		return true
	case protoreflect.BytesKind:
		g.P("idx, err := ", excelutilsPackage.Ident("BytesToIndex"), "(", name, ")")
		g.P("if err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
		return true
	case protoreflect.EnumKind:
		g.P("idx := ", excelutilsPackage.Ident("IntegerToIndex"), "(", name, ")")
	default:
		g.P("idx, err := ", excelutilsPackage.Ident("AnyToIndex"), "(", name, ")")
		g.P("if err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
		return true
	}

	return false
}

func fieldsToIndex(g *protogen.GeneratedFile, fieldDecls generic.UnorderedSliceMap[string, *FieldDecl], panicForNotFound bool, notFoundArgs string) (hashVerification bool) {
	if fieldDecls.Len() <= 1 {
		return singleFieldToIndex(g, fieldDecls[0].K, fieldDecls[0].V, panicForNotFound, notFoundArgs)
	}

	g.P("h := ", excelutilsPackage.Ident("NewHash"), "()")

	fieldDecls.Each(func(name string, decl *FieldDecl) {
		if decl.Field.Desc.IsMap() {
			g.P("if err := ", excelutilsPackage.Ident("MapToHash"), "(h, ", name, "); err != nil {")
			if panicForNotFound {
				g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
			} else {
				g.P("\treturn nil, false")
			}
			g.P("}")
			return
		}

		if decl.Field.Desc.IsList() {
			g.P("if err := ", excelutilsPackage.Ident("ListToHash"), "(h, ", name, "); err != nil {")
			if panicForNotFound {
				g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
			} else {
				g.P("\treturn nil, false")
			}
			g.P("}")
			return
		}

		g.P("if err := ", excelutilsPackage.Ident("AnyToHash"), "(h, ", name, "); err != nil {")
		if panicForNotFound {
			g.P("\tpanic(", excelutilsPackage.Ident("NewErrNotFound("), notFoundArgs, "))")
		} else {
			g.P("\treturn nil, false")
		}
		g.P("}")
	})

	g.P("idx := h.Sum64()")
	g.P()

	return true
}

func compareFunc(fd *FieldDecl) []any {
	compareField := fd.Field
	ty := fd.GOType

	if fd.Field.Desc.IsMap() {
		compareField = fd.Field.Message.Fields[1]
		idx := strings.IndexByte(fd.GOType, ']')
		ty = fd.GOType[idx+1:]
	}

	if fd.Field.Desc.IsList() {
		ty = strings.TrimPrefix(fd.GOType, "[]")
	}

	switch compareField.Desc.Kind() {
	case protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.StringKind, protoreflect.EnumKind:
		return []any{fmt.Sprintf("func(a, b %s) bool { return a == b }", ty)}
	case protoreflect.FloatKind:
		return []any{"func(a, b ", ty, ") bool { return ", mathPackage.Ident("Float32bits"), "(a) == ", mathPackage.Ident("Float32bits"), "(b) }"}
	case protoreflect.DoubleKind:
		return []any{"func(a, b ", ty, ") bool { return ", mathPackage.Ident("Float64bits"), "(a) == ", mathPackage.Ident("Float64bits"), "(b) }"}
	case protoreflect.BytesKind:
		return []any{"func(a, b ", ty, ") bool { return ", bytesPackage.Ident("Compare"), "(a, b) == 0 }"}
	default:
		return []any{"func(a, b ", ty, ") bool { return ", excelutilsPackage.Ident("ProtoMessageEqual"), "(a, b) }"}
	}
}

func emitRowMatchesFunc(g *protogen.GeneratedFile, rowType protogen.GoIdent, fieldDecls generic.UnorderedSliceMap[string, *FieldDecl]) {
	g.P("matchesRow := func(row *", rowType, ") bool {")

	fieldDecls.Each(func(name string, decl *FieldDecl) {
		if decl.Field.Desc.IsMap() {
			output := []any{"\tif !", excelutilsPackage.Ident("MapEqual"), "(", name, ", row.", name, ", "}
			output = append(output, compareFunc(decl)...)
			output = append(output, ") {")
			g.P(output...)
			g.P("\t\treturn false")
			g.P("\t}")
			return
		}

		if decl.Field.Desc.IsList() {
			output := []any{"\tif !", excelutilsPackage.Ident("ListEqual"), "(", name, ", row.", name, ", "}
			output = append(output, compareFunc(decl)...)
			output = append(output, ") {")
			g.P(output...)
			g.P("\t\treturn false")
			g.P("\t}")
			return
		}

		switch decl.Field.Desc.Kind() {
		case protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.StringKind, protoreflect.EnumKind:
			g.P("\tif ", name, " != row.", name, " {")
			g.P("\t\treturn false")
			g.P("\t}")

		case protoreflect.FloatKind:
			g.P("\tif ", mathPackage.Ident("Float32bits"), "(", name, ") != ", mathPackage.Ident("Float32bits"), "(row.", name, ") {")
			g.P("\t\treturn false")
			g.P("\t}")

		case protoreflect.DoubleKind:
			g.P("\tif ", mathPackage.Ident("Float64bits"), "(", name, ") != ", mathPackage.Ident("Float64bits"), "(row.", name, ") {")
			g.P("\t\treturn false")
			g.P("\t}")

		case protoreflect.BytesKind:
			g.P("\tif ", bytesPackage.Ident("Compare"), "(", name, ", row.", name, ") != 0 {")
			g.P("\t\treturn false")
			g.P("\t}")

		default:
			g.P("\tif !", excelutilsPackage.Ident("ProtoMessageEqual"), "(", name, ", row.", name, ") {")
			g.P("\t\treturn false")
			g.P("\t}")
		}
	})

	g.P("\treturn true")
	g.P("}")
	g.P()
}
