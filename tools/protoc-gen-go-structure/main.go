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

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	protoPackage  = protogen.GoImportPath("google.golang.org/protobuf/proto")
	mapsPackage   = protogen.GoImportPath("maps")
	slicesPackage = protogen.GoImportPath("slices")
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if f.Generate {
				generateFile(gen, f)
			}
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	fileName := file.GeneratedFilenamePrefix + ".structure.go"
	g := gen.NewGeneratedFile(fileName, file.GoImportPath)

	genGeneratedHeader(gen, file, g)

	g.P("package ", file.GoPackageName)
	g.P()

	g.Import(protoPackage)

	for _, m := range file.Messages {
		g.P("// Clone clone message ", m.GoIdent, "")
		g.P("func (x *", m.GoIdent, ") Clone() *", m.GoIdent, " {")
		g.P("\treturn ", protoPackage.Ident("Clone"), "(x).(*", m.GoIdent, ")")
		g.P("}")

		for _, f := range m.Fields {
			if f.Desc.IsList() {
				genListClone(g, m, f)
				continue
			}

			if f.Desc.IsMap() {
				mapDecl := mapFieldGoType(g, f)

				switch f.Desc.MapValue().Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					messageField := f.Message.Fields[1]

					g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
					g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", mapDecl, " {")
					g.P("\tif x == nil || x.", f.GoName, " == nil {")
					g.P("\t\treturn nil")
					g.P("\t}")
					g.P("\tcopied := ", mapsPackage.Ident("Clone"), "(x.", f.GoName, ")")
					g.P("\tfor k, v := range copied {")
					g.P("\t\tif v == nil {")
					g.P("\t\t\tcopied[k] = ", messageZeroExpr(g, messageField.Message))
					g.P("\t\t\tcontinue")
					g.P("\t\t}")
					g.P("\t\tcopied[k] = v.Clone()")
					g.P("\t}")
					g.P("\treturn copied")
					g.P("}")
					g.P()

				case protoreflect.BytesKind:
					g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
					g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", mapDecl, " {")
					g.P("\tif x == nil || x.", f.GoName, " == nil {")
					g.P("\t\treturn nil")
					g.P("\t}")
					g.P("\tcopied := make(", mapDecl, ", len(x.", f.GoName, "))")
					g.P("\tfor k, v := range x.", f.GoName, " {")
					g.P("\t\tcopied[k] = append([]byte{}, v...)")
					g.P("\t}")
					g.P("\treturn copied")
					g.P("}")
					g.P()

				default:
					g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
					g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", mapDecl, " {")
					g.P("\tif x == nil || x.", f.GoName, " == nil {")
					g.P("\t\treturn nil")
					g.P("\t}")
					g.P("\treturn ", mapsPackage.Ident("Clone"), "(x.", f.GoName, ")")
					g.P("}")
					g.P()
				}
				continue
			}

			switch f.Desc.Kind() {
			case protoreflect.BytesKind:
				g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
				g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", singularFieldGoType(g, f), " {")
				g.P("\tif x == nil || x.", f.GoName, " == nil {")
				g.P("\t\treturn nil")
				g.P("\t}")
				g.P("\treturn append([]byte{}, x.", f.GoName, "...)")
				g.P("}")
				g.P()
				continue

			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
				g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", singularFieldGoType(g, f), " {")
				g.P("\tif x == nil || x.", f.GoName, " == nil {")
				g.P("\t\treturn nil")
				g.P("\t}")
				g.P("\treturn x.", f.GoName, ".Clone()")
				g.P("}")
				g.P()
				continue
			}
		}
	}
}

func genListClone(g *protogen.GeneratedFile, m *protogen.Message, f *protogen.Field) {
	g.P("// Clone", f.GoName, " clone message field ", m.GoIdent, ".", f.GoName)
	g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", listFieldGoType(g, f), " {")
	g.P("\tif x == nil || x.", f.GoName, " == nil {")
	g.P("\t\treturn nil")
	g.P("\t}")

	switch {
	case f.Message != nil:
		g.P("\tcopied := make(", listFieldGoType(g, f), ", len(x.", f.GoName, "))")
		g.P("\tfor i, v := range x.", f.GoName, " {")
		g.P("\t\tif v == nil {")
		g.P("\t\t\tcopied[i] = &", g.QualifiedGoIdent(f.Message.GoIdent), "{}")
		g.P("\t\t\tcontinue")
		g.P("\t\t}")
		g.P("\t\tcopied[i] = v.Clone()")
		g.P("\t}")
		g.P("\treturn copied")
	case f.Desc.Kind() == protoreflect.BytesKind:
		g.P("\tcopied := make(", listFieldGoType(g, f), ", len(x.", f.GoName, "))")
		g.P("\tfor i, v := range x.", f.GoName, " {")
		g.P("\t\tcopied[i] = append([]byte{}, v...)")
		g.P("\t}")
		g.P("\treturn copied")
	default:
		g.P("\treturn ", slicesPackage.Ident("Clone"), "(x.", f.GoName, ")")
	}

	g.P("}")
	g.P()
}

func listFieldGoType(g *protogen.GeneratedFile, f *protogen.Field) string {
	return "[]" + singularFieldGoType(g, f)
}

func mapFieldGoType(g *protogen.GeneratedFile, f *protogen.Field) string {
	return fmt.Sprintf("map[%s]%s", singularFieldGoType(g, f.Message.Fields[0]), singularFieldGoType(g, f.Message.Fields[1]))
}

func singularFieldGoType(g *protogen.GeneratedFile, f *protogen.Field) string {
	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(f.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "*" + g.QualifiedGoIdent(f.Message.GoIdent)
	default:
		return "any"
	}
}

func messageZeroExpr(g *protogen.GeneratedFile, msg *protogen.Message) string {
	return "&" + string(g.QualifiedGoIdent(msg.GoIdent)) + "{}"
}

func genGeneratedHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("// Code generated by protoc-gen-go-structure. DO NOT EDIT.")

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
