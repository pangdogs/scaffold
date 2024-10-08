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
	"git.golaxy.org/core/utils/generic"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"slices"
	"strings"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			regProtobufTypes(protoregistry.GlobalTypes, f.Desc)
		}
		for _, f := range gen.Files {
			if f.Generate {
				generateFile(gen, f)
			}
		}
		return nil
	})
}

const (
	protoPackage      = protogen.GoImportPath("google.golang.org/protobuf/proto")
	excelutilsPackage = protogen.GoImportPath("git.golaxy.org/scaffold/tools/excelc/excelutils")
	bytesPackage      = protogen.GoImportPath("bytes")
	slicesPackage     = protogen.GoImportPath("slices")
	cmpPackage        = protogen.GoImportPath("cmp")
)

type FieldDecl struct {
	Field  *protogen.Field
	GOType string
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	fileName := file.GeneratedFilenamePrefix + ".excel.go"
	g := gen.NewGeneratedFile(fileName, file.GoImportPath)

	genGeneratedHeader(gen, file, g)

	g.P("package ", file.GoPackageName)
	g.P()

	g.Import(protoPackage)

	ext, err := parseExtensions(file)
	if err != nil {
		panic(err)
	}

	indexTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexType.Enum", file.GoPackageName))
	indexType, err := protoregistry.GlobalTypes.FindEnumByName(indexTypeName)
	if err != nil {
		panic(fmt.Errorf("解析Protobuf类型 %q 失败，%s", indexTypeName, err))
	}

	for i, m := range file.Messages {
		pbMsg := file.Proto.MessageType[i]

		isTable := proto.GetExtension(pbMsg.Options, ext.IsTable).(bool)
		if !isTable {
			continue
		}

		fieldRowsIdx := slices.IndexFunc(pbMsg.Field, func(pbField *descriptorpb.FieldDescriptorProto) bool {
			return proto.GetExtension(pbField.Options, ext.IsRows).(bool)
		})
		if fieldRowsIdx < 0 {
			continue
		}
		fieldRows := m.Fields[fieldRowsIdx]

		for j, f := range m.Fields {
			indexTypeValue, ok := proto.GetExtension(pbMsg.Field[j].Options, ext.IndexTyp).(protoreflect.EnumNumber)
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
					panic(fmt.Errorf("解析Protobuf类型 %q 失败，未找到索引字段 %q", fieldRows.Message.Desc.Name(), indexFieldName))
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

			g.P("func (x *", m.GoIdent, ") QueryBy", f.GoName, "(", indexArgs.String(), ") (*", fieldRows.Message.GoIdent, ", bool) {")
			g.P("if x.", fieldRows.GoName, " == nil {")
			g.P("\treturn nil, false")
			g.P("}")

			switch indexTypeValueDesc.Name() {
			case "UniqueIndex":
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls)

				g.P("offset, ok := x.", f.GoName, "[idx]")
				g.P("if !ok {")
				g.P("\treturn nil, false")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[offset]")
				g.P()

				if hv {
					hashVerification(g, indexFieldDecls)
				}

				g.P("return row, true")

			case "UniqueSortedIndex":
				g.P()

				hv := fieldsToIndex(g, indexFieldDecls)

				g.P("offset, ok := ", slicesPackage.Ident("BinarySearchFunc"), "(x.", f.GoName, ", idx, func(item *IndexItem, idx uint64) int { return ", cmpPackage.Ident("Compare"), "(item.Value, idx) } )")
				g.P("if !ok {")
				g.P("\treturn nil, false")
				g.P("}")
				g.P()

				g.P("row := x.", fieldRows.GoName, "[offset]")
				g.P()

				if hv {
					hashVerification(g, indexFieldDecls)
				}

				g.P("return row, true")
			}

			g.P("}")
			g.P()
		}
	}
}

func genGeneratedHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("// Code generated by protoc-gen-go-variant. DO NOT EDIT.")

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

type ProtobufDescriptors interface {
	Enums() protoreflect.EnumDescriptors
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}

func regProtobufTypes(pbTypes *protoregistry.Types, desc ProtobufDescriptors) error {
	for i := range desc.Extensions().Len() {
		ext := desc.Extensions().Get(i)

		_, err := pbTypes.FindExtensionByName(ext.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}

		err = pbTypes.RegisterExtension(dynamicpb.NewExtensionType(ext))
		if err != nil {
			return err
		}
	}

	for i := range desc.Enums().Len() {
		enum := desc.Enums().Get(i)

		_, err := pbTypes.FindExtensionByName(enum.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}

		err = pbTypes.RegisterEnum(dynamicpb.NewEnumType(enum))
		if err != nil {
			return err
		}
	}

	for i := range desc.Messages().Len() {
		msg := desc.Messages().Get(i)

		_, err := pbTypes.FindExtensionByName(msg.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}

		err = pbTypes.RegisterMessage(dynamicpb.NewMessageType(msg))
		if err != nil {
			return err
		}

		err = regProtobufTypes(pbTypes, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

type Extensions struct {
	IsTable, IsColumns, IsRows, IndexTyp, IndexFields protoreflect.ExtensionType
}

func parseExtensions(file *protogen.File) (*Extensions, error) {
	extensions := &Extensions{}
	var err error

	extName := protoreflect.FullName(fmt.Sprintf("%s.IsTable", file.GoPackageName))
	extensions.IsTable, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IsColumns", file.GoPackageName))
	extensions.IsColumns, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IsRows", file.GoPackageName))
	extensions.IsRows, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexTyp", file.GoPackageName))
	extensions.IndexTyp, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexFields", file.GoPackageName))
	extensions.IndexFields, err = protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	return extensions, nil
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

func singleFieldToIndex(g *protogen.GeneratedFile, name string, decl *FieldDecl) (hashVerification bool) {
	defer g.P()

	if decl.Field.Desc.IsMap() {
		g.P("idx, err := ", excelutilsPackage.Ident("MapToIndex"), "(", name, ")")
		g.P("if err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
		return true
	}

	if decl.Field.Desc.IsList() {
		g.P("idx, err := ", excelutilsPackage.Ident("ListToIndex"), "(", name, ")")
		g.P("if err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
		return true
	}

	switch decl.Field.Desc.Kind() {
	case protoreflect.BoolKind:
		g.P("idx := ", excelutilsPackage.Ident("BooleanToIndex"), "(", name, ")")
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		g.P("idx := ", excelutilsPackage.Ident("IntegerToIndex"), "(", name, ")")
	case protoreflect.FloatKind:
		g.P("idx := ", excelutilsPackage.Ident("FloatToIndex"), "(", name, ")")
	case protoreflect.DoubleKind:
		g.P("idx := ", excelutilsPackage.Ident("DoubleToIndex"), "(", name, ")")
	case protoreflect.StringKind:
		g.P("idx, err := ", excelutilsPackage.Ident("StringToIndex"), "(", name, ")")
		g.P("if err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
		return true
	case protoreflect.BytesKind:
		g.P("idx, err := ", excelutilsPackage.Ident("BytesToIndex"), "(", name, ")")
		g.P("if err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
		return true
	case protoreflect.EnumKind:
		g.P("idx := ", excelutilsPackage.Ident("IntegerToIndex"), "(", name, ")")
	default:
		g.P("idx, err := ", excelutilsPackage.Ident("AnyToIndex"), "(", name, ")")
		g.P("if err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
		return true
	}

	return false
}

func fieldsToIndex(g *protogen.GeneratedFile, fieldDecls generic.UnorderedSliceMap[string, *FieldDecl]) (hashVerification bool) {
	if fieldDecls.Len() <= 1 {
		return singleFieldToIndex(g, fieldDecls[0].K, fieldDecls[0].V)
	}

	g.P("h := ", excelutilsPackage.Ident("NewHash"), "()")
	g.P()

	fieldDecls.Each(func(name string, decl *FieldDecl) {
		defer g.P()

		if decl.Field.Desc.IsMap() {
			g.P("if err := ", excelutilsPackage.Ident("MapToHash"), "(h, ", name, "); err != nil {")
			g.P("\treturn nil, false")
			g.P("}")
			return
		}

		if decl.Field.Desc.IsList() {
			g.P("if err := ", excelutilsPackage.Ident("ListToHash"), "(h, ", name, "); err != nil {")
			g.P("\treturn nil, false")
			g.P("}")
			return
		}

		g.P("if err := ", excelutilsPackage.Ident("AnyToHash"), "(h, ", name, "); err != nil {")
		g.P("\treturn nil, false")
		g.P("}")
	})

	g.P("idx := h.Sum64()")
	g.P()

	return true
}

func compareFunc(fd *FieldDecl) []any {
	ty := fd.GOType

	if fd.Field.Desc.IsMap() {
		idx := strings.IndexByte(fd.GOType, ']')
		ty = fd.GOType[idx+1:]
	}

	if fd.Field.Desc.IsList() {
		ty = strings.TrimPrefix(fd.GOType, "[]")
	}

	switch fd.Field.Desc.Kind() {
	case protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.FloatKind, protoreflect.DoubleKind, protoreflect.StringKind, protoreflect.EnumKind:
		return []any{fmt.Sprintf("func(a, b %s) bool { return a == b }", ty)}
	case protoreflect.BytesKind:
		return []any{"func(a, b ", ty, ") bool { return ", bytesPackage.Ident("Compare"), "(a, b) == 0 }"}
	default:
		return []any{"func(a, b ", ty, ") bool { return ", protoPackage.Ident("Equal"), "(a, b) }"}
	}
}

func hashVerification(g *protogen.GeneratedFile, fieldDecls generic.UnorderedSliceMap[string, *FieldDecl]) {
	fieldDecls.Each(func(name string, decl *FieldDecl) {
		defer g.P()

		if decl.Field.Desc.IsMap() {
			output := []any{"if !", excelutilsPackage.Ident("MapEqual"), "(", name, ", row.", name, ", "}
			output = append(output, compareFunc(decl)...)
			output = append(output, ") {")

			g.P(output...)
			g.P("\treturn nil, false")
			g.P("}")
			return
		}

		if decl.Field.Desc.IsList() {
			output := []any{"if !", excelutilsPackage.Ident("ListEqual"), "(", name, ", row.", name, ", "}
			output = append(output, compareFunc(decl)...)
			output = append(output, ") {")

			g.P(output...)
			g.P("\treturn nil, false")
			g.P("}")
			return
		}

		switch decl.Field.Desc.Kind() {
		case protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.FloatKind, protoreflect.DoubleKind, protoreflect.StringKind, protoreflect.EnumKind:
			g.P("if ", name, " != row.", name, " {")
			g.P("\treturn nil, false")
			g.P("}")

		case protoreflect.BytesKind:
			g.P("if ", bytesPackage.Ident("Compare"), "(", name, ", row.", name, ") != 0 {")
			g.P("\treturn nil, false")
			g.P("}")

		default:
			g.P("if !", protoPackage.Ident("Equal"), "(", name, ", row.", name, ") {")
			g.P("\treturn nil, false")
			g.P("}")
		}
	})

	g.P()
}
