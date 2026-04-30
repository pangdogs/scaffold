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
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type GDScriptTableDecl struct {
	Name             string
	ProtoAlias       string
	ExcelAlias       string
	ProtoType        string
	ExcelType        string
	ChunkedExcelType string
}

type GDScriptExportMemberDecl struct {
	Name string
	Ref  string
}

func genGDScriptCode(outDir string) {
	extensions, err := parseExtensions(protoregistry.GlobalTypes)
	if err != nil {
		log.Panicf("parse proto file failed, %s", err)
	}

	var msgDecls generic.SliceMap[string, protoreflect.MessageType]

	protoregistry.GlobalTypes.RangeMessages(func(msg protoreflect.MessageType) bool {
		isTable := proto.GetExtension(msg.Descriptor().Options(), extensions.IsTable).(bool)
		if !isTable {
			return true
		}
		msgDecls.Add(string(msg.Descriptor().FullName()), msg)
		return true
	})

	type FileImportDecl struct {
		ProtoAlias string
		ExcelAlias string
		BasePath   string
	}

	fileImports := generic.SliceMap[string, FileImportDecl]{}
	exportMembers := make([]GDScriptExportMemberDecl, 0)
	tables := make([]GDScriptTableDecl, 0, msgDecls.Len())

	msgDecls.Each(func(_ string, msg protoreflect.MessageType) {
		filePath := string(msg.Descriptor().ParentFile().Path())
		basePath := strings.TrimSuffix(filepath.ToSlash(filePath), filepath.Ext(filePath))

		importDecl, ok := fileImports.Get(filePath)
		if !ok {
			aliasBase := gdscriptAliasIdentifier(basePath)
			importDecl = FileImportDecl{
				ProtoAlias: aliasBase + "PB",
				ExcelAlias: aliasBase + "Excel",
				BasePath:   basePath,
			}
			fileImports.Add(filePath, importDecl)
		}

		typeName := gdscriptTypeIdentifier(string(msg.Descriptor().Name()))
		tables = append(tables, GDScriptTableDecl{
			Name:             typeName,
			ProtoAlias:       importDecl.ProtoAlias,
			ExcelAlias:       importDecl.ExcelAlias,
			ProtoType:        importDecl.ProtoAlias + "." + typeName,
			ExcelType:        importDecl.ExcelAlias + "." + typeName,
			ChunkedExcelType: importDecl.ExcelAlias + "." + chunkedTableTypeName(typeName),
		})
	})

	protoregistry.GlobalTypes.RangeMessages(func(msg protoreflect.MessageType) bool {
		desc := msg.Descriptor()
		if desc.IsMapEntry() || desc.Parent() != desc.ParentFile() {
			return true
		}

		filePath := string(desc.ParentFile().Path())
		importDecl, ok := fileImports.Get(filePath)
		if !ok {
			return true
		}

		isTable := proto.GetExtension(desc.Options(), extensions.IsTable).(bool)
		if isTable {
			return true
		}

		typeName := gdscriptTypeIdentifier(string(desc.Name()))
		exportMembers = append(exportMembers, GDScriptExportMemberDecl{
			Name: typeName,
			Ref:  importDecl.ProtoAlias + "." + typeName,
		})
		return true
	})

	protoregistry.GlobalTypes.RangeEnums(func(enum protoreflect.EnumType) bool {
		desc := enum.Descriptor()
		if desc.Parent() != desc.ParentFile() {
			return true
		}

		filePath := string(desc.ParentFile().Path())
		importDecl, ok := fileImports.Get(filePath)
		if !ok {
			return true
		}

		enumName := gdscriptTypeIdentifier(string(desc.Name()))
		exportMembers = append(exportMembers, GDScriptExportMemberDecl{
			Name: enumName,
			Ref:  importDecl.ProtoAlias + "." + enumName,
		})
		return true
	})

	const tmpl = `{{.Comment}}

extends Node

# import scripts
{{- range .Imports}}
const {{.ProtoAlias}} = preload("./{{.BasePath}}.pb.gd")
const {{.ExcelAlias}} = preload("./{{.BasePath}}.excel.gd")
{{- end}}

# export classes and enums
{{- range .ExportMembers}}
const {{.Name}} := {{.Ref}}
{{- end}}

# export tables
{{- range .Tables}}
var {{.Name}}: {{.ExcelType}}
{{- end}}

const DEFAULT_DATA_DIR := "{{.DefaultDataDir}}"

var _data_dir := DEFAULT_DATA_DIR

func _init(data_dir: String = DEFAULT_DATA_DIR) -> void:
	_data_dir = data_dir

func _ready() -> void:
	if !_load_data(_data_dir):
		push_error("failed to load excel data, data_dir=%s" % _data_dir)

func _load_data(dir: String) -> bool:	
	var tabs := Tables.load_from_files(dir)
	if tabs == null:
		return false
{{- range .Tables}}
	{{.Name}} = tabs.{{.Name}}
{{- end}}
	return true

class Tables:
{{- range .Tables}}
	var {{.Name}}: {{.ExcelType}}
{{- end}}

	static func load_from_files(dir: String) -> Tables:
		var tabs := Tables.new()
{{range .Tables}}
		var {{.Name}}_msg := {{.ProtoType}}.new()
		var {{.Name}}_path := dir.path_join("{{.Name}}.bin")
		if _load_table_index_file({{.Name}}_msg, {{.Name}}_path):
			tabs.{{.Name}} = {{.ChunkedExcelType}}.new({{.Name}}_msg, {{.Name}}_path)
		elif _load_table_file({{.Name}}_msg, {{.Name}}_path):	
			tabs.{{.Name}} = {{.ExcelType}}.new({{.Name}}_msg)
		else:
			tabs.{{.Name}} = {{.ExcelType}}.new()
{{end}}
		return tabs

	static func _load_table_file(msg: ProtoMessage, path: String) -> bool:
		var start_usec := Time.get_ticks_usec()
		var file := FileAccess.open(path, FileAccess.READ)
		if file == null:
			push_warning("failed to open excel table file, file_path=%s" % path)
			return false
		var stream := ProtoInputFile.new(file)
		msg.reset()
		if !msg.deserialize(stream):
			push_error("failed to deserialize excel table file, file_path=%s" % path)
			return false
		var elapsed_ms := float(Time.get_ticks_usec() - start_usec) / 1000.0
		print("excel table file loaded, file_path=%s, elapsed_ms=%.3f" % [path, elapsed_ms])
		return true

	static func _load_table_index_file(msg, base_path: String) -> bool:
		return _load_table_file(msg, base_path + ".idx")
`

	type TmplArgs struct {
		Comment        string
		DefaultDataDir string
		ExportMembers  []GDScriptExportMemberDecl
		Imports        []FileImportDecl
		Tables         []GDScriptTableDecl
	}

	imports := make([]FileImportDecl, 0, fileImports.Len())
	fileImports.Each(func(_ string, decl FileImportDecl) {
		imports = append(imports, decl)
	})

	args := TmplArgs{
		Comment: fmt.Sprintf(`# Code generated by %[1]s; DO NOT EDIT.
# Command: %[1]s %[2]s
# Note: This file is auto-generated. DO NOT EDIT THIS FILE DIRECTLY.`, strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), strings.Join(os.Args[1:], " ")),
		DefaultDataDir: strings.ReplaceAll(viper.GetString("gdscript_default_data_dir"), `\`, `/`),
		ExportMembers:  pie.SortUsing(exportMembers, func(i, j GDScriptExportMemberDecl) bool { return i.Name < j.Name }),
		Imports:        pie.SortUsing(imports, func(i, j FileImportDecl) bool { return i.BasePath < j.BasePath }),
		Tables:         pie.SortUsing(tables, func(i, j GDScriptTableDecl) bool { return i.Name < j.Name }),
	}

	outFilePath, _ := filepath.Abs(filepath.Join(outDir, "tables.gd"))

	t := template.Must(template.New("code_gdscript").Parse(tmpl))

	os.MkdirAll(outDir, os.ModePerm)

	file, err := os.OpenFile(outFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	if err := t.Execute(file, args); err != nil {
		log.Panic(err)
	}
}

func gdscriptAliasIdentifier(s string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upperNext {
				b.WriteRune(unicode.ToUpper(r))
				upperNext = false
			} else {
				b.WriteRune(r)
			}
			continue
		}
		upperNext = true
	}
	if b.Len() == 0 {
		return "Proto"
	}

	out := b.String()
	if unicode.IsDigit(rune(out[0])) {
		return "Proto" + out
	}
	return out
}

func gdscriptTypeIdentifier(s string) string {
	if s == "" {
		return "_"
	}

	var b strings.Builder
	for i, r := range s {
		if unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			b.WriteRune('_')
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}

	out := b.String()
	if gdscriptKeyword(out) {
		out += "_"
	}
	return out
}

func chunkedTableTypeName(tableName string) string {
	if strings.HasSuffix(tableName, "Table") {
		return strings.TrimSuffix(tableName, "Table") + "ChunkedTable"
	}
	return tableName + "ChunkedTable"
}

func gdscriptKeyword(s string) bool {
	switch s {
	case "and", "as", "assert", "await", "break", "class", "class_name", "const", "continue",
		"elif", "else", "enum", "extends", "false", "for", "func", "if", "in", "is", "match",
		"namespace", "not", "null", "or", "pass", "return", "self", "signal", "static", "super",
		"tool", "true", "var", "while", "yield":
		return true
	default:
		return false
	}
}
