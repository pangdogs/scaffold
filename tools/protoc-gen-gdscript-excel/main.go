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
	"flag"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	indexTypeHashUnique   = "HashUniqueIndex"
	indexTypeSortedUnique = "SortedUniqueIndex"
)

type TableDecl struct {
	Message            *protogen.Message
	RowsField          *protogen.Field
	RowsMessage        *protogen.Message
	ChunkManifestField *protogen.Field
	ChunkListField     *protogen.Field
	ChunkOffsetField   *protogen.Field
	ChunkCountField    *protogen.Field
	IndexMethods       []IndexMethodDecl
}

type IndexMethodDecl struct {
	Field            *protogen.Field
	IndexTypeName    string
	IndexFields      []*protogen.Field
	LookupMethodName string
}

type ProtoDescriptors interface {
	Enums() protoreflect.EnumDescriptors
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}

type Extensions struct {
	IsTable,
	IndexType,
	IndexFields protoreflect.ExtensionType
}

type GeneratorConfig struct {
	StringAsStringName bool
}

var config GeneratorConfig

func main() {
	var flags flag.FlagSet
	stringAsStringName := flags.Bool("string_as_stringname", false, "map proto string fields to GDScript StringName")

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		config = GeneratorConfig{
			StringAsStringName: *stringAsStringName,
		}
		for _, f := range gen.Files {
			if err := registerProtoTypes(protoregistry.GlobalTypes, f.Desc); err != nil {
				return err
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
	ext, err := parseExtensions(file)
	if err != nil {
		return err
	}

	tables, err := collectTables(file, ext)
	if err != nil {
		return err
	}
	if len(tables) <= 0 {
		return nil
	}

	g := gen.NewGeneratedFile(file.GeneratedFilenamePrefix+".excel.gd", "")
	emitGeneratedHeader(gen, file, g)

	protoImportAlias := importAliasIdentifier(path.Base(file.GeneratedFilenamePrefix))
	messageTypeNames := collectMessageTypeNames(file)
	g.P("const ", protoImportAlias, " = preload(", strconv.Quote("./"+path.Base(file.GeneratedFilenamePrefix)+".pb.gd"), ")")
	g.P()

	for _, table := range tables {
		if err := emitTableWrapper(g, table, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}

	return nil
}

func registerProtoTypes(pbTypes *protoregistry.Types, desc ProtoDescriptors) error {
	for i := 0; i < desc.Extensions().Len(); i++ {
		ext := desc.Extensions().Get(i)
		_, err := pbTypes.FindExtensionByName(ext.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}
		if err := pbTypes.RegisterExtension(dynamicpb.NewExtensionType(ext)); err != nil {
			return err
		}
	}

	for i := 0; i < desc.Enums().Len(); i++ {
		enum := desc.Enums().Get(i)
		_, err := pbTypes.FindEnumByName(enum.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}
		if err := pbTypes.RegisterEnum(dynamicpb.NewEnumType(enum)); err != nil {
			return err
		}
	}

	for i := 0; i < desc.Messages().Len(); i++ {
		msg := desc.Messages().Get(i)
		_, err := pbTypes.FindMessageByName(msg.FullName())
		if !errors.Is(err, protoregistry.NotFound) {
			continue
		}
		if err := pbTypes.RegisterMessage(dynamicpb.NewMessageType(msg)); err != nil {
			return err
		}
		if err := registerProtoTypes(pbTypes, msg); err != nil {
			return err
		}
	}

	return nil
}

func parseExtensions(file *protogen.File) (*Extensions, error) {
	result := &Extensions{}

	var err error
	result.IsTable, err = findExtension(file, "IsTable")
	if err != nil {
		return nil, err
	}
	result.IndexType, err = findExtension(file, "IndexType_")
	if err != nil {
		return nil, err
	}
	result.IndexFields, err = findExtension(file, "IndexFields")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func findExtension(file *protogen.File, name string) (protoreflect.ExtensionType, error) {
	extName := protoFullName(file, name)
	ext, err := protoregistry.GlobalTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find extension %q: %w", extName, err)
	}
	return ext, nil
}

func collectTables(file *protogen.File, ext *Extensions) ([]TableDecl, error) {
	indexType, err := findIndexTypeEnum(file)
	if err != nil {
		return nil, err
	}

	var tables []TableDecl
	for i, msg := range file.Messages {
		pbMsg := file.Proto.MessageType[i]
		isTable := proto.GetExtension(pbMsg.Options, ext.IsTable).(bool)
		if !isTable {
			continue
		}

		rowsFieldIndex := slices.IndexFunc(msg.Fields, func(field *protogen.Field) bool {
			return string(field.Desc.Name()) == "Rows"
		})
		if rowsFieldIndex < 0 {
			return nil, fmt.Errorf("table %s must declare exactly one rows field", msg.Desc.FullName())
		}

		rowsField := msg.Fields[rowsFieldIndex]
		if rowsField.Message == nil {
			return nil, fmt.Errorf("rows field %q in %s must be a message", rowsField.Desc.Name(), msg.Desc.FullName())
		}

		chunkManifestField := fieldByName(msg.Fields, "ChunkManifest")
		if chunkManifestField == nil {
			return nil, fmt.Errorf("table %s must declare chunk manifest field %q", msg.Desc.FullName(), "ChunkManifest")
		}
		if chunkManifestField.Message == nil {
			return nil, fmt.Errorf("chunk manifest field %q in %s must be a message", chunkManifestField.Desc.Name(), msg.Desc.FullName())
		}

		chunkListField := fieldByName(chunkManifestField.Message.Fields, "Chunks")
		if chunkListField == nil {
			return nil, fmt.Errorf("chunk manifest %s must declare field %q", chunkManifestField.Message.Desc.FullName(), "Chunks")
		}
		if chunkListField.Message == nil {
			return nil, fmt.Errorf("chunk list field %q in %s must be a repeated message", chunkListField.Desc.Name(), chunkManifestField.Message.Desc.FullName())
		}

		chunkOffsetField := fieldByName(chunkListField.Message.Fields, "Offset")
		if chunkOffsetField == nil {
			return nil, fmt.Errorf("chunk message %s must declare field %q", chunkListField.Message.Desc.FullName(), "Offset")
		}

		chunkCountField := fieldByName(chunkListField.Message.Fields, "Count")
		if chunkCountField == nil {
			return nil, fmt.Errorf("chunk message %s must declare field %q", chunkListField.Message.Desc.FullName(), "Count")
		}

		table := TableDecl{
			Message:            msg,
			RowsField:          rowsField,
			RowsMessage:        rowsField.Message,
			ChunkManifestField: chunkManifestField,
			ChunkListField:     chunkListField,
			ChunkOffsetField:   chunkOffsetField,
			ChunkCountField:    chunkCountField,
		}

		indexMethods, err := collectIndexMethods(msg, pbMsg, table.RowsMessage, ext, indexType)
		if err != nil {
			return nil, err
		}
		table.IndexMethods = indexMethods

		tables = append(tables, table)
	}

	return tables, nil
}

func fieldByName(fields []*protogen.Field, name protoreflect.Name) *protogen.Field {
	for _, field := range fields {
		if field.Desc.Name() == name {
			return field
		}
	}
	return nil
}

func findIndexTypeEnum(file *protogen.File) (protoreflect.EnumType, error) {
	indexTypeName := protoFullName(file, "IndexType.Enum")
	indexType, err := protoregistry.GlobalTypes.FindEnumByName(indexTypeName)
	if err != nil {
		return nil, fmt.Errorf("find enum %q: %w", indexTypeName, err)
	}
	return indexType, nil
}

func protoFullName(file *protogen.File, suffix string) protoreflect.FullName {
	if pkg := string(file.Desc.Package()); pkg != "" {
		return protoreflect.FullName(pkg + "." + suffix)
	}
	return protoreflect.FullName(suffix)
}

func collectIndexMethods(msg *protogen.Message, pbMsg *descriptorpb.DescriptorProto, rowsMessage *protogen.Message, ext *Extensions, indexType protoreflect.EnumType) ([]IndexMethodDecl, error) {
	var methods []IndexMethodDecl

	for fieldIndex, field := range msg.Fields {
		indexTypeName, err := resolveIndexTypeName(pbMsg.Field[fieldIndex], ext, indexType)
		if err != nil {
			return nil, err
		}
		if indexTypeName == "" {
			continue
		}

		indexFieldsValue := proto.GetExtension(pbMsg.Field[fieldIndex].Options, ext.IndexFields).(string)
		indexFields, err := resolveIndexFields(rowsMessage, indexFieldsValue)
		if err != nil {
			return nil, err
		}

		methodSuffix, err := buildIndexMethodSuffix(indexTypeName, indexFields)
		if err != nil {
			return nil, err
		}

		methods = append(methods, IndexMethodDecl{
			Field:            field,
			IndexTypeName:    indexTypeName,
			IndexFields:      indexFields,
			LookupMethodName: "lookup_by_" + methodSuffix,
		})
	}

	return methods, nil
}

func resolveIndexTypeName(field *descriptorpb.FieldDescriptorProto, ext *Extensions, indexType protoreflect.EnumType) (string, error) {
	indexTypeValue, ok := proto.GetExtension(field.Options, ext.IndexType).(protoreflect.EnumNumber)
	if !ok || indexTypeValue <= 0 {
		return "", nil
	}

	indexTypeValueDesc := indexType.Descriptor().Values().ByNumber(indexTypeValue)
	if indexTypeValueDesc == nil {
		return "", fmt.Errorf("field %q has unknown index type value %d", field.GetName(), indexTypeValue)
	}

	return string(indexTypeValueDesc.Name()), nil
}

func resolveIndexFields(rowsMessage *protogen.Message, indexFieldsValue string) ([]*protogen.Field, error) {
	var fields []*protogen.Field
	for _, indexFieldName := range strings.Split(indexFieldsValue, ",") {
		indexFieldName = strings.TrimSpace(indexFieldName)
		if indexFieldName == "" {
			continue
		}

		index := slices.IndexFunc(rowsMessage.Fields, func(field *protogen.Field) bool {
			return string(field.Desc.Name()) == indexFieldName
		})
		if index < 0 {
			return nil, fmt.Errorf("index field %q not found in %s", indexFieldName, rowsMessage.Desc.FullName())
		}

		fields = append(fields, rowsMessage.Fields[index])
	}
	if len(fields) <= 0 {
		return nil, fmt.Errorf("index fields for %s cannot be empty", rowsMessage.Desc.FullName())
	}
	return fields, nil
}

func buildIndexMethodSuffix(indexTypeName string, indexFields []*protogen.Field) (string, error) {
	parts := make([]string, 0, len(indexFields))
	for _, indexField := range indexFields {
		name := safeIdentifier(string(indexField.Desc.Name()))
		if name == "" {
			return "", fmt.Errorf("index field %q produced an empty method suffix segment", indexField.Desc.FullName())
		}
		parts = append(parts, name)
	}
	if len(parts) <= 0 {
		return "", fmt.Errorf("index fields cannot be empty")
	}

	suffix := strings.Join(parts, "")
	switch indexTypeName {
	case indexTypeHashUnique:
		return "hash_unique_index_" + suffix, nil
	case indexTypeSortedUnique:
		return "sorted_unique_index_" + suffix, nil
	default:
		return "", fmt.Errorf("unsupported index type %q", indexTypeName)
	}
}

func collectMessageTypeNames(file *protogen.File) map[protoreflect.FullName]string {
	names := make(map[protoreflect.FullName]string)
	var walk func(*protogen.Message)
	walk = func(msg *protogen.Message) {
		if msg.Desc.IsMapEntry() {
			return
		}
		names[msg.Desc.FullName()] = safeIdentifier(msg.GoIdent.GoName)
		for _, nested := range msg.Messages {
			walk(nested)
		}
	}
	for _, msg := range file.Messages {
		walk(msg)
	}
	return names
}

func emitGeneratedHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("# Code generated by protoc-gen-gdscript-excel. DO NOT EDIT.")
	if v := gen.Request.GetCompilerVersion(); v != nil {
		protocVersion := fmt.Sprintf("v%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
		if suffix := v.GetSuffix(); suffix != "" {
			protocVersion += "-" + suffix
		}
		g.P("# protoc ", protocVersion)
	}
	g.P("# source: ", file.Desc.Path())
	g.P()
	g.P("extends RefCounted")
	g.P()
}

func emitTableWrapper(g *protogen.GeneratedFile, table TableDecl, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	tableName := safeIdentifier(table.Message.GoIdent.GoName)
	chunkedTableName := chunkedTableTypeName(tableName)
	rowType := protoImportAlias + "." + safeIdentifier(table.RowsMessage.GoIdent.GoName)
	protoTableType := protoImportAlias + "." + tableName
	rowFieldName := safeIdentifier(table.RowsField.GoName)
	chunkManifestFieldName := safeIdentifier(table.ChunkManifestField.GoName)
	chunksFieldName := safeIdentifier(table.ChunkListField.GoName)
	chunkOffsetFieldName := safeIdentifier(table.ChunkOffsetField.GoName)
	chunkCountFieldName := safeIdentifier(table.ChunkCountField.GoName)

	g.P("class ", tableName, ":")
	g.P("\tvar _msg: ", protoTableType)
	g.P()
	g.P("\tfunc _init(msg: ", protoTableType, " = null) -> void:")
	g.P("\t\t_msg = msg if msg != null else ", protoTableType, ".new()")
	g.P()
	g.P("\tfunc rows() -> Array[", rowType, "]:")
	g.P("\t\treturn _msg.", rowFieldName, ".duplicate()")
	g.P()
	g.P("\tfunc rows_async() -> Array[", rowType, "]:")
	g.P("\t\treturn rows()")
	g.P()
	g.P("\tfunc row_count() -> int:")
	g.P("\t\treturn _msg.", rowFieldName, ".size()")
	g.P()
	g.P("\tfunc row_at(offset: int) -> ", rowType, ":")
	g.P("\t\tif offset < 0 or offset >= _msg.", rowFieldName, ".size():")
	g.P("\t\t\treturn null")
	g.P("\t\treturn _msg.", rowFieldName, "[offset]")
	g.P()
	g.P("\tfunc row_at_async(offset: int) -> ", rowType, ":")
	g.P("\t\treturn row_at(offset)")
	g.P()

	if len(table.IndexMethods) > 0 {
		if err := emitDefaultLookupMethod(g, table.IndexMethods[0], rowType, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}

	for _, method := range table.IndexMethods {
		if err := emitLookupMethod(g, table, method, rowType, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}
	for _, method := range table.IndexMethods {
		if requiresHashIndex(method.IndexFields) {
			if err := emitIndexHashMethod(g, method, protoImportAlias, messageTypeNames); err != nil {
				return err
			}
			if err := emitIndexMatchMethod(g, method, rowType, protoImportAlias, messageTypeNames); err != nil {
				return err
			}
		}
	}

	g.P("class ", chunkedTableName, " extends ", tableName, ":")
	g.P("\tvar _chunk_loader: ExcelUtils.ChunkLoader")
	g.P()
	g.P("\tfunc _init(msg: ", protoTableType, ", chunk_base_path: String) -> void:")
	g.P("\t\tsuper(msg)")
	g.P("\t\t_chunk_loader = ExcelUtils.ChunkLoader.new(")
	g.P("\t\t\tchunk_base_path,")
	g.P("\t\t\t_msg.", chunkManifestFieldName, ".", chunksFieldName, ".size() if _msg.", chunkManifestFieldName, " != null else 0,")
	g.P("\t\t\tfunc(): return ", protoTableType, ".new()")
	g.P("\t\t)")
	g.P()
	emitChunkedPublicMethods(
		g,
		rowType,
		chunkManifestFieldName,
		chunksFieldName,
		chunkOffsetFieldName,
		chunkCountFieldName,
	)
	if len(table.IndexMethods) > 0 {
		if err := emitChunkedAsyncDefaultLookupMethod(g, table.IndexMethods[0], rowType, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}
	for _, method := range table.IndexMethods {
		if err := emitChunkedAsyncLookupMethod(g, method, rowType, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}
	emitChunkLoaderInternalMethods(
		g,
		protoTableType,
		rowType,
		rowFieldName,
		chunkManifestFieldName,
		chunksFieldName,
		chunkOffsetFieldName,
		chunkCountFieldName,
	)

	g.P()
	return nil
}

func chunkedTableTypeName(tableName string) string {
	if strings.HasSuffix(tableName, "Table") {
		return strings.TrimSuffix(tableName, "Table") + "ChunkedTable"
	}
	return tableName + "ChunkedTable"
}

func emitDefaultLookupMethod(g *protogen.GeneratedFile, method IndexMethodDecl, rowType, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	argNames := gdscriptArgumentNames(method.IndexFields)

	g.P("\tfunc lookup(", argList, ") -> ", rowType, ":")
	g.P("\t\treturn ", method.LookupMethodName, "(", argNames, ")")
	g.P()
	g.P("\tfunc lookup_async(", argList, ") -> ", rowType, ":")
	g.P("\t\treturn lookup(", argNames, ")")
	g.P()
	return nil
}

func emitLookupMethod(g *protogen.GeneratedFile, table TableDecl, method IndexMethodDecl, rowType, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	argNames := gdscriptArgumentNames(method.IndexFields)
	indexFieldName := safeIdentifier(method.Field.GoName)

	g.P("\tfunc ", method.LookupMethodName, "(", argList, ") -> ", rowType, ":")
	g.P("\t\tif row_count() <= 0:")
	g.P("\t\t\treturn null")

	if len(method.IndexFields) == 1 && supportsDirectIndex(method.IndexFields[0]) {
		idxExpr, err := directIndexExpression(gdscriptArgumentName(method.IndexFields[0]), method.IndexFields[0])
		if err != nil {
			return err
		}
		g.P("\t\tvar idx := ", idxExpr)
	} else {
		g.P("\t\tvar idx := ", indexHashMethodName(method), "(", argNames, ")")
	}

	if err := emitLookupOffset(g, method, indexFieldName); err != nil {
		return err
	}

	g.P("\t\tif offset < 0 or offset >= row_count():")
	g.P("\t\t\treturn null")
	g.P("\t\tvar row := row_at(offset)")
	g.P("\t\tif row == null:")
	g.P("\t\t\treturn null")
	if requiresHashIndex(method.IndexFields) {
		g.P("\t\tif !", indexMatchMethodName(method), "(row, ", argNames, "):")
		g.P("\t\t\tvar bucket = _msg.", indexFieldName, "Conflict.get(idx)")
		g.P("\t\t\tif bucket == null:")
		g.P("\t\t\t\treturn null")
		g.P("\t\t\tfor conflict_offset in bucket.Offsets:")
		g.P("\t\t\t\trow = row_at(conflict_offset)")
		g.P("\t\t\t\tif row == null:")
		g.P("\t\t\t\t\tcontinue")
		g.P("\t\t\t\tif ", indexMatchMethodName(method), "(row, ", argNames, "):")
		g.P("\t\t\t\t\treturn row")
		g.P("\t\t\treturn null")
	}
	g.P("\t\treturn row")
	g.P()
	g.P("\tfunc ", method.LookupMethodName, "_async(", argList, ") -> ", rowType, ":")
	g.P("\t\treturn ", method.LookupMethodName, "(", argNames, ")")
	g.P()
	return nil
}

func emitChunkedAsyncDefaultLookupMethod(g *protogen.GeneratedFile, method IndexMethodDecl, rowType, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	argNames := gdscriptArgumentNames(method.IndexFields)

	g.P("\tfunc lookup_async(", argList, ") -> ", rowType, ":")
	g.P("\t\treturn await ", method.LookupMethodName, "_async(", argNames, ")")
	g.P()
	return nil
}

func emitChunkedAsyncLookupMethod(g *protogen.GeneratedFile, method IndexMethodDecl, rowType, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	argNames := gdscriptArgumentNames(method.IndexFields)
	indexFieldName := safeIdentifier(method.Field.GoName)

	g.P("\tfunc ", method.LookupMethodName, "_async(", argList, ") -> ", rowType, ":")
	g.P("\t\tif row_count() <= 0:")
	g.P("\t\t\treturn null")

	if len(method.IndexFields) == 1 && supportsDirectIndex(method.IndexFields[0]) {
		idxExpr, err := directIndexExpression(gdscriptArgumentName(method.IndexFields[0]), method.IndexFields[0])
		if err != nil {
			return err
		}
		g.P("\t\tvar idx := ", idxExpr)
	} else {
		g.P("\t\tvar idx := ", indexHashMethodName(method), "(", argNames, ")")
	}

	if err := emitLookupOffset(g, method, indexFieldName); err != nil {
		return err
	}

	g.P("\t\tif offset < 0 or offset >= row_count():")
	g.P("\t\t\treturn null")
	g.P("\t\tvar row := await row_at_async(offset)")
	g.P("\t\tif row == null:")
	g.P("\t\t\treturn null")
	if requiresHashIndex(method.IndexFields) {
		g.P("\t\tif !", indexMatchMethodName(method), "(row, ", argNames, "):")
		g.P("\t\t\tvar bucket = _msg.", indexFieldName, "Conflict.get(idx)")
		g.P("\t\t\tif bucket == null:")
		g.P("\t\t\t\treturn null")
		g.P("\t\t\tfor conflict_offset in bucket.Offsets:")
		g.P("\t\t\t\trow = await row_at_async(conflict_offset)")
		g.P("\t\t\t\tif row == null:")
		g.P("\t\t\t\t\tcontinue")
		g.P("\t\t\t\tif ", indexMatchMethodName(method), "(row, ", argNames, "):")
		g.P("\t\t\t\t\treturn row")
		g.P("\t\t\treturn null")
	}
	g.P("\t\treturn row")
	g.P()
	return nil
}

func emitChunkedPublicMethods(g *protogen.GeneratedFile, rowType, chunkManifestFieldName, chunksFieldName, chunkOffsetFieldName, chunkCountFieldName string) {
	g.P("\tfunc rows() -> Array[", rowType, "]:")
	g.P("\t\tif !_ensure_all_rows_loaded():")
	g.P("\t\t\treturn []")
	g.P("\t\treturn _build_rows_array()")
	g.P()
	g.P("\tfunc rows_async() -> Array[", rowType, "]:")
	g.P("\t\tif !await _ensure_all_rows_loaded_async():")
	g.P("\t\t\treturn []")
	g.P("\t\treturn _build_rows_array()")
	g.P()
	g.P("\tfunc row_count() -> int:")
	g.P("\t\tvar chunks = _msg.", chunkManifestFieldName, ".", chunksFieldName)
	g.P("\t\tif chunks.is_empty():")
	g.P("\t\t\treturn 0")
	g.P("\t\treturn chunks[chunks.size() - 1].", chunkOffsetFieldName, " + chunks[chunks.size() - 1].", chunkCountFieldName)
	g.P()
	g.P("\tfunc row_at(offset: int) -> ", rowType, ":")
	g.P("\t\tif offset < 0 or offset >= row_count():")
	g.P("\t\t\treturn null")
	g.P("\t\tvar chunk_index := _chunk_index_for_offset(offset)")
	g.P("\t\tif !_chunk_loader.ensure_loaded(chunk_index):")
	g.P("\t\t\treturn null")
	g.P("\t\tvar chunk_rows: Array[", rowType, "] = _chunk_loader.rows(chunk_index)")
	g.P("\t\tvar row_offset := offset - _msg.", chunkManifestFieldName, ".", chunksFieldName, "[chunk_index].", chunkOffsetFieldName)
	g.P("\t\tif row_offset < 0 or row_offset >= chunk_rows.size():")
	g.P("\t\t\treturn null")
	g.P("\t\treturn chunk_rows[row_offset]")
	g.P()
	g.P("\tfunc row_at_async(offset: int) -> ", rowType, ":")
	g.P("\t\tif offset < 0 or offset >= row_count():")
	g.P("\t\t\treturn null")
	g.P("\t\tvar chunk_index := _chunk_index_for_offset(offset)")
	g.P("\t\tif !await _chunk_loader.ensure_loaded_async(chunk_index):")
	g.P("\t\t\treturn null")
	g.P("\t\tvar chunk_rows: Array[", rowType, "] = _chunk_loader.rows(chunk_index)")
	g.P("\t\tvar row_offset := offset - _msg.", chunkManifestFieldName, ".", chunksFieldName, "[chunk_index].", chunkOffsetFieldName)
	g.P("\t\tif row_offset < 0 or row_offset >= chunk_rows.size():")
	g.P("\t\t\treturn null")
	g.P("\t\treturn chunk_rows[row_offset]")
	g.P()
}

func emitChunkLoaderInternalMethods(g *protogen.GeneratedFile, protoTableType, rowType, rowFieldName, chunkManifestFieldName, chunksFieldName, chunkOffsetFieldName, chunkCountFieldName string) {
	g.P("\tfunc _chunk_index_for_offset(offset: int) -> int:")
	g.P("\t\tvar chunks = _msg.", chunkManifestFieldName, ".", chunksFieldName)
	g.P("\t\tvar low := 0")
	g.P("\t\tvar high := chunks.size() - 1")
	g.P("\t\twhile low <= high:")
	g.P("\t\t\t@warning_ignore(\"integer_division\")")
	g.P("\t\t\tvar mid := low + int((high - low) / 2)")
	g.P("\t\t\tif offset < chunks[mid].", chunkOffsetFieldName, ":")
	g.P("\t\t\t\thigh = mid - 1")
	g.P("\t\t\telif offset >= chunks[mid].", chunkOffsetFieldName, " + chunks[mid].", chunkCountFieldName, ":")
	g.P("\t\t\t\tlow = mid + 1")
	g.P("\t\t\telse:")
	g.P("\t\t\t\treturn mid")
	g.P("\t\treturn -1")
	g.P()
	g.P("\tfunc _ensure_all_rows_loaded() -> bool:")
	g.P("\t\tfor chunk_index in range(_msg.", chunkManifestFieldName, ".", chunksFieldName, ".size()):")
	g.P("\t\t\tif !_chunk_loader.ensure_loaded(chunk_index):")
	g.P("\t\t\t\treturn false")
	g.P("\t\treturn true")
	g.P()
	g.P("\tfunc _ensure_all_rows_loaded_async() -> bool:")
	g.P("\t\tfor chunk_index in range(_msg.", chunkManifestFieldName, ".", chunksFieldName, ".size()):")
	g.P("\t\t\tif !await _chunk_loader.ensure_loaded_async(chunk_index):")
	g.P("\t\t\t\treturn false")
	g.P("\t\treturn true")
	g.P()
	g.P("\tfunc _build_rows_array() -> Array[", rowType, "]:")
	g.P("\t\t@warning_ignore(\"shadowed_variable\")")
	g.P("\t\tvar rows: Array[", rowType, "] = []")
	g.P("\t\trows.resize(row_count())")
	g.P("\t\tfor chunk_index in range(_msg.", chunkManifestFieldName, ".", chunksFieldName, ".size()):")
	g.P("\t\t\tvar chunk_rows: Array[", rowType, "] = _chunk_loader.rows(chunk_index)")
	g.P("\t\t\tfor row_offset in range(_msg.", chunkManifestFieldName, ".", chunksFieldName, "[chunk_index].", chunkCountFieldName, "):")
	g.P("\t\t\t\trows[_msg.", chunkManifestFieldName, ".", chunksFieldName, "[chunk_index].", chunkOffsetFieldName, " + row_offset] = chunk_rows[row_offset]")
	g.P("\t\treturn rows")
	g.P()
}

func emitLookupOffset(g *protogen.GeneratedFile, method IndexMethodDecl, indexFieldName string) error {
	switch method.IndexTypeName {
	case indexTypeHashUnique:
		g.P("\t\tvar offset = _msg.", indexFieldName, ".get(idx)")
		g.P("\t\tif offset == null:")
		g.P("\t\t\treturn null")
	case indexTypeSortedUnique:
		g.P("\t\tif _msg.", indexFieldName, " == null:")
		g.P("\t\t\treturn null")
		g.P("\t\tvar idx_offset := ExcelUtils.binary_search_u64(_msg.", indexFieldName, ".Values, idx)")
		g.P("\t\tif idx_offset < 0:")
		g.P("\t\t\treturn null")
		g.P("\t\tvar offset := _msg.", indexFieldName, ".Offsets[idx_offset]")
	default:
		return fmt.Errorf("unsupported index type %q", method.IndexTypeName)
	}
	return nil
}

func emitIndexHashMethod(g *protogen.GeneratedFile, method IndexMethodDecl, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	g.P("\tfunc ", indexHashMethodName(method), "(", argList, ") -> int:")
	g.P("\t\tvar hasher := ProtoUtils.new_hasher()")
	for _, field := range method.IndexFields {
		if err := emitHashStatements(g, "\t\t", "hasher", gdscriptArgumentName(field), field, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}
	g.P("\t\treturn hasher.sum64()")
	g.P()
	return nil
}

func emitIndexMatchMethod(g *protogen.GeneratedFile, method IndexMethodDecl, rowType, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	argList := "row: " + rowType
	args, err := gdscriptArgumentList(method.IndexFields, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	if args != "" {
		argList += ", " + args
	}

	g.P("\tfunc ", indexMatchMethodName(method), "(", argList, ") -> bool:")
	for _, field := range method.IndexFields {
		if err := emitEqualityStatements(g, "\t\t", gdscriptArgumentName(field), "row."+safeIdentifier(field.GoName), field, protoImportAlias, messageTypeNames); err != nil {
			return err
		}
	}
	g.P("\t\treturn true")
	g.P()
	return nil
}

func emitHashStatements(g *protogen.GeneratedFile, indent, hasherName, valueExpr string, field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	if field.Desc.IsMap() {
		if field.Message == nil || len(field.Message.Fields) < 2 {
			return fmt.Errorf("map field %s has invalid entry descriptor", field.Desc.FullName())
		}
		keyField := field.Message.Fields[0]
		keyHasher, err := hashCallExpression(hasherName, "key", keyField, protoImportAlias, messageTypeNames)
		if err != nil {
			return err
		}
		valueHasher, err := hashCallExpression(hasherName, "value", field.Message.Fields[1], protoImportAlias, messageTypeNames)
		if err != nil {
			return err
		}
		g.P(indent, "ProtoUtils.hash_dictionary(", hasherName, ", ", valueExpr, ", func(key): ", keyHasher, ", func(value): ", valueHasher, hashDictionaryKeyOrderSuffix(keyField), ")")
		return nil
	}

	if field.Desc.IsList() {
		valueHasher, err := hashCallExpression(hasherName, "value", field, protoImportAlias, messageTypeNames)
		if err != nil {
			return err
		}
		g.P(indent, "ProtoUtils.hash_array(", hasherName, ", ", valueExpr, ", func(value): ", valueHasher, ")")
		return nil
	}

	callExpr, err := hashCallExpression(hasherName, valueExpr, field, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	g.P(indent, callExpr)
	return nil
}

func emitEqualityStatements(g *protogen.GeneratedFile, indent, leftExpr, rightExpr string, field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	if field.Desc.IsMap() {
		if field.Message == nil || len(field.Message.Fields) < 2 {
			return fmt.Errorf("map field %s has invalid entry descriptor", field.Desc.FullName())
		}
		valueEqualExpr, err := equalCallExpression("left", "right", field.Message.Fields[1], protoImportAlias, messageTypeNames)
		if err != nil {
			return err
		}
		g.P(indent, "if !ProtoUtils.equal_dictionary(", leftExpr, ", ", rightExpr, ", func(left, right): return ", valueEqualExpr, "):")
		g.P(indent, "\treturn false")
		return nil
	}

	if field.Desc.IsList() {
		valueEqualExpr, err := equalCallExpression("left", "right", field, protoImportAlias, messageTypeNames)
		if err != nil {
			return err
		}
		g.P(indent, "if !ProtoUtils.equal_array(", leftExpr, ", ", rightExpr, ", func(left, right): return ", valueEqualExpr, "):")
		g.P(indent, "\treturn false")
		return nil
	}

	return emitEqualityValueStatements(g, indent, leftExpr, rightExpr, field, protoImportAlias, messageTypeNames)
}

func emitEqualityValueStatements(g *protogen.GeneratedFile, indent, leftExpr, rightExpr string, field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) error {
	equalExpr, err := equalCallExpression(leftExpr, rightExpr, field, protoImportAlias, messageTypeNames)
	if err != nil {
		return err
	}
	g.P(indent, "if !(", equalExpr, "):")
	g.P(indent, "\treturn false")
	return nil
}

func hashCallExpression(hasherName, valueExpr string, field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) (string, error) {
	if field.Message != nil && !field.Desc.IsMap() {
		typeExpr, err := gdscriptSingularTypeExpression(field, protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		return "ProtoUtils.hash_message(" + hasherName + ", " + valueExpr + ", func(): return " + typeExpr + ".new())", nil
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ProtoUtils.hash_bool(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.StringKind:
		return "ProtoUtils.hash_string(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.BytesKind:
		return "ProtoUtils.hash_bytes(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.FloatKind:
		return "ProtoUtils.hash_float32(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.DoubleKind:
		return "ProtoUtils.hash_float64(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "ProtoUtils.hash_int32(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "ProtoUtils.hash_uint32(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "ProtoUtils.hash_int64(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "ProtoUtils.hash_uint64(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.EnumKind:
		return "ProtoUtils.hash_enum(" + hasherName + ", " + valueExpr + ")", nil
	default:
		return "", fmt.Errorf("unsupported hash field kind %s for %s", field.Desc.Kind(), field.Desc.FullName())
	}
}

func hashDictionaryKeyOrderSuffix(keyField *protogen.Field) string {
	switch keyField.Desc.Kind() {
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return ", ProtoUtils.DictionaryKeyOrder.UINT64"
	default:
		return ""
	}
}

func equalCallExpression(leftExpr, rightExpr string, field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) (string, error) {
	if field.Message != nil && !field.Desc.IsMap() {
		typeExpr, err := gdscriptSingularTypeExpression(field, protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		return "ProtoUtils.equal_message(" + leftExpr + ", " + rightExpr + ", func(): return " + typeExpr + ".new())", nil
	}
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "String(" + leftExpr + ") == String(" + rightExpr + ")", nil
	case protoreflect.BytesKind:
		return "ProtoUtils.equal_bytes(" + leftExpr + ", " + rightExpr + ")", nil
	case protoreflect.FloatKind:
		return "ProtoUtils.equal_float32(" + leftExpr + ", " + rightExpr + ")", nil
	case protoreflect.DoubleKind:
		return "ProtoUtils.equal_float64(" + leftExpr + ", " + rightExpr + ")", nil
	default:
		return leftExpr + " == " + rightExpr, nil
	}
}

func gdscriptArgumentList(fields []*protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) (string, error) {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		typeExpr, err := gdscriptTypeExpression(field, protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		parts = append(parts, gdscriptArgumentName(field)+": "+typeExpr)
	}
	return strings.Join(parts, ", "), nil
}

func gdscriptArgumentNames(fields []*protogen.Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, gdscriptArgumentName(field))
	}
	return strings.Join(parts, ", ")
}

func gdscriptArgumentName(field *protogen.Field) string {
	name := safeIdentifier(string(field.Desc.Name()))
	if name == "" {
		return "_value"
	}
	return safeIdentifier(name)
}

func gdscriptTypeExpression(field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) (string, error) {
	if field.Desc.IsMap() {
		if field.Message == nil || len(field.Message.Fields) < 2 {
			return "", fmt.Errorf("map field %s has invalid entry descriptor", field.Desc.FullName())
		}
		keyType, err := gdscriptSingularTypeExpression(field.Message.Fields[0], protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		valueType, err := gdscriptSingularTypeExpression(field.Message.Fields[1], protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		return "Dictionary[" + keyType + ", " + valueType + "]", nil
	}
	if field.Desc.IsList() {
		itemType, err := gdscriptSingularTypeExpression(field, protoImportAlias, messageTypeNames)
		if err != nil {
			return "", err
		}
		return "Array[" + itemType + "]", nil
	}
	return gdscriptSingularTypeExpression(field, protoImportAlias, messageTypeNames)
}

func gdscriptSingularTypeExpression(field *protogen.Field, protoImportAlias string, messageTypeNames map[protoreflect.FullName]string) (string, error) {
	if field.Enum != nil {
		return protoImportAlias + "." + enumTypeReferenceName(field.Enum, messageTypeNames), nil
	}
	if field.Message != nil && !field.Desc.IsMap() {
		return protoImportAlias + "." + safeIdentifier(field.Message.GoIdent.GoName), nil
	}

	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool", nil
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "StringName", nil
		}
		return "String", nil
	case protoreflect.BytesKind:
		return "PackedByteArray", nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "float", nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind,
		protoreflect.EnumKind:
		return "int", nil
	default:
		return "", fmt.Errorf("unsupported gdscript field kind %s for %s", field.Desc.Kind(), field.Desc.FullName())
	}
}

func enumTypeReferenceName(enum *protogen.Enum, messageTypeNames map[protoreflect.FullName]string) string {
	if parent, ok := enum.Desc.Parent().(protoreflect.MessageDescriptor); ok {
		if messageTypeName, exists := messageTypeNames[parent.FullName()]; exists {
			return messageTypeName + "." + safeIdentifier(string(enum.Desc.Name()))
		}
	}
	return safeIdentifier(enum.GoIdent.GoName)
}

func directIndexExpression(argName string, field *protogen.Field) (string, error) {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ExcelUtils.boolean_to_index(" + argName + ")", nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "ExcelUtils.integer_to_index(" + argName + ")", nil
	case protoreflect.FloatKind:
		return "ExcelUtils.float_to_index(" + argName + ")", nil
	case protoreflect.DoubleKind:
		return "ExcelUtils.double_to_index(" + argName + ")", nil
	case protoreflect.EnumKind:
		return "ExcelUtils.integer_to_index(" + argName + ")", nil
	default:
		return "", fmt.Errorf("unsupported direct index field kind %s for %s", field.Desc.Kind(), field.Desc.FullName())
	}
}

func requiresHashIndex(fields []*protogen.Field) bool {
	return len(fields) != 1 || !supportsDirectIndex(fields[0])
}

func supportsDirectIndex(field *protogen.Field) bool {
	if field.Desc.IsMap() || field.Desc.IsList() {
		return false
	}

	switch field.Desc.Kind() {
	case protoreflect.BoolKind,
		protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind,
		protoreflect.Fixed64Kind,
		protoreflect.FloatKind,
		protoreflect.DoubleKind,
		protoreflect.EnumKind:
		return true
	default:
		return false
	}
}

func indexHashMethodName(method IndexMethodDecl) string {
	return "_" + method.LookupMethodName + "_index"
}

func indexMatchMethodName(method IndexMethodDecl) string {
	return "_" + method.LookupMethodName + "_match"
}

func importAliasIdentifier(s string) string {
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
	if b.Len() <= 0 {
		return "ProtoPB"
	}

	out := b.String()
	if unicode.IsDigit(rune(out[0])) {
		return "Proto" + out + "PB"
	}
	return out + "PB"
}

func safeIdentifier(s string) string {
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
	if isGDScriptKeyword(out) {
		out += "_"
	}
	return out
}

func isGDScriptKeyword(s string) bool {
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
