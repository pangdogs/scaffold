package main

import (
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const (
	DependencyProtobuf = "excelc"
)

func genDependencyProtobuf(outDir string) {
	const tmpl = `{{.Comment}}
syntax = 'proto3';

// package
package {{.Package}};

// imports
import "google/protobuf/descriptor.proto";

// options
{{- range .Options}}
{{.}}
{{- end}}

message IndexItem {
	uint64 Value = 1;
	uint32 Offset = 2;
} 

message IndexType {
	enum Enum {
		None = 0;
		UniqueIndex = 1;
		UniqueSortedIndex = 2;
	}
}

extend google.protobuf.MessageOptions {
	optional bool IsColumns = {{Add .CustomOptions 101}};
	optional bool IsTable = {{Add .CustomOptions 102}};
}

extend google.protobuf.FieldOptions {
	optional string Separator = {{Add .CustomOptions 201}};
	optional string FieldAlias = {{Add .CustomOptions 202}};
	optional int32 UniqueIndex = {{Add .CustomOptions 203}};
	optional int32 UniqueSortedIndex = {{Add .CustomOptions 204}};
	optional bool IsRows = {{Add .CustomOptions 205}};
	optional IndexType.Enum IndexTyp = {{Add .CustomOptions 206}};
	optional string IndexFields = {{Add .CustomOptions 207}};
}

extend google.protobuf.EnumValueOptions {
	optional string EnumValueAlias = {{Add .CustomOptions 301}};
}
`

	type TmplArgs struct {
		Comment       string
		Package       string
		Options       []string
		CustomOptions int
	}

	args := TmplArgs{
		Comment: fmt.Sprintf(`// Proto definition generated by %[1]s.
// Command: %[1]s %[2]s
// Note: This file is auto-generated. DO NOT EDIT THIS FILE DIRECTLY.`, strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), strings.Join(os.Args[1:], " ")),
		Package:       viper.GetString("pb_package"),
		CustomOptions: viper.GetInt("pb_custom_options"),
	}

	for k, v := range viper.GetStringMapString("pb_options") {
		args.Options = append(args.Options, fmt.Sprintf("option %s='%s';", k, v))
	}
	sort.Strings(args.Options)

	outFilePath, _ := filepath.Abs(filepath.Join(outDir, fmt.Sprintf("%s.proto", DependencyProtobuf)))

	t := template.Must(template.New("code").
		Funcs(template.FuncMap{
			"Add": func(n ...int) int { return pie.Sum(n) },
		}).
		Parse(tmpl))

	os.MkdirAll(outDir, os.ModePerm)

	outFile, err := os.OpenFile(outFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	err = t.Execute(outFile, args)
	if err != nil {
		panic(err)
	}
}

func genProtobuf(file *excelize.File, globalDecls *generic.SliceMap[Type, *Decl], outDir string) {
	typeDecls := parseTypeDecls(file, globalDecls)

	typeDecls.Each(func(ty Type, decl *Decl) {
		if !globalDecls.TryAdd(ty, decl) {
			panic(fmt.Errorf("读取Excel文件 %q 失败，重复定义类型 %q", file.Path, ty))
		}
	})

	columnDecls := parseTableDecls(file, globalDecls)

	columnDecls.Each(func(ty Type, decl *Decl) {
		if !globalDecls.TryAdd(ty, decl) {
			panic(fmt.Errorf("读取Excel文件 %q 失败，重复定义类型 %q", file.Path, ty))
		}
	})

	if typeDecls.Len() <= 0 && columnDecls.Len() <= 0 {
		return
	}

	const tmpl = `{{.Comment}}
syntax = 'proto3';

// package
package {{.Package}};

// imports
{{- range .Imports}}
{{.}}
{{- end}}

// options
{{- range .Options}}
{{.}}
{{- end}}

// enums
{{- range .Enums}}
message {{.Type}} {
	enum Enum {
		{{- range .EnumFields}}
		{{.K}} = {{.V.Value}}{{- .V.ProtobufMeta -}}; // {{.V.Alias}} - {{.V.Comment}}
		{{- end}}
	}
}
{{end}}

// structs
{{- range .Structs}}
message {{.Type}} {
	{{- range $i, $kv := .StructFields}}
	{{$kv.V.ProtobufType}} {{$kv.K}} = {{Incr $i}}{{- $kv.V.ProtobufMeta -}}; // {{.V.Alias}} - {{.V.Comment}}
	{{- end}}
}
{{end}}

// table and columns
{{- $package := .Package -}}
{{- range .Columns}}
message {{.ProtobufType}} {
	option ({{$package}}.IsColumns) = true;
	{{- range $i, $kv := .StructFields}}
	{{$kv.V.ProtobufType}} {{$kv.K}} = {{Incr $i}}{{- $kv.V.ProtobufMeta -}}; // {{.V.Alias}} - {{.V.Comment}}
	{{- end}}
}

message {{TableName .ProtobufType}} {
	option ({{$package}}.IsTable) = true;
	repeated {{.ProtobufType}} Rows = 1 [({{$package}}.IsRows) = true];
  	{{- $StructUniqueIndexesCount := .StructUniqueIndexes.Len -}}
	{{- range $i, $kv := .StructUniqueIndexes}}
	map<uint64, uint32> UniqueIndex{{$kv.K}} = {{Add $i 2}} [({{$package}}.IndexTyp) = UniqueIndex, ({{$package}}.IndexFields) = '{{$kv.V}}']; 
	{{- end}}
	{{- $StructUniqueSortedIndexesCount := .StructUniqueSortedIndexes.Len -}}
	{{- range $i, $kv := .StructUniqueSortedIndexes}}
	repeated IndexItem UniqueSortedIndex{{$kv.K}} = {{Add $i 2 $StructUniqueIndexesCount}} [({{$package}}.IndexTyp) = UniqueSortedIndex, ({{$package}}.IndexFields) = '{{$kv.V}}'];
	{{- end}}
}
{{end}}
`

	type TmplArgs struct {
		Comment string
		Package string
		Imports []string
		Options []string
		Enums   []*Decl
		Structs []*Decl
		Columns []*Decl
	}

	args := TmplArgs{
		Comment: fmt.Sprintf(`// Proto definition generated by %[1]s.
// Command: %[1]s %[2]s
// Excel File: %[3]s
// Note: This file is auto-generated. DO NOT EDIT THIS FILE DIRECTLY.`, strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])), strings.Join(os.Args[1:], " "), file.Path),
		Package: viper.GetString("pb_package"),
		Imports: []string{fmt.Sprintf("import '%s.proto';", DependencyProtobuf)},
	}

	outFilePath, _ := filepath.Abs(filepath.Join(outDir, strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))+".proto"))

	for _, v := range viper.GetStringSlice("pb_imports") {
		impFilePath := v

		if !filepath.IsAbs(v) {
			impFilePath, _ = filepath.Abs(filepath.Join(outDir, v))
		}

		if impFilePath == outFilePath {
			continue
		}

		args.Imports = append(args.Imports, fmt.Sprintf("import '%s';", v))
	}
	sort.Strings(args.Imports)

	for k, v := range viper.GetStringMapString("pb_options") {
		args.Options = append(args.Options, fmt.Sprintf("option %s='%s';", k, v))
	}
	sort.Strings(args.Options)

	typeDecls.Each(func(ty Type, decl *Decl) {
		if decl.IsEnum {
			args.Enums = append(args.Enums, decl)
		}
		if decl.IsStruct {
			args.Structs = append(args.Structs, decl)
		}
	})

	columnDecls.Each(func(ty Type, decl *Decl) {
		if decl.IsTable {
			args.Columns = append(args.Columns, decl)
		}
	})

	t := template.Must(template.New("code").
		Funcs(template.FuncMap{
			"Incr":      func(n int) int { return n + 1 },
			"Add":       func(n ...int) int { return pie.Sum(n) },
			"TableName": func(s string) string { return strings.TrimSuffix(s, "Columns") + "Table" },
		}).
		Parse(tmpl))

	os.MkdirAll(outDir, os.ModePerm)

	outFile, err := os.OpenFile(outFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	err = t.Execute(outFile, args)
	if err != nil {
		panic(err)
	}
}
