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

const (
	protoPackage = protogen.GoImportPath("google.golang.org/protobuf/proto")
	mapsPackage  = protogen.GoImportPath("maps")
)

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	fileName := file.GeneratedFilenamePrefix + ".structure.go"
	g := gen.NewGeneratedFile(fileName, file.GoImportPath)

	genGeneratedHeader(gen, file, g)

	g.P("package ", file.GoPackageName)
	g.P()

	g.Import(protoPackage)

	for _, m := range file.Messages {
		{
			g.P("// FillNil 将所有为nil的字段填充0值")
			g.P("func (x *", m.GoIdent, ") FillNil() *", m.GoIdent, " {")

			for _, f := range m.Fields {
				if f.Desc.IsList() {
					continue
				}

				if f.Desc.IsMap() {
					var mapValue string

					switch f.Desc.MapValue().Kind() {
					case protoreflect.MessageKind, protoreflect.GroupKind:
						mapValue = "*" + string(f.Desc.MapValue().Message().Name())
					default:
						mapValue = f.Desc.MapValue().Kind().String()
					}

					g.P("\tif x.", f.GoName, " == nil {")
					g.P("\t\tx.", f.GoName, " = map[", f.Desc.MapKey().Kind(), "]", mapValue, "{}")
					g.P("\t}")
					continue
				}

				switch f.Desc.Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					g.P("\tif x.", f.GoName, " == nil {")
					g.P("\t\tx.", f.GoName, " = &", f.Desc.Message().Name(), "{}")
					g.P("\t}")
					continue
				}
			}

			g.P("\treturn x")
			g.P("}")
			g.P()
		}

		{
			g.P("// DeepFillNil 递归将所有为nil的固定字段填充0值")
			g.P("func (x *", m.GoIdent, ") DeepFillNil() *", m.GoIdent, " {")

			for _, f := range m.Fields {
				if f.Desc.IsList() {
					continue
				}

				if f.Desc.IsMap() {
					var mapValue string

					switch f.Desc.MapValue().Kind() {
					case protoreflect.MessageKind, protoreflect.GroupKind:
						mapValue = "*" + string(f.Desc.MapValue().Message().Name())
					default:
						mapValue = f.Desc.MapValue().Kind().String()
					}

					g.P("\tif x.", f.GoName, " == nil {")
					g.P("\t\tx.", f.GoName, " = map[", f.Desc.MapKey().Kind(), "]", mapValue, "{}")
					g.P("\t}")
					continue
				}

				switch f.Desc.Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					g.P("\tif x.", f.GoName, " == nil {")
					g.P("\t\tx.", f.GoName, " = &", f.Desc.Message().Name(), "{}")
					g.P("\t}")
					g.P("\tx.", f.GoName, ".DeepFillNil()")
					continue
				}
			}

			g.P("\treturn x")
			g.P("}")
			g.P()
		}

		g.P("// Clone 克隆")
		g.P("func (x *", m.GoIdent, ") Clone() *", m.GoIdent, " {")
		g.P("\treturn ", protoPackage.Ident("Clone"), "(x).(*", m.GoIdent, ")")
		g.P("}")

		for _, f := range m.Fields {
			if f.Desc.IsList() {
				continue
			}

			if f.Desc.IsMap() {
				var mapValue string

				switch f.Desc.MapValue().Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					mapValue = "*" + string(f.Desc.MapValue().Message().Name())
				default:
					mapValue = f.Desc.MapValue().Kind().String()
				}

				mapDecl := fmt.Sprintf("map[%s]%s", f.Desc.MapKey().Kind(), mapValue)

				g.P("// FillNil", f.GoName, " 字段 ", f.GoName, " 为nil时填充0值")
				g.P("func (x *", m.GoIdent, ") FillNil", f.GoName, "() ", mapDecl, " {")
				g.P("\tif x.", f.GoName, " == nil {")
				g.P("\t\tx.", f.GoName, " = map[", f.Desc.MapKey().Kind(), "]", mapValue, "{}")
				g.P("\t}")
				g.P("\treturn x.", f.GoName, "")
				g.P("}")
				g.P()
				continue
			}

			switch f.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P("// FillNil", f.GoName, " 字段 ", f.GoName, " 为nil时填充0值")
				g.P("func (x *", m.GoIdent, ") FillNil", f.GoName, "() *", f.Desc.Message().Name(), " {")
				g.P("\treturn x.", f.GoName, ".FillNil()")
				g.P("}")
				g.P()
				continue
			}
		}

		for _, f := range m.Fields {
			if f.Desc.IsList() {
				continue
			}

			if f.Desc.IsMap() {

				switch f.Desc.MapValue().Kind() {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					mapValue := "*" + string(f.Desc.MapValue().Message().Name())
					mapDecl := fmt.Sprintf("map[%s]%s", f.Desc.MapKey().Kind(), mapValue)

					g.P("// Clone", f.GoName, " 克隆字段 ", f.GoName)
					g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", mapDecl, " {")
					g.P("\tcopied := ", mapsPackage.Ident("Clone"), "(x.", f.GoName, ")")
					g.P("\tfor k, v := range copied {")
					g.P("\t\tcopied[k] = v.Clone()")
					g.P("\t}")
					g.P("\treturn copied")
					g.P("}")
					g.P()

				default:
					mapValue := f.Desc.MapValue().Kind().String()
					mapDecl := fmt.Sprintf("map[%s]%s", f.Desc.MapKey().Kind(), mapValue)

					g.P("// Clone", f.GoName, " 克隆字段 ", f.GoName)
					g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() ", mapDecl, " {")
					g.P("\treturn ", mapsPackage.Ident("Clone"), "(x.", f.GoName, ")")
					g.P("}")
					g.P()
				}
				continue
			}

			switch f.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P("// Clone", f.GoName, " 克隆字段 ", f.GoName)
				g.P("func (x *", m.GoIdent, ") Clone", f.GoName, "() *", f.Desc.Message().Name(), " {")
				g.P("\treturn x.", f.GoName, ".Clone()")
				g.P("}")
				g.P()
				continue
			}
		}
	}
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
