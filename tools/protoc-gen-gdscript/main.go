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
	"flag"
	"fmt"
	"hash/fnv"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"git.golaxy.org/framework/net/gap/variant"
	"github.com/elliotchance/pie/v2"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GeneratorConfig struct {
	StringAsStringName bool
	Deterministic      bool
	GapVariant         bool
}

var config GeneratorConfig

func main() {
	var flags flag.FlagSet
	stringAsStringName := flags.Bool("string_as_string_name", false, "map proto string fields to GDScript StringName")
	deterministic := flags.Bool("deterministic", false, "serialize map fields in deterministic key order")
	gapVariant := flags.Bool("gap_variant", false, "generate messages as ProtoGAPVariant implementations")

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		config = GeneratorConfig{
			StringAsStringName: *stringAsStringName,
			Deterministic:      *deterministic,
			GapVariant:         *gapVariant,
		}
		generatedPrefixes := map[string]string{}
		for _, f := range gen.Files {
			generatedPrefixes[f.Desc.Path()] = f.GeneratedFilenamePrefix
		}
		for _, f := range gen.Files {
			if f.Generate {
				if err := generateFile(gen, f, generatedPrefixes); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File, generatedPrefixes map[string]string) error {
	g := gen.NewGeneratedFile(file.GeneratedFilenamePrefix+".pb.gd", "")

	enums := collectEnums(file)
	messages := collectMessages(file)
	usedDeps := collectDependencies(file)
	importAliases, err := collectImportAliases(file, usedDeps)
	if err != nil {
		return err
	}

	emitGeneratedHeader(gen, file, g)

	if err := emitImportAliasConstants(g, file, usedDeps, importAliases, generatedPrefixes); err != nil {
		return err
	}

	emitScriptStaticInit(g, file, messages)

	for _, enum := range enums {
		emitEnum(g, enum)
	}

	for _, msg := range messages {
		if err := emitMessage(g, file, msg, importAliases); err != nil {
			return err
		}
	}

	return nil
}

func collectEnums(file *protogen.File) []*protogen.Enum {
	return pie.SortUsing(file.Enums, func(a, b *protogen.Enum) bool { return a.Desc.Name() < b.Desc.Name() })
}

func collectMessages(file *protogen.File) []*protogen.Message {
	var msgs []*protogen.Message
	var walk func(*protogen.Message)
	walk = func(msg *protogen.Message) {
		if msg.Desc.IsMapEntry() {
			return
		}
		msgs = append(msgs, msg)
		for _, nested := range msg.Messages {
			walk(nested)
		}
	}
	for _, msg := range file.Messages {
		walk(msg)
	}
	return pie.SortUsing(msgs, func(a, b *protogen.Message) bool { return a.Desc.Name() < b.Desc.Name() })
}

func collectDependencies(file *protogen.File) map[string]struct{} {
	usedDeps := map[string]struct{}{}
	var walk func(*protogen.Message)
	walk = func(msg *protogen.Message) {
		if msg.Desc.IsMapEntry() {
			return
		}
		for _, field := range msg.Fields {
			collectDependenciesFromField(file, field, usedDeps)
		}
		for _, nested := range msg.Messages {
			walk(nested)
		}
	}
	for _, msg := range file.Messages {
		walk(msg)
	}
	return usedDeps
}

func collectDependenciesFromField(file *protogen.File, field *protogen.Field, usedDeps map[string]struct{}) {
	if field.Desc.IsMap() {
		if len(field.Message.Fields) >= 2 {
			collectTypeDependency(file, field.Message.Fields[0], usedDeps)
			collectTypeDependency(file, field.Message.Fields[1], usedDeps)
		}
		return
	}
	collectTypeDependency(file, field, usedDeps)
}

func collectTypeDependency(file *protogen.File, field *protogen.Field, usedDeps map[string]struct{}) {
	if field.Enum != nil && field.Enum.Desc.ParentFile().Path() != file.Desc.Path() {
		usedDeps[field.Enum.Desc.ParentFile().Path()] = struct{}{}
	}
	if field.Message != nil && field.Message.Desc.ParentFile().Path() != file.Desc.Path() {
		usedDeps[field.Message.Desc.ParentFile().Path()] = struct{}{}
	}
}

func collectImportAliases(file *protogen.File, usedDeps map[string]struct{}) (map[string]string, error) {
	aliases := map[string]string{}
	seen := map[string]int{}
	for _, dep := range file.Proto.Dependency {
		if _, ok := usedDeps[dep]; !ok {
			continue
		}
		base := path.Base(strings.TrimSuffix(dep, path.Ext(dep)))
		baseAlias := importAliasIdentifier(base)
		seen[baseAlias]++
		alias := baseAlias
		if seen[baseAlias] > 1 {
			alias = fmt.Sprintf("%s%d", baseAlias, seen[baseAlias])
		}
		if _, exists := aliases[dep]; exists {
			return nil, fmt.Errorf("duplicate import alias mapping for dependency %q", dep)
		}
		aliases[dep] = alias
	}
	return aliases, nil
}

func emitGeneratedHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("# Code generated by protoc-gen-gdscript. DO NOT EDIT.")
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

func emitImportAliasConstants(g *protogen.GeneratedFile, file *protogen.File, usedDeps map[string]struct{}, importAliases map[string]string, generatedPrefixes map[string]string) error {
	var count int
	for _, dep := range file.Proto.Dependency {
		if _, ok := usedDeps[dep]; !ok {
			continue
		}
		alias, ok := importAliases[dep]
		if !ok {
			return fmt.Errorf("missing import alias for dependency %q in %s", dep, file.Desc.Path())
		}
		toPrefix, ok := generatedPrefixes[dep]
		if !ok {
			return fmt.Errorf("missing generated prefix for dependency %q in %s", dep, file.Desc.Path())
		}
		rel, err := relativeGeneratedPath(file.GeneratedFilenamePrefix, toPrefix)
		if err != nil {
			return fmt.Errorf("resolve generated path from %q to %q: %w", file.GeneratedFilenamePrefix, toPrefix, err)
		}
		g.P("const ", alias, " = preload(", strconv.Quote(rel+".pb.gd"), ")")
		count++
	}
	if count > 0 {
		g.P()
	}
	return nil
}

func relativeGeneratedPath(fromPrefix, toPrefix string) (string, error) {
	fromDir := path.Dir(path.Clean(fromPrefix))
	if fromDir == "." {
		fromDir = ""
	}
	rel, err := filepath.Rel(filepath.FromSlash(fromDir), filepath.FromSlash(path.Clean(toPrefix)))
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel, nil
}

func emitEnum(g *protogen.GeneratedFile, enum *protogen.Enum) {
	if len(enum.Values) <= 0 {
		return
	}
	enumName := safeIdentifier(enum.GoIdent.GoName)

	g.P("enum ", enumName, " {")
	for _, value := range enum.Values {
		g.P("\t", safeIdentifier(string(value.Desc.Name())), " = ", value.Desc.Number(), ",")
	}
	g.P("}")
	g.P()
	emitEnumJSONHelper(g, enum, 0, enumName)
}

func emitMessage(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	if msg.Desc.IsMapEntry() {
		return nil
	}
	msgName := safeIdentifier(msg.GoIdent.GoName)

	g.P("class ", msgName, ":")
	g.P("\textends ", messageBaseType())
	g.P()

	for _, enum := range msg.Enums {
		emitIndentedEnum(g, enum, 1)
	}

	if len(msg.Fields) <= 0 && len(msg.Enums) <= 0 {
		emitEmptyMessageMethods(g, file, msg, msgName)
		return nil
	}

	if err := emitMessageFields(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitSerializeMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitDeserializeMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitToDictMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitFromDictMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitSizeMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitResetMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitCloneMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitHashToMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	if err := emitEqualsMethod(g, file, msg, importAliases); err != nil {
		return err
	}
	emitTypeIDMethod(g, file, msg)
	return nil
}

func emitIndentedEnum(g *protogen.GeneratedFile, enum *protogen.Enum, indentLevel int) {
	if len(enum.Values) <= 0 {
		return
	}
	indent := strings.Repeat("\t", indentLevel)
	enumName := safeIdentifier(string(enum.Desc.Name()))

	g.P(indent, "enum ", enumName, " {")
	for _, value := range enum.Values {
		g.P(indent, "\t", safeIdentifier(string(value.Desc.Name())), " = ", value.Desc.Number(), ",")
	}
	g.P(indent, "}")
	g.P()
	emitEnumJSONHelper(g, enum, indentLevel, enumName)
}

func emitMessageFields(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	if len(msg.Fields) <= 0 {
		return nil
	}
	for _, field := range msg.Fields {
		typeExpr, err := fieldTypeExpression(file, field, importAliases)
		if err != nil {
			return err
		}
		defaultExpr, err := fieldDefaultValueExpression(file, field, importAliases)
		if err != nil {
			return err
		}
		g.P("\tvar ", safeIdentifier(field.GoName), ": ", typeExpr, " = ", defaultExpr)
	}
	g.P()
	return nil
}

func emitEmptyMessageMethods(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, msgName string) {
	g.P("\tfunc serialize(pb_stream: ProtoOutputStream) -> bool:")
	g.P("\t\tif pb_stream.get_error() != OK:")
	g.P("\t\t\treturn false")
	g.P("\t\treturn true")
	g.P()
	g.P("\tfunc deserialize(pb_stream: ProtoInputStream) -> bool:")
	g.P("\t\twhile !pb_stream.eof():")
	g.P("\t\t\tvar pb_tag := ProtoUtils.decode_tag(pb_stream)")
	g.P("\t\t\tif pb_stream.get_error() != OK:")
	g.P("\t\t\t\treturn false")
	g.P("\t\t\tif !ProtoUtils.skip_field(pb_stream, ProtoUtils.get_tag_wire_type(pb_tag)):")
	g.P("\t\t\t\treturn false")
	g.P("\t\treturn true")
	g.P()
	g.P("\tfunc to_dict(json_emit_default: bool = false, json_enum_as_string: bool = true) -> Dictionary:")
	g.P("\t\treturn {}")
	g.P()
	g.P("\t@warning_ignore(\"unused_parameter\")")
	g.P("\tfunc from_dict(json_dict: Dictionary) -> bool:")
	g.P("\t\treturn true")
	g.P()
	g.P("\tfunc size() -> int:")
	g.P("\t\treturn 0")
	g.P()
	g.P("\tfunc reset() -> void:")
	g.P("\t\tpass")
	g.P()
	g.P("\tfunc clone() -> ProtoMessage:")
	g.P("\t\treturn ", msgName, ".new()")
	g.P()
	g.P("\tfunc hash_to(pb_hasher: ProtoHasher) -> void:")
	g.P("\t\tProtoUtils.hash_message_fields(pb_hasher, 0)")
	g.P()
	g.P("\tfunc equals(pb_other: ProtoMessage) -> bool:")
	g.P("\t\treturn pb_other is ", msgName)
	g.P()
	emitTypeIDMethod(g, file, msg)
}

func emitSerializeMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc serialize(pb_stream: ProtoOutputStream) -> bool:")
	g.P("\t\tif pb_stream.get_error() != OK:")
	g.P("\t\t\treturn false")
	if len(msg.Fields) <= 0 {
		g.P("\t\treturn true")
		g.P()
		return nil
	}
	for _, field := range msg.Fields {
		if err := emitSerializeField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn true")
	g.P()
	return nil
}

func emitSerializeField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	fieldNumber := int(field.Desc.Number())
	fieldType := fieldTypeConst(field)
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		valueField := field.Message.Fields[1]
		iterExpr := name
		if config.Deterministic {
			iterExpr = "ProtoUtils.sorted_dictionary_keys(" + name + dictionaryKeyOrderSuffix(keyField) + ")"
		}
		g.P("\t\tfor pb_key in ", iterExpr, ":")
		g.P("\t\t\tvar pb_value := ", name, "[pb_key]")
		g.P("\t\t\tif !ProtoUtils.encode_tag(pb_stream, ", fieldNumber, ", ProtoFieldDescriptor.FieldType.TYPE_MAP):")
		g.P("\t\t\t\treturn false")
		g.P("\t\t\t@warning_ignore(\"confusable_local_declaration\")")
		g.P(
			"\t\t\tvar pb_entry_size := ProtoUtils.sizeof_dictionary_entry(pb_key, pb_value, ",
			tagSizeLiteral(1, fieldTypeConst(keyField)),
			", func(pb_key): return ",
			scalarSizeExpression("pb_key", keyField, file, importAliases),
			", ",
			tagSizeLiteral(2, fieldTypeConst(valueField)),
			", func(pb_value): return ",
			valueSizeExpression("pb_value", valueField, file, importAliases),
			", func(pb_value): return ",
			shouldSerializeExpression("pb_value", valueField),
			")",
		)
		g.P("\t\t\tif !ProtoUtils.encode_varint(pb_stream, pb_entry_size):")
		g.P("\t\t\t\treturn false")
		g.P("\t\t\tif !ProtoUtils.encode_tag(pb_stream, 1, ", fieldTypeConst(keyField), "):")
		g.P("\t\t\t\treturn false")
		emitEncodeValue(g, "\t\t\t", "pb_key", keyField, file, importAliases)
		g.P("\t\t\tif ", shouldSerializeExpression("pb_value", valueField), ":")
		g.P("\t\t\t\tif !ProtoUtils.encode_tag(pb_stream, 2, ", fieldTypeConst(valueField), "):")
		g.P("\t\t\t\t\treturn false")
		if err := emitEncodeValue(g, "\t\t\t\t", "pb_value", valueField, file, importAliases); err != nil {
			return err
		}
		return nil
	}
	if field.Desc.IsList() {
		if isPackedField(field) {
			g.P("\t\tif !", name, ".is_empty():")
			g.P("\t\t\tif !ProtoUtils.encode_tag(pb_stream, ", fieldNumber, ", ", fieldType, "):")
			g.P("\t\t\t\treturn false")
			g.P("\t\t\tvar pb_data_size := ProtoUtils.sizeof_array_payload(", name, ", func(pb_value): return ", scalarSizeExpression("pb_value", field, file, importAliases), ")")
			g.P("\t\t\tif !ProtoUtils.encode_varint(pb_stream, pb_data_size):")
			g.P("\t\t\t\treturn false")
			g.P("\t\t\tfor pb_value in ", name, ":")
			if err := emitEncodeValue(g, "\t\t\t\t", "pb_value", field, file, importAliases); err != nil {
				return err
			}
			return nil
		}
		g.P("\t\tfor pb_value in ", name, ":")
		g.P("\t\t\tif !ProtoUtils.encode_tag(pb_stream, ", fieldNumber, ", ", fieldType, "):")
		g.P("\t\t\t\treturn false")
		if err := emitEncodeValue(g, "\t\t\t", "pb_value", field, file, importAliases); err != nil {
			return err
		}
		return nil
	}
	g.P("\t\tif ", shouldSerializeExpression(name, field), ":")
	g.P("\t\t\tif !ProtoUtils.encode_tag(pb_stream, ", fieldNumber, ", ", fieldType, "):")
	g.P("\t\t\t\treturn false")
	if err := emitEncodeValue(g, "\t\t\t", name, field, file, importAliases); err != nil {
		return err
	}
	return nil
}

func emitEncodeValue(g *protogen.GeneratedFile, indent, valueExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) error {
	if field.Message != nil && !field.Desc.IsMap() {
		g.P(indent, "if !ProtoUtils.encode_message(pb_stream, ", valueExpr, "):")
		g.P(indent, "\treturn false")
		return nil
	}
	g.P(indent, "if !", encodeValueCall(valueExpr, field), ":")
	g.P(indent, "\treturn false")
	return nil
}

func emitDeserializeMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc deserialize(pb_stream: ProtoInputStream) -> bool:")
	g.P("\t\twhile !pb_stream.eof():")
	g.P("\t\t\tvar pb_tag := ProtoUtils.decode_tag(pb_stream)")
	g.P("\t\t\tif pb_stream.get_error() != OK:")
	g.P("\t\t\t\treturn false")
	g.P("\t\t\tvar pb_field_number := ProtoUtils.get_tag_field_number(pb_tag)")
	g.P("\t\t\tvar pb_wire_type := ProtoUtils.get_tag_wire_type(pb_tag)")
	g.P("\t\t\tmatch pb_field_number:")
	for _, field := range msg.Fields {
		if err := emitDeserializeField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\t\t\t_:")
	g.P("\t\t\t\t\tif !ProtoUtils.skip_field(pb_stream, pb_wire_type):")
	g.P("\t\t\t\t\t\treturn false")
	g.P("\t\treturn true")
	g.P()
	return nil
}

func emitDeserializeField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	fieldNumber := int(field.Desc.Number())
	name := safeIdentifier(field.GoName)
	g.P("\t\t\t\t", fieldNumber, ":")
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		valueField := field.Message.Fields[1]
		g.P("\t\t\t\t\tif pb_wire_type != ProtoFieldDescriptor.WireType.WIRETYPE_LENGTH_DELIMITED:")
		g.P("\t\t\t\t\t\treturn false")
		g.P("\t\t\t\t\tvar pb_entry_size := ProtoUtils.decode_varint(pb_stream)")
		g.P("\t\t\t\t\tif pb_stream.get_error() != OK or pb_entry_size < 0:")
		g.P("\t\t\t\t\t\treturn false")
		g.P("\t\t\t\t\tvar pb_entry_stream := ProtoLimitedInputStream.new(pb_stream, pb_entry_size)")
		g.P("\t\t\t\t\tvar pb_entry_key := ", defaultMapKeyExpression(keyField))
		entryValueExpr, err := defaultMapValueExpression(file, valueField, importAliases)
		if err != nil {
			return err
		}
		entryValueType, err := fieldSingularTypeExpression(file, valueField, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\t\t\t\tvar pb_entry_value: ", entryValueType, " = ", entryValueExpr)
		g.P("\t\t\t\t\twhile !pb_entry_stream.eof():")
		g.P("\t\t\t\t\t\tvar pb_entry_tag := ProtoUtils.decode_tag(pb_entry_stream)")
		g.P("\t\t\t\t\t\tif pb_entry_stream.get_error() != OK:")
		g.P("\t\t\t\t\t\t\treturn false")
		g.P("\t\t\t\t\t\tvar pb_entry_field_number := ProtoUtils.get_tag_field_number(pb_entry_tag)")
		g.P("\t\t\t\t\t\tvar pb_entry_wire_type := ProtoUtils.get_tag_wire_type(pb_entry_tag)")
		g.P("\t\t\t\t\t\tmatch pb_entry_field_number:")
		g.P("\t\t\t\t\t\t\t1:")
		if err := emitCheckedDecodedAssignment(g, "\t\t\t\t\t\t\t\t", "pb_entry_key", keyField, file, importAliases, "pb_entry_stream", "pb_entry_wire_type"); err != nil {
			return err
		}
		g.P("\t\t\t\t\t\t\t2:")
		if err := emitCheckedDecodedAssignment(g, "\t\t\t\t\t\t\t\t", "pb_entry_value", valueField, file, importAliases, "pb_entry_stream", "pb_entry_wire_type"); err != nil {
			return err
		}
		g.P("\t\t\t\t\t\t\t_:")
		g.P("\t\t\t\t\t\t\t\tif !ProtoUtils.skip_field(pb_entry_stream, pb_entry_wire_type):")
		g.P("\t\t\t\t\t\t\t\t\treturn false")
		g.P("\t\t\t\t\t", name, "[pb_entry_key] = pb_entry_value")
		return nil
	}
	if field.Desc.IsList() {
		if isPackedField(field) {
			g.P("\t\t\t\t\tif pb_wire_type == ProtoFieldDescriptor.WireType.WIRETYPE_LENGTH_DELIMITED:")
			g.P("\t\t\t\t\t\tvar pb_packed_size := ProtoUtils.decode_varint(pb_stream)")
			g.P("\t\t\t\t\t\tif pb_stream.get_error() != OK or pb_packed_size < 0:")
			g.P("\t\t\t\t\t\t\treturn false")
			g.P("\t\t\t\t\t\tvar pb_packed_stream := ProtoLimitedInputStream.new(pb_stream, pb_packed_size)")
			g.P("\t\t\t\t\t\twhile !pb_packed_stream.eof():")
			if err := emitDecodedAppend(g, "\t\t\t\t\t\t\t", name, field, file, importAliases, "pb_packed_stream"); err != nil {
				return err
			}
			g.P("\t\t\t\t\telif pb_wire_type == ", wireTypeConst(field), ":")
			if err := emitDecodedAppend(g, "\t\t\t\t\t\t", name, field, file, importAliases, "pb_stream"); err != nil {
				return err
			}
			g.P("\t\t\t\t\telse:")
			g.P("\t\t\t\t\t\treturn false")
			return nil
		}
		g.P("\t\t\t\t\tif pb_wire_type != ", wireTypeConst(field), ":")
		g.P("\t\t\t\t\t\treturn false")
		if err := emitDecodedAppend(g, "\t\t\t\t\t", name, field, file, importAliases, "pb_stream"); err != nil {
			return err
		}
		return nil
	}
	g.P("\t\t\t\t\tif pb_wire_type != ", wireTypeConst(field), ":")
	g.P("\t\t\t\t\t\treturn false")
	if err := emitDecodedAssignment(g, "\t\t\t\t\t", name, field, file, importAliases, "pb_stream"); err != nil {
		return err
	}
	return nil
}

func emitDecodedAppend(g *protogen.GeneratedFile, indent, target string, field *protogen.Field, file *protogen.File, importAliases map[string]string, streamName string) error {
	if field.Message != nil && !field.Desc.IsMap() {
		msgType, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return err
		}
		g.P(indent, "var pb_value := ", msgType, ".new()")
		g.P(indent, "if !ProtoUtils.decode_message(", streamName, ", pb_value):")
		g.P(indent, "\treturn false")
		g.P(indent, target, ".append(pb_value)")
		return nil
	}
	if field.Enum != nil {
		g.P(indent, `@warning_ignore("int_as_enum_without_cast")`)
	}
	g.P(indent, "var pb_value := ", decodeValueExpression(field, streamName))
	g.P(indent, "if ", streamName, ".get_error() != OK:")
	g.P(indent, "\treturn false")
	g.P(indent, target, ".append(pb_value)")
	return nil
}

func emitDecodedAssignment(g *protogen.GeneratedFile, indent, target string, field *protogen.Field, file *protogen.File, importAliases map[string]string, streamName string) error {
	if field.Message != nil && !field.Desc.IsMap() {
		msgType, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return err
		}
		g.P(indent, "var pb_value := ", msgType, ".new()")
		g.P(indent, "if !ProtoUtils.decode_message(", streamName, ", pb_value):")
		g.P(indent, "\treturn false")
		g.P(indent, target, " = pb_value")
		return nil
	}
	g.P(indent, "var pb_value := ", decodeValueExpression(field, streamName))
	g.P(indent, "if ", streamName, ".get_error() != OK:")
	g.P(indent, "\treturn false")
	if field.Enum != nil {
		g.P(indent, `@warning_ignore("int_as_enum_without_cast")`)
	}
	g.P(indent, target, " = pb_value")
	return nil
}

func emitCheckedDecodedAssignment(g *protogen.GeneratedFile, indent, target string, field *protogen.Field, file *protogen.File, importAliases map[string]string, streamName, wireTypeExpr string) error {
	g.P(indent, "if ", wireTypeExpr, " != ", wireTypeConst(field), ":")
	g.P(indent, "\treturn false")
	if field.Message != nil && !field.Desc.IsMap() {
		g.P(indent, "if !ProtoUtils.decode_message(", streamName, ", ", target, "):")
		g.P(indent, "\treturn false")
		return nil
	}
	if field.Enum != nil {
		g.P(indent, `@warning_ignore("int_as_enum_without_cast")`)
	}
	g.P(indent, target, " = ", decodeValueExpression(field, streamName))
	g.P(indent, "if ", streamName, ".get_error() != OK:")
	g.P(indent, "\treturn false")
	return nil
}

func emitEnumJSONHelper(g *protogen.GeneratedFile, enum *protogen.Enum, indentLevel int, typeRef string) {
	if len(enum.Values) <= 0 {
		return
	}
	indent := strings.Repeat("\t", indentLevel)
	toStringName, fromStringName := enumStringHelperNames(enum)
	g.P(indent, "static func ", toStringName, "(value: int) -> String:")
	g.P(indent, "\tmatch value:")
	for _, value := range enum.Values {
		valueName := safeIdentifier(string(value.Desc.Name()))
		g.P(indent, "\t\t", typeRef, ".", valueName, ":")
		g.P(indent, "\t\t\treturn ", strconv.Quote(string(value.Desc.Name())))
	}
	g.P(indent, "\treturn str(value)")
	g.P()
	g.P(indent, "static func ", fromStringName, "(value: Variant) -> int:")
	g.P(indent, "\tmatch value:")
	for _, value := range enum.Values {
		valueName := safeIdentifier(string(value.Desc.Name()))
		g.P(indent, "\t\t", strconv.Quote(string(value.Desc.Name())), ":")
		g.P(indent, "\t\t\treturn ", typeRef, ".", valueName)
	}
	g.P(indent, "\treturn int(value)")
	g.P()
}

func emitToDictMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc to_dict(json_emit_default: bool = false, json_enum_as_string: bool = true) -> Dictionary:")
	g.P("\t\tvar json_dict := {}")
	for _, field := range msg.Fields {
		if err := emitToDictField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn json_dict")
	g.P()
	return nil
}

func emitToDictField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	jsonName := strconv.Quote(field.Desc.JSONName())
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		valueField := field.Message.Fields[1]
		g.P("\t\tif json_emit_default or !", name, ".is_empty():")
		g.P("\t\t\tvar pb_dict := {}")
		g.P("\t\t\tfor pb_key in ", name, ":")
		g.P("\t\t\t\tvar pb_value := ", name, "[pb_key]")
		if shouldIgnoreIncompatibleTernaryInToDict(valueField) {
			g.P("\t\t\t\t@warning_ignore(\"incompatible_ternary\")")
		}
		valueExpr, err := jsonToDictValueExpression("pb_value", valueField, file, importAliases, "json_emit_default", "json_enum_as_string")
		if err != nil {
			return err
		}
		g.P("\t\t\t\tpb_dict[", jsonMapKeyToDictExpression("pb_key", keyField), "] = ", valueExpr)
		g.P("\t\t\tjson_dict[", jsonName, "] = pb_dict")
		return nil
	}
	if field.Desc.IsList() {
		g.P("\t\tif json_emit_default or !", name, ".is_empty():")
		g.P("\t\t\tvar pb_array := []")
		g.P("\t\t\tfor pb_value in ", name, ":")
		if shouldIgnoreIncompatibleTernaryInToDict(field) {
			g.P("\t\t\t\t@warning_ignore(\"incompatible_ternary\")")
		}
		valueExpr, err := jsonToDictValueExpression("pb_value", field, file, importAliases, "json_emit_default", "json_enum_as_string")
		if err != nil {
			return err
		}
		g.P("\t\t\t\tpb_array.append(", valueExpr, ")")
		g.P("\t\t\tjson_dict[", jsonName, "] = pb_array")
		return nil
	}
	g.P("\t\tif json_emit_default or ", shouldSerializeExpression(name, field), ":")
	if shouldIgnoreIncompatibleTernaryInToDict(field) {
		g.P("\t\t\t@warning_ignore(\"incompatible_ternary\")")
	}
	valueExpr, err := jsonToDictValueExpression(name, field, file, importAliases, "json_emit_default", "json_enum_as_string")
	if err != nil {
		return err
	}
	g.P("\t\t\tjson_dict[", jsonName, "] = ", valueExpr)
	return nil
}

func emitFromDictMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc from_dict(json_dict: Dictionary) -> bool:")
	for _, field := range msg.Fields {
		if err := emitFromDictField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn true")
	g.P()
	return nil
}

func emitFromDictField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	jsonName := strconv.Quote(field.Desc.JSONName())
	fieldValueName := "pb_field"
	g.P("\t\tif json_dict.has(", jsonName, "):")
	g.P("\t\t\tvar ", fieldValueName, " = json_dict[", jsonName, "]")
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		valueField := field.Message.Fields[1]
		g.P("\t\t\tif !(", fieldValueName, " is Dictionary):")
		g.P("\t\t\t\treturn false")
		g.P("\t\t\t", name, " = {}")
		g.P("\t\t\tfor pb_key in ", fieldValueName, ":")
		g.P("\t\t\t\tvar pb_entry_key := ", jsonMapKeyFromDictExpression("pb_key", keyField))
		if valueField.Message != nil && !valueField.Desc.IsMap() {
			msgType, err := fieldMessageTypeReference(file, valueField.Message, importAliases)
			if err != nil {
				return err
			}
			g.P("\t\t\t\tvar pb_entry_value = ", fieldValueName, "[pb_key]")
			g.P("\t\t\t\tif pb_entry_value == null:")
			g.P("\t\t\t\t\t", name, "[pb_entry_key] = null")
			g.P("\t\t\t\telse:")
			g.P("\t\t\t\t\tif !(pb_entry_value is Dictionary):")
			g.P("\t\t\t\t\t\treturn false")
			g.P("\t\t\t\t\tvar pb_msg := ", msgType, ".new()")
			g.P("\t\t\t\t\tif !pb_msg.from_dict(pb_entry_value):")
			g.P("\t\t\t\t\t\treturn false")
			g.P("\t\t\t\t\t", name, "[pb_entry_key] = pb_msg")
			return nil
		}
		g.P("\t\t\t\tvar pb_entry_value = ", fieldValueName, "[pb_key]")
		valueExpr, err := jsonFromDictValueExpression("pb_entry_value", valueField, file, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\t\t\t", name, "[pb_entry_key] = ", valueExpr)
		return nil
	}
	if field.Desc.IsList() {
		g.P("\t\t\tif !(", fieldValueName, " is Array):")
		g.P("\t\t\t\treturn false")
		g.P("\t\t\t", name, " = []")
		g.P("\t\t\tfor pb_value in ", fieldValueName, ":")
		if field.Message != nil && !field.Desc.IsMap() {
			msgType, err := fieldMessageTypeReference(file, field.Message, importAliases)
			if err != nil {
				return err
			}
			g.P("\t\t\t\tif pb_value == null:")
			g.P("\t\t\t\t\t", name, ".append(null)")
			g.P("\t\t\t\telse:")
			g.P("\t\t\t\t\tif !(pb_value is Dictionary):")
			g.P("\t\t\t\t\t\treturn false")
			g.P("\t\t\t\t\tvar pb_msg := ", msgType, ".new()")
			g.P("\t\t\t\t\tif !pb_msg.from_dict(pb_value):")
			g.P("\t\t\t\t\t\treturn false")
			g.P("\t\t\t\t\t", name, ".append(pb_msg)")
			return nil
		}
		valueExpr, err := jsonFromDictValueExpression("pb_value", field, file, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\t\t\t", name, ".append(", valueExpr, ")")
		return nil
	}
	if field.Message != nil && !field.Desc.IsMap() {
		msgType, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\t\tif ", fieldValueName, " == null:")
		g.P("\t\t\t\t", name, " = null")
		g.P("\t\t\telse:")
		g.P("\t\t\t\tif !(", fieldValueName, " is Dictionary):")
		g.P("\t\t\t\t\treturn false")
		g.P("\t\t\t\tvar pb_msg := ", msgType, ".new()")
		g.P("\t\t\t\tif !pb_msg.from_dict(", fieldValueName, "):")
		g.P("\t\t\t\t\treturn false")
		g.P("\t\t\t\t", name, " = pb_msg")
		return nil
	}
	valueExpr, err := jsonFromDictValueExpression(fieldValueName, field, file, importAliases)
	if err != nil {
		return err
	}
	if field.Enum != nil {
		g.P("\t\t\t@warning_ignore(\"int_as_enum_without_cast\")")
	}
	g.P("\t\t\t", name, " = ", valueExpr)
	return nil
}

func emitSizeMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc size() -> int:")
	g.P("\t\tvar pb_msg_size := 0")
	for _, field := range msg.Fields {
		if err := emitSizeField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn pb_msg_size")
	g.P()
	return nil
}

func emitSizeField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	fieldNumber := int(field.Desc.Number())
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		valueField := field.Message.Fields[1]
		g.P(
			"\t\tpb_msg_size += ProtoUtils.sizeof_dictionary(",
			name,
			", ",
			tagSizeLiteral(fieldNumber, "ProtoFieldDescriptor.FieldType.TYPE_MAP"),
			", ",
			tagSizeLiteral(1, fieldTypeConst(keyField)),
			", func(pb_key): return ",
			scalarSizeExpression("pb_key", keyField, file, importAliases),
			", ",
			tagSizeLiteral(2, fieldTypeConst(valueField)),
			", func(pb_value): return ",
			valueSizeExpression("pb_value", valueField, file, importAliases),
			", func(pb_value): return ",
			shouldSerializeExpression("pb_value", valueField),
			")",
		)
		return nil
	}
	if field.Desc.IsList() {
		if isPackedField(field) {
			g.P("\t\tpb_msg_size += ProtoUtils.sizeof_packed_array(", name, ", ", tagSizeLiteral(fieldNumber, fieldTypeConst(field)), ", func(pb_value): return ", scalarSizeExpression("pb_value", field, file, importAliases), ")")
			return nil
		}
		g.P("\t\tpb_msg_size += ProtoUtils.sizeof_array(", name, ", ", tagSizeLiteral(fieldNumber, fieldTypeConst(field)), ", func(pb_value): return ", valueSizeExpression("pb_value", field, file, importAliases), ")")
		return nil
	}
	g.P("\t\tif ", shouldSerializeExpression(name, field), ":")
	g.P("\t\t\tpb_msg_size += ", tagSizeLiteral(fieldNumber, fieldTypeConst(field)), " + ", valueSizeExpression(name, field, file, importAliases))
	return nil
}

func emitResetMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc reset() -> void:")
	if len(msg.Fields) <= 0 {
		g.P("\t\tpass")
		g.P()
		return nil
	}
	for _, field := range msg.Fields {
		name := safeIdentifier(field.GoName)
		defaultExpr, err := fieldDefaultValueExpression(file, field, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\t", name, " = ", defaultExpr)
	}
	g.P()
	return nil
}

func emitTypeIDMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message) {
	if !config.GapVariant {
		return
	}
	g.P("\tfunc type_id() -> int:")
	g.P("\t\treturn ", makeTypeId(string(file.Desc.Package()), string(msg.Desc.Name())))
	g.P()
}

func emitCloneMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc clone() -> ProtoMessage:")
	g.P("\t\tvar pb_msg := ", safeIdentifier(msg.GoIdent.GoName), ".new()")
	for _, field := range msg.Fields {
		if err := emitCloneField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn pb_msg")
	g.P()
	return nil
}

func emitHashToMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	g.P("\tfunc hash_to(pb_hasher: ProtoHasher) -> void:")
	g.P("\t\tProtoUtils.hash_message_fields(pb_hasher, ", len(msg.Fields), ")")
	if len(msg.Fields) <= 0 {
		g.P()
		return nil
	}
	for _, field := range msg.Fields {
		if err := emitHashToField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P()
	return nil
}

func emitHashToField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	if field.Desc.IsMap() {
		keyField := field.Message.Fields[0]
		keyHasher, err := hashCallableExpression("pb_hasher", "pb_key", file, keyField, importAliases)
		if err != nil {
			return err
		}
		valueHasher, err := hashCallableExpression("pb_hasher", "pb_value", file, field.Message.Fields[1], importAliases)
		if err != nil {
			return err
		}
		g.P("\t\tProtoUtils.hash_dictionary(pb_hasher, ", name, ", func(pb_key): ", keyHasher, ", func(pb_value): ", valueHasher, dictionaryKeyOrderSuffix(keyField), ")")
		return nil
	}
	if field.Desc.IsList() {
		valueHasher, err := hashCallableExpression("pb_hasher", "pb_value", file, field, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\tProtoUtils.hash_array(pb_hasher, ", name, ", func(pb_value): ", valueHasher, ")")
		return nil
	}
	callExpr, err := hashCallExpression("pb_hasher", name, file, field, importAliases)
	if err != nil {
		return err
	}
	g.P("\t\t", callExpr)
	return nil
}

func emitEqualsMethod(g *protogen.GeneratedFile, file *protogen.File, msg *protogen.Message, importAliases map[string]string) error {
	msgName := safeIdentifier(msg.GoIdent.GoName)
	g.P("\tfunc equals(pb_other: ProtoMessage) -> bool:")
	g.P("\t\tvar pb_other_msg := pb_other as ", msgName)
	g.P("\t\tif pb_other_msg == null:")
	g.P("\t\t\treturn false")
	for _, field := range msg.Fields {
		if err := emitEqualsField(g, file, field, importAliases); err != nil {
			return err
		}
	}
	g.P("\t\treturn true")
	g.P()
	return nil
}

func emitEqualsField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	if field.Desc.IsMap() {
		valueField := field.Message.Fields[1]
		valueEqualExpr, err := equalCallExpression("pb_a", "pb_b", valueField, file, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\tif !ProtoUtils.equal_dictionary(", name, ", pb_other_msg.", name, ", func(pb_a, pb_b): return ", valueEqualExpr, "):")
		g.P("\t\t\treturn false")
		return nil
	}
	if field.Desc.IsList() {
		valueEqualExpr, err := equalCallExpression("pb_a", "pb_b", field, file, importAliases)
		if err != nil {
			return err
		}
		g.P("\t\tif !ProtoUtils.equal_array(", name, ", pb_other_msg.", name, ", func(pb_a, pb_b): return ", valueEqualExpr, "):")
		g.P("\t\t\treturn false")
		return nil
	}
	return emitEqualsValueComparison(g, "\t\t", name, "pb_other_msg."+name, field, file, importAliases)
}

func emitEqualsValueComparison(g *protogen.GeneratedFile, indent, leftExpr, rightExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) error {
	equalExpr, err := equalCallExpression(leftExpr, rightExpr, field, file, importAliases)
	if err != nil {
		return err
	}
	g.P(indent, "if !(", equalExpr, "):")
	g.P(indent, "\treturn false")
	return nil
}

func emitCloneField(g *protogen.GeneratedFile, file *protogen.File, field *protogen.Field, importAliases map[string]string) error {
	name := safeIdentifier(field.GoName)
	if field.Desc.IsMap() {
		valueField := field.Message.Fields[1]
		g.P("\t\tfor pb_key in ", name, ":")
		if valueField.Message != nil && !valueField.Desc.IsMap() {
			g.P("\t\t\tvar pb_value := ", name, "[pb_key]")
			g.P("\t\t\tpb_msg.", name, "[pb_key] = pb_value.clone() if pb_value != null else null")
		} else {
			g.P("\t\t\tpb_msg.", name, "[pb_key] = ", name, "[pb_key]")
		}
		return nil
	}
	if field.Desc.IsList() {
		if field.Message != nil && !field.Desc.IsMap() {
			g.P("\t\tfor pb_value in ", name, ":")
			g.P("\t\t\tpb_msg.", name, ".append(pb_value.clone() if pb_value != null else null)")
			return nil
		}
		g.P("\t\tpb_msg.", name, " = ", name, ".duplicate()")
		return nil
	}
	if field.Desc.Kind() == protoreflect.BytesKind {
		g.P("\t\tpb_msg.", name, " = ", name, ".duplicate()")
		return nil
	}
	if field.Message != nil && !field.Desc.IsMap() {
		g.P("\t\tpb_msg.", name, " = ", name, ".clone() if ", name, " != null else null")
		return nil
	}
	g.P("\t\tpb_msg.", name, " = ", name)
	return nil
}

func emitScriptStaticInit(g *protogen.GeneratedFile, file *protogen.File, messages []*protogen.Message) {
	if !config.GapVariant || len(messages) <= 0 {
		return
	}
	g.P("static func _static_init() -> void:")
	if config.GapVariant {
		for _, msg := range messages {
			g.P("\tGAPVariants.register_custom_type(", makeTypeId(string(file.Desc.Package()), string(msg.Desc.Name())), ", ", safeIdentifier(msg.GoIdent.GoName), ")")
		}
	}
	g.P()
}

func fieldTypeExpression(file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Desc.IsMap() {
		keyType := "Variant"
		valueType := "Variant"
		keyType, err := fieldSingularTypeExpression(file, field.Message.Fields[0], importAliases)
		if err != nil {
			return "", err
		}
		valueType, err = fieldSingularTypeExpression(file, field.Message.Fields[1], importAliases)
		if err != nil {
			return "", err
		}
		return "Dictionary[" + keyType + ", " + valueType + "]", nil
	}
	if field.Desc.IsList() {
		itemType, err := fieldSingularTypeExpression(file, field, importAliases)
		if err != nil {
			return "", err
		}
		return "Array[" + itemType + "]", nil
	}
	return fieldSingularTypeExpression(file, field, importAliases)
}

func fieldSingularTypeExpression(file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Enum != nil {
		return fieldEnumTypeReference(file, field.Enum, importAliases)
	}
	if field.Message != nil && !field.Desc.IsMap() {
		return fieldMessageTypeReference(file, field.Message, importAliases)
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
		return "Variant", nil
	}
}

func fieldDefaultValueExpression(file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Desc.IsMap() {
		return "{}", nil
	}
	if field.Desc.IsList() {
		return "[]", nil
	}
	if field.Enum != nil {
		return fieldEnumValueReference(file, field, importAliases)
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "false", nil
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "StringName()", nil
		}
		return `""`, nil
	case protoreflect.BytesKind:
		return "PackedByteArray()", nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "0.0", nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "null", nil
	default:
		return "0", nil
	}
}

func enumStringHelperNames(enum *protogen.Enum) (string, string) {
	enumName := safeIdentifier(string(enum.Desc.Name()))
	return enumName + "_to_string", enumName + "_from_string"
}

func enumJSONHelperReference(file *protogen.File, enum *protogen.Enum, importAliases map[string]string) (string, error) {
	helperName, _ := enumStringHelperNames(enum)
	var parts []string
	if enum.Desc.ParentFile().Path() != file.Desc.Path() {
		alias, ok := importAliases[enum.Desc.ParentFile().Path()]
		if !ok {
			return "", fmt.Errorf("missing import alias for enum %q from dependency %q", enum.Desc.FullName(), enum.Desc.ParentFile().Path())
		}
		parts = append(parts, alias)
	}
	if parentClass, ok := enumContainingMessageClassName(enum); ok {
		parts = append(parts, parentClass)
	}
	parts = append(parts, helperName)
	return strings.Join(parts, "."), nil
}

func enumJSONHelperFromReference(file *protogen.File, enum *protogen.Enum, importAliases map[string]string) (string, error) {
	_, helperName := enumStringHelperNames(enum)
	var parts []string
	if enum.Desc.ParentFile().Path() != file.Desc.Path() {
		alias, ok := importAliases[enum.Desc.ParentFile().Path()]
		if !ok {
			return "", fmt.Errorf("missing import alias for enum %q from dependency %q", enum.Desc.FullName(), enum.Desc.ParentFile().Path())
		}
		parts = append(parts, alias)
	}
	if parentClass, ok := enumContainingMessageClassName(enum); ok {
		parts = append(parts, parentClass)
	}
	parts = append(parts, helperName)
	return strings.Join(parts, "."), nil
}

func enumContainingMessageClassName(enum *protogen.Enum) (string, bool) {
	parent := enum.Desc.Parent()
	var names []string
	for parent != nil {
		msg, ok := parent.(protoreflect.MessageDescriptor)
		if !ok {
			break
		}
		names = append([]string{safeIdentifier(string(msg.Name()))}, names...)
		parent = msg.Parent()
	}
	if len(names) <= 0 {
		return "", false
	}
	return strings.Join(names, "_"), true
}

func jsonToDictValueExpression(valueExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string, emitDefaultExpr, enumAsStringExpr string) (string, error) {
	if field.Enum != nil {
		helperRef, err := enumJSONHelperReference(file, field.Enum, importAliases)
		if err != nil {
			return "", err
		}
		return helperRef + "(" + valueExpr + ") if " + enumAsStringExpr + " else " + valueExpr, nil
	}
	if field.Message != nil && !field.Desc.IsMap() {
		return valueExpr + ".to_dict(" + emitDefaultExpr + ", " + enumAsStringExpr + ") if " + valueExpr + " != null else null", nil
	}
	switch field.Desc.Kind() {
	case protoreflect.BytesKind:
		return "Marshalls.raw_to_base64(" + valueExpr + ")", nil
	default:
		return valueExpr, nil
	}
}

func shouldIgnoreIncompatibleTernaryInToDict(field *protogen.Field) bool {
	return field.Enum != nil || (field.Message != nil && !field.Desc.IsMap())
}

func jsonFromDictValueExpression(valueExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) (string, error) {
	if field.Enum != nil {
		helperRef, err := enumJSONHelperFromReference(file, field.Enum, importAliases)
		if err != nil {
			return "", err
		}
		return helperRef + "(" + valueExpr + ")", nil
	}
	if field.Message != nil && !field.Desc.IsMap() {
		return "", fmt.Errorf("message field %q requires statement-based JSON decoding", field.Desc.FullName())
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool(" + valueExpr + ")", nil
	case protoreflect.StringKind:
		return valueExpr, nil
	case protoreflect.BytesKind:
		return "Marshalls.base64_to_raw(str(" + valueExpr + "))", nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "float(" + valueExpr + ")", nil
	default:
		return "int(" + valueExpr + ")", nil
	}
}

func jsonMapKeyToDictExpression(keyExpr string, field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return `"true" if ` + keyExpr + ` else "false"`
	case protoreflect.StringKind:
		return keyExpr
	default:
		return "str(" + keyExpr + ")"
	}
}

func jsonMapKeyFromDictExpression(keyExpr string, field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return `str(` + keyExpr + `).to_lower() == "true"`
	case protoreflect.StringKind:
		return keyExpr
	default:
		return "int(" + keyExpr + ")"
	}
}

func fieldMessageTypeReference(file *protogen.File, msg *protogen.Message, importAliases map[string]string) (string, error) {
	name := safeIdentifier(msg.GoIdent.GoName)
	if msg.Desc.ParentFile().Path() == file.Desc.Path() {
		return name, nil
	}
	alias, ok := importAliases[msg.Desc.ParentFile().Path()]
	if !ok {
		return "", fmt.Errorf("missing import alias for message %q from dependency %q", msg.Desc.FullName(), msg.Desc.ParentFile().Path())
	}
	return alias + "." + name, nil
}

func fieldEnumTypeReference(file *protogen.File, enum *protogen.Enum, importAliases map[string]string) (string, error) {
	name := enumQualifiedName(enum)
	if enum.Desc.ParentFile().Path() == file.Desc.Path() {
		return name, nil
	}
	alias, ok := importAliases[enum.Desc.ParentFile().Path()]
	if !ok {
		return "", fmt.Errorf("missing import alias for enum %q from dependency %q", enum.Desc.FullName(), enum.Desc.ParentFile().Path())
	}
	return alias + "." + name, nil
}

func fieldEnumValueReference(file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Enum == nil || field.Enum.Desc.Values().Len() <= 0 {
		return "0", nil
	}
	typeRef, err := fieldEnumTypeReference(file, field.Enum, importAliases)
	if err != nil {
		return "", err
	}
	return typeRef + "." + safeIdentifier(string(field.Enum.Desc.Values().Get(0).Name())), nil
}

func enumQualifiedName(enum *protogen.Enum) string {
	parts := []string{}
	parent := enum.Desc.Parent()
	for parent != nil {
		if msg, ok := parent.(protoreflect.MessageDescriptor); ok {
			parts = append([]string{safeIdentifier(string(msg.Name()))}, parts...)
			parent = msg.Parent()
			continue
		}
		break
	}
	parts = append(parts, safeIdentifier(string(enum.Desc.Name())))
	return strings.Join(parts, ".")
}

func wireTypeConst(field *protogen.Field) string {
	return "ProtoFieldDescriptor.get_field_wire_type(" + fieldTypeConst(field) + ")"
}

func fieldTypeConst(field *protogen.Field) string {
	if field.Desc.IsMap() {
		return "ProtoFieldDescriptor.FieldType.TYPE_MAP"
	}
	switch field.Desc.Kind() {
	case protoreflect.DoubleKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_DOUBLE"
	case protoreflect.FloatKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_FLOAT"
	case protoreflect.Int64Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_INT64"
	case protoreflect.Uint64Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_UINT64"
	case protoreflect.Int32Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_INT32"
	case protoreflect.Fixed64Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_FIXED64"
	case protoreflect.Fixed32Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_FIXED32"
	case protoreflect.BoolKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_BOOL"
	case protoreflect.StringKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_STRING"
	case protoreflect.GroupKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_GROUP"
	case protoreflect.MessageKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_MESSAGE"
	case protoreflect.BytesKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_BYTES"
	case protoreflect.Uint32Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_UINT32"
	case protoreflect.EnumKind:
		return "ProtoFieldDescriptor.FieldType.TYPE_ENUM"
	case protoreflect.Sfixed32Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_SFIXED32"
	case protoreflect.Sfixed64Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_SFIXED64"
	case protoreflect.Sint32Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_SINT32"
	case protoreflect.Sint64Kind:
		return "ProtoFieldDescriptor.FieldType.TYPE_SINT64"
	default:
		return "ProtoFieldDescriptor.FieldType.TYPE_MESSAGE"
	}
}

func isPackedField(field *protogen.Field) bool {
	if !field.Desc.IsList() || field.Desc.IsMap() {
		return false
	}
	if field.Message != nil || field.Desc.Kind() == protoreflect.StringKind || field.Desc.Kind() == protoreflect.BytesKind {
		return false
	}
	return true
}

func shouldSerializeExpression(valueExpr string, field *protogen.Field) string {
	if field.Message != nil && !field.Desc.IsMap() {
		return valueExpr + " != null"
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return valueExpr
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "!" + valueExpr + ".is_empty()"
	case protoreflect.FloatKind, protoreflect.DoubleKind,
		protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind,
		protoreflect.EnumKind:
		return valueExpr + " != 0"
	default:
		return valueExpr + " != null"
	}
}

func encodeValueCall(valueExpr string, field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ProtoUtils.encode_bool(pb_stream, " + valueExpr + ")"
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "ProtoUtils.encode_string_name(pb_stream, " + valueExpr + ")"
		}
		return "ProtoUtils.encode_string(pb_stream, " + valueExpr + ")"
	case protoreflect.BytesKind:
		return "ProtoUtils.encode_bytes(pb_stream, " + valueExpr + ")"
	case protoreflect.FloatKind:
		return "ProtoUtils.encode_float(pb_stream, " + valueExpr + ")"
	case protoreflect.DoubleKind:
		return "ProtoUtils.encode_double(pb_stream, " + valueExpr + ")"
	case protoreflect.Sint32Kind:
		return "ProtoUtils.encode_zigzag32(pb_stream, " + valueExpr + ")"
	case protoreflect.Sint64Kind:
		return "ProtoUtils.encode_zigzag64(pb_stream, " + valueExpr + ")"
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		return "ProtoUtils.encode_fixed32(pb_stream, " + valueExpr + ")"
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
		return "ProtoUtils.encode_fixed64(pb_stream, " + valueExpr + ")"
	default:
		return "ProtoUtils.encode_varint(pb_stream, " + valueExpr + ")"
	}
}

func decodeValueExpression(field *protogen.Field, streamName string) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ProtoUtils.decode_bool(" + streamName + ")"
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "ProtoUtils.decode_string_name(" + streamName + ")"
		}
		return "ProtoUtils.decode_string(" + streamName + ")"
	case protoreflect.BytesKind:
		return "ProtoUtils.decode_bytes(" + streamName + ")"
	case protoreflect.FloatKind:
		return "ProtoUtils.decode_float(" + streamName + ")"
	case protoreflect.DoubleKind:
		return "ProtoUtils.decode_double(" + streamName + ")"
	case protoreflect.Sint32Kind:
		return "ProtoUtils.decode_zigzag32(" + streamName + ")"
	case protoreflect.Sint64Kind:
		return "ProtoUtils.decode_zigzag64(" + streamName + ")"
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		return "ProtoUtils.decode_fixed32(" + streamName + ")"
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
		return "ProtoUtils.decode_fixed64(" + streamName + ")"
	default:
		return "ProtoUtils.decode_varint(" + streamName + ")"
	}
}

func scalarSizeExpression(valueExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ProtoUtils.SIZEOF_BOOL"
	case protoreflect.StringKind:
		return "ProtoUtils.sizeof_string(" + valueExpr + ")"
	case protoreflect.BytesKind:
		return "ProtoUtils.sizeof_bytes(" + valueExpr + ")"
	case protoreflect.FloatKind:
		return "ProtoUtils.SIZEOF_FLOAT32"
	case protoreflect.DoubleKind:
		return "ProtoUtils.SIZEOF_FLOAT64"
	case protoreflect.Sint32Kind:
		return "ProtoUtils.sizeof_zigzag32(" + valueExpr + ")"
	case protoreflect.Sint64Kind:
		return "ProtoUtils.sizeof_zigzag64(" + valueExpr + ")"
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		return "ProtoUtils.SIZEOF_FIXED32"
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
		return "ProtoUtils.SIZEOF_FIXED64"
	default:
		return "ProtoUtils.sizeof_varint(" + valueExpr + ")"
	}
}

func valueSizeExpression(valueExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) string {
	if field.Message != nil && !field.Desc.IsMap() {
		return "ProtoUtils.sizeof_message(" + valueExpr + ")"
	}
	return scalarSizeExpression(valueExpr, field, file, importAliases)
}

func hashCallExpression(hasherName, valueExpr string, file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Message != nil && !field.Desc.IsMap() {
		typeRef, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return "", err
		}
		return "ProtoUtils.hash_message(" + hasherName + ", " + valueExpr + ", func(): return " + typeRef + ".new())", nil
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "ProtoUtils.hash_bool(" + hasherName + ", " + valueExpr + ")", nil
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "ProtoUtils.hash_string_name(" + hasherName + ", " + valueExpr + ")", nil
		}
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

func hashCallableExpression(hasherName, valueExpr string, file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	return hashCallExpression(hasherName, valueExpr, file, field, importAliases)
}

func dictionaryKeyOrderSuffix(keyField *protogen.Field) string {
	switch keyField.Desc.Kind() {
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return ", ProtoUtils.DictionaryKeyOrder.UINT64"
	default:
		return ""
	}
}

func equalCallExpression(leftExpr, rightExpr string, field *protogen.Field, file *protogen.File, importAliases map[string]string) (string, error) {
	if field.Message != nil && !field.Desc.IsMap() {
		typeRef, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return "", err
		}
		return "ProtoUtils.equal_message(" + leftExpr + ", " + rightExpr + ", func(): return " + typeRef + ".new())", nil
	}
	switch field.Desc.Kind() {
	case protoreflect.FloatKind:
		return "ProtoUtils.equal_float32(" + leftExpr + ", " + rightExpr + ")", nil
	case protoreflect.DoubleKind:
		return "ProtoUtils.equal_float64(" + leftExpr + ", " + rightExpr + ")", nil
	default:
		return leftExpr + " == " + rightExpr, nil
	}
}

func defaultMapKeyExpression(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "StringName()"
		}
		return `""`
	default:
		return "0"
	}
}

func defaultMapValueExpression(file *protogen.File, field *protogen.Field, importAliases map[string]string) (string, error) {
	if field.Message != nil && !field.Desc.IsMap() {
		msgType, err := fieldMessageTypeReference(file, field.Message, importAliases)
		if err != nil {
			return "", err
		}
		return msgType + ".new()", nil
	}
	if field.Enum != nil {
		return fieldEnumValueReference(file, field, importAliases)
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "false", nil
	case protoreflect.StringKind:
		if config.StringAsStringName {
			return "StringName()", nil
		}
		return `""`, nil
	case protoreflect.BytesKind:
		return "PackedByteArray()", nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "0.0", nil
	default:
		return "0", nil
	}
}

func messageBaseType() string {
	if config.GapVariant {
		return "ProtoGAPVariant"
	}
	return "ProtoMessage"
}

func makeTypeId(pkgName, msgName string) variant.TypeId {
	hash := fnv.New32a()
	hash.Write([]byte(pkgName + "." + msgName))
	return variant.TypeId(variant.TypeId_Customize + hash.Sum32())
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

func isGDScriptKeyword(s string) bool {
	switch s {
	case "and", "as", "assert", "await", "break", "class", "class_name", "const", "continue",
		"elif", "else", "enum", "extends", "false", "for", "func", "if", "in", "is", "match",
		"namespace", "not", "null", "or", "pass", "return", "self", "signal", "static", "super",
		"tool", "true", "var", "while", "yield":
		return true

	case "serialize", "deserialize", "to_dict", "from_dict", "size", "reset", "new", "clone",
		"hash_to", "equals", "type_id",
		"json_dict", "json_emit_default", "json_enum_as_string",
		"pb_stream", "pb_tag", "pb_field_number", "pb_wire_type", "pb_key", "pb_value",
		"pb_data_size", "pb_entry_size", "pb_entry_stream", "pb_entry_key", "pb_entry_value",
		"pb_entry_tag", "pb_entry_field_number", "pb_entry_wire_type", "pb_packed_size",
		"pb_packed_stream", "pb_msg_size", "pb_msg", "pb_hasher", "pb_other",
		"pb_other_msg", "pb_a", "pb_b", "pb_array", "pb_dict", "pb_field":
		return true

	default:
		return false
	}
}

func tagSizeLiteral(fieldNumber int, fieldType string) string {
	wireType := 0
	switch fieldType {
	case "ProtoFieldDescriptor.FieldType.TYPE_DOUBLE", "ProtoFieldDescriptor.FieldType.TYPE_FIXED64", "ProtoFieldDescriptor.FieldType.TYPE_SFIXED64":
		wireType = 1
	case "ProtoFieldDescriptor.FieldType.TYPE_STRING", "ProtoFieldDescriptor.FieldType.TYPE_MESSAGE", "ProtoFieldDescriptor.FieldType.TYPE_BYTES", "ProtoFieldDescriptor.FieldType.TYPE_MAP":
		wireType = 2
	case "ProtoFieldDescriptor.FieldType.TYPE_FLOAT", "ProtoFieldDescriptor.FieldType.TYPE_FIXED32", "ProtoFieldDescriptor.FieldType.TYPE_SFIXED32":
		wireType = 5
	default:
		wireType = 0
	}
	return strconv.Itoa(varintSize((fieldNumber << 3) | wireType))
}

func varintSize(value int) int {
	if value < 0 {
		return 10
	}
	size := 1
	for value >= 0x80 {
		value >>= 7
		size++
	}
	return size
}
