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
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/scaffold/tools/excelc/excelutils"
	"github.com/elliotchance/pie/v2"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/yaml.v3"
)

func genProtoMessage(file *excelize.File) proto.Message {
	sheets := slices.DeleteFunc(file.GetSheetList(), func(sheet string) bool {
		return sheet == "" || !unicode.IsLetter(rune(sheet[0]))
	})
	if len(sheets) <= 0 {
		return nil
	}

	pbTypes := protoregistry.GlobalTypes

	extensions, err := parseExtensions(pbTypes)
	if err != nil {
		log.Panicf("read excel file %q failed, %s", file.Path, err)
	}

	var columnsType, tableType protoreflect.MessageType
	var tableMsg protoreflect.Message
	var tableHashUniqueIndexes, tableSortedUniqueIndexes generic.UnorderedSliceMap[string, []protoreflect.FieldDescriptor]
	var tableHashIndexes, tableSortedIndexes generic.UnorderedSliceMap[string, []protoreflect.FieldDescriptor]
	tableSortedUniqueIndexesData := map[string]*generic.SliceMap[uint64, uint32]{}
	type sortedIndexEntry struct {
		Value  uint64
		Offset uint32
	}
	tableSortedIndexesData := map[string][]sortedIndexEntry{}

	type Column struct {
		Name  string
		Index int
		Meta  string
		Field protoreflect.FieldDescriptor
	}
	var definitionColumns []*Column
	var definitionColumnsByName map[string]*Column
	var definitionFieldsByName map[string]protoreflect.FieldDescriptor

	type OffsetLine struct {
		Sheet string
		Line  int
	}

	var offsetLines []OffsetLine

	for sheetIndex, sheet := range sheets {
		func() {
			rows, err := file.Rows(sheet)
			if err != nil {
				log.Panicf("read excel file %q sheet %q failed, %s", file.Path, sheet, err)
			}
			defer rows.Close()

			var columns []*Column
			definitionSheet := sheetIndex == 0

			for i := 1; rows.Next(); i++ {
				if i < SheetTableHeader+SheetTableHeaderSize {
					switch i {
					case 1:
						row, err := rows.Columns()
						if err != nil {
							log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, sheet, i, err)
						}

						cells := Cells(row)
						for j := range row {
							columns = append(columns, &Column{
								Name:  snake2Camel(cells.Get(j)),
								Index: j,
							})
						}

						for i, col := range columns {
							if col.Name == "" {
								columns = columns[:i]
								break
							}
						}

						columns = slices.DeleteFunc(columns, func(decl *Column) bool {
							return decl.Name == "" || !unicode.IsLetter(rune(decl.Name[0]))
						})

						if len(columns) <= 0 {
							return
						}

						columnsByName := make(map[string]*Column, len(columns))
						for _, column := range columns {
							if previous := columnsByName[column.Name]; previous != nil {
								log.Panicf("read excel file %q sheet %q row %d failed: duplicate column %q at columns %d and %d", file.Path, sheet, i, column.Name, previous.Index+1, column.Index+1)
							}
							columnsByName[column.Name] = column
						}

						if definitionSheet {
							definitionColumns = columns
							definitionColumnsByName = columnsByName
						} else {
							for _, column := range columns {
								if definitionColumnsByName[column.Name] == nil {
									log.Panicf("read excel file %q sheet %q row %d column %d failed: column %q is not defined in first data sheet %q", file.Path, sheet, i, column.Index+1, column.Name, sheets[0])
								}
							}
						}

					case SheetTableColumnMeta:
						if !definitionSheet {
							continue
						}

						row, err := rows.Columns()
						if err != nil {
							log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, sheet, i, err)
						}

						cells := Cells(row)
						for _, column := range columns {
							if column.Index < 0 || column.Index >= len(row) {
								continue
							}
							column.Meta = cells.Get(column.Index)
						}
					}
					continue
				}

				if columnsType == nil {
					columnsName := fmt.Sprintf("%s.%s", viper.GetString("pb_package"), snake2Camel(strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path)))+"Columns")

					columnsType, err = pbTypes.FindMessageByName(protoreflect.FullName(columnsName))
					if err != nil {
						log.Panicf("parse proto type %q failed, %s", columnsName, err)
					}
				}

				if tableType == nil {
					tableName := fmt.Sprintf("%s.%s", viper.GetString("pb_package"), snake2Camel(strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path)))+"Table")

					tableType, err = pbTypes.FindMessageByName(protoreflect.FullName(tableName))
					if err != nil {
						log.Panicf("parse proto type %q failed, %s", tableName, err)
					}

					for j := range tableType.Descriptor().Fields().Len() {
						field := tableType.Descriptor().Fields().Get(j)

						indexTypeValue, ok := proto.GetExtension(field.Options(), extensions.IndexType).(string)
						if !ok || indexTypeValue == "" {
							continue
						}
						indexKind := indexType(indexTypeValue)

						indexFields := proto.GetExtension(field.Options(), extensions.IndexFields).(string)
						if indexFields == "" {
							continue
						}

						fieldDescs := make([]protoreflect.FieldDescriptor, 0, len(strings.Split(indexFields, ",")))
						for _, indexFieldName := range strings.Split(indexFields, ",") {
							fieldDesc := columnsType.Descriptor().Fields().ByName(protoreflect.Name(indexFieldName))
							if fieldDesc == nil {
								log.Panicf("parse proto type %q failed, index field %q not found", columnsType.Descriptor().FullName(), indexFieldName)
							}
							fieldDescs = append(fieldDescs, fieldDesc)
						}

						switch indexKind {
						case indexTypeHashUnique:
							tableHashUniqueIndexes.Add(string(field.Name()), fieldDescs)
						case indexTypeSortedUnique:
							tableSortedUniqueIndexes.Add(string(field.Name()), fieldDescs)
						case indexTypeHash:
							tableHashIndexes.Add(string(field.Name()), fieldDescs)
						case indexTypeSorted:
							tableSortedIndexes.Add(string(field.Name()), fieldDescs)
						default:
							log.Panicf("parse proto field %q failed, unsupported index type %q", field.FullName(), indexKind)
						}
					}

					tableMsg = tableType.New()
				}

				if definitionFieldsByName == nil {
					definitionFieldsByName = make(map[string]protoreflect.FieldDescriptor, len(definitionColumns))
					for columnIdx, column := range definitionColumns {
						meta, err := parseMeta(column.Meta)
						if err != nil {
							log.Panicf("read excel file %q sheet %q failed: parse meta %q for column %q failed, %s", file.Path, sheets[0], column.Meta, column.Name, err)
						}
						if !meta.MatchTargets() {
							continue
						}

						field := columnsType.Descriptor().Fields().ByName(protoreflect.Name(column.Name))
						if field == nil {
							log.Panicf("parse proto type %q failed: column %q from first data sheet %q was not found", columnsType.Descriptor().FullName(), column.Name, sheets[0])
						}

						expectedFieldNumber := protoreflect.FieldNumber(columnIdx + 1)
						if meta.PbFieldNumber != nil {
							expectedFieldNumber = protoreflect.FieldNumber(*meta.PbFieldNumber)
						}
						if field.Number() != expectedFieldNumber {
							log.Panicf("parse proto field %q failed: field number is %d, but first data sheet %q configures %d", field.FullName(), field.Number(), sheets[0], expectedFieldNumber)
						}

						definitionFieldsByName[column.Name] = field
					}

					for fieldIdx := range columnsType.Descriptor().Fields().Len() {
						field := columnsType.Descriptor().Fields().Get(fieldIdx)
						if definitionFieldsByName[string(field.Name())] == nil {
							log.Panicf("parse proto type %q failed: field %q is not defined in first data sheet %q", columnsType.Descriptor().FullName(), field.Name(), sheets[0])
						}
					}
				}

				for _, column := range columns {
					column.Field = definitionFieldsByName[column.Name]
				}

				row, err := rows.Columns()
				if err != nil {
					log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, sheet, i, err)
				}
				cells := Cells(row)

				if func() bool {
					for _, column := range columns {
						if cells.Get(column.Index) != "" {
							return false
						}
					}
					return true
				}() {
					continue
				}

				rowMsg := columnsType.New()

				for _, column := range columns {
					if column.Field == nil {
						continue
					}

					if err := setFieldFromString(rowMsg, column.Field, cells.Get(column.Index), extensions); err != nil {
						log.Panicf("read excel file %q sheet %q row %d column %q failed, %s", file.Path, sheet, i, column.Field.Name(), err)
					}
				}

				tableRows := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName("Rows"))
				offset := uint32(tableRows.List().Len())
				offsetLines = append(offsetLines, OffsetLine{
					Sheet: sheet,
					Line:  i,
				})
				tableRows.List().Append(protoreflect.ValueOf(rowMsg))

				tableHashUniqueIndexes.Each(func(indexName string, fields []protoreflect.FieldDescriptor) {
					tableIndex := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName)))

					if len(fields) > 1 || excelutils.ProtoMessageFieldNeedHashIndex(fields[0]) {
						h := excelutils.NewHash()

						for _, fieldDesc := range fields {
							fieldValue := rowMsg.Get(fieldDesc)

							if err := excelutils.AnyToHash(h, fieldValue); err != nil {
								log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
							}
						}

						key := protoreflect.ValueOfUint64(h.Sum64()).MapKey()

						if existed := tableIndex.Map().Get(key); existed.IsValid() {
							duplicateOffset, duplicated := findIndexDuplicateOffset(tableMsg, indexName, key, uint32(existed.Uint()), rowMsg, fields)
							if duplicated {
								conflictedRow := offsetLines[duplicateOffset]
								log.Panicf("read excel file %q sheet %q row %d failed: index %q value %d conflicts with sheet %q row %d", file.Path, sheet, i, indexName, h.Sum64(), conflictedRow.Sheet, conflictedRow.Line)
							}

							log.Printf("read excel file %q sheet %q row %d warning: index %q value %d collides with sheet %q row %d; stored in collision bucket", file.Path, sheet, i, indexName, h.Sum64(), offsetLines[existed.Uint()].Sheet, offsetLines[existed.Uint()].Line)
							appendIndexOffset(tableMsg, indexName+"Collisions", key, offset)
							return
						}

						tableIndex.Map().Set(key, protoreflect.ValueOfUint32(offset))

					} else {
						indexValue, err := excelutils.ProtoMessageFieldToIndex(rowMsg, fields[0])
						if err != nil {
							log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
						}

						key := protoreflect.ValueOfUint64(indexValue).MapKey()

						if existed := tableIndex.Map().Get(key); existed.IsValid() {
							conflictedRow := offsetLines[existed.Uint()]
							log.Panicf("read excel file %q sheet %q row %d failed: index %q value %d conflicts with sheet %q row %d", file.Path, sheet, i, indexName, indexValue, conflictedRow.Sheet, conflictedRow.Line)
						}

						tableIndex.Map().Set(protoreflect.ValueOfUint64(indexValue).MapKey(), protoreflect.ValueOfUint32(offset))
					}
				})

				tableSortedUniqueIndexes.Each(func(indexName string, fields []protoreflect.FieldDescriptor) {
					indexData, ok := tableSortedUniqueIndexesData[indexName]
					if !ok {
						indexData = &generic.SliceMap[uint64, uint32]{}
						tableSortedUniqueIndexesData[indexName] = indexData
					}

					if len(fields) > 1 || excelutils.ProtoMessageFieldNeedHashIndex(fields[0]) {
						h := excelutils.NewHash()

						for _, fieldDesc := range fields {
							fieldValue := rowMsg.Get(fieldDesc)

							if err := excelutils.AnyToHash(h, fieldValue); err != nil {
								log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
							}
						}

						if existed, ok := indexData.Get(h.Sum64()); ok {
							key := protoreflect.ValueOfUint64(h.Sum64()).MapKey()

							duplicateOffset, duplicated := findIndexDuplicateOffset(tableMsg, indexName, key, existed, rowMsg, fields)
							if duplicated {
								conflictedRow := offsetLines[duplicateOffset]
								log.Panicf("read excel file %q sheet %q row %d failed: index %q value %d conflicts with sheet %q row %d", file.Path, sheet, i, indexName, h.Sum64(), conflictedRow.Sheet, conflictedRow.Line)
							}

							log.Printf("read excel file %q sheet %q row %d warning: index %q value %d collides with sheet %q row %d; stored in collision bucket", file.Path, sheet, i, indexName, h.Sum64(), offsetLines[existed].Sheet, offsetLines[existed].Line)
							appendIndexOffset(tableMsg, indexName+"Collisions", key, offset)
							return
						}

						indexData.Add(h.Sum64(), offset)

					} else {
						indexValue, err := excelutils.ProtoMessageFieldToIndex(rowMsg, fields[0])
						if err != nil {
							log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
						}

						if existed, ok := indexData.Get(indexValue); ok {
							conflictedRow := offsetLines[existed]
							log.Panicf("read excel file %q sheet %q row %d failed: index %q value %d conflicts with sheet %q row %d", file.Path, sheet, i, indexName, indexValue, conflictedRow.Sheet, conflictedRow.Line)
						}

						indexData.Add(indexValue, offset)
					}
				})

				tableHashIndexes.Each(func(indexName string, fields []protoreflect.FieldDescriptor) {
					indexValue, err := computeIndexValue(rowMsg, fields)
					if err != nil {
						log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
					}

					appendIndexOffset(
						tableMsg,
						indexName,
						protoreflect.ValueOfUint64(indexValue).MapKey(),
						offset,
					)
				})

				tableSortedIndexes.Each(func(indexName string, fields []protoreflect.FieldDescriptor) {
					indexValue, err := computeIndexValue(rowMsg, fields)
					if err != nil {
						log.Panicf("read excel file %q sheet %q row %d failed: compute index %q value failed, %s", file.Path, sheet, i, indexName, err)
					}

					tableSortedIndexesData[indexName] = append(tableSortedIndexesData[indexName], sortedIndexEntry{
						Value:  indexValue,
						Offset: offset,
					})
				})
			}
		}()
	}

	if tableMsg == nil {
		return nil
	}

	tableSortedUniqueIndexes.Each(func(indexName string, _ []protoreflect.FieldDescriptor) {
		tableIndexField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName))
		if tableIndexField == nil {
			log.Panicf("parse proto type %q failed: index field %q not found", tableMsg.Descriptor().FullName(), indexName)
		}

		tableIndex := tableMsg.Mutable(tableIndexField).Message()
		valuesField := tableIndex.Descriptor().Fields().ByName("Values")
		if valuesField == nil {
			log.Panicf("parse proto type %q failed: field %q not found", tableIndex.Descriptor().FullName(), "Values")
		}
		offsetsField := tableIndex.Descriptor().Fields().ByName("Offsets")
		if offsetsField == nil {
			log.Panicf("parse proto type %q failed: field %q not found", tableIndex.Descriptor().FullName(), "Offsets")
		}

		indexData, ok := tableSortedUniqueIndexesData[indexName]
		if !ok {
			return
		}

		indexData.Each(func(value uint64, offset uint32) {
			tableIndex.Mutable(valuesField).List().Append(protoreflect.ValueOfUint64(value))
			tableIndex.Mutable(offsetsField).List().Append(protoreflect.ValueOfUint32(offset))
		})
	})

	tableSortedIndexes.Each(func(indexName string, _ []protoreflect.FieldDescriptor) {
		tableIndexField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName))
		if tableIndexField == nil {
			log.Panicf("parse proto type %q failed: index field %q not found", tableMsg.Descriptor().FullName(), indexName)
		}

		tableIndex := tableMsg.Mutable(tableIndexField).Message()
		valuesField := tableIndex.Descriptor().Fields().ByName("Values")
		if valuesField == nil {
			log.Panicf("parse proto type %q failed: field %q not found", tableIndex.Descriptor().FullName(), "Values")
		}
		startsField := tableIndex.Descriptor().Fields().ByName("Starts")
		if startsField == nil {
			log.Panicf("parse proto type %q failed: field %q not found", tableIndex.Descriptor().FullName(), "Starts")
		}
		offsetsField := tableIndex.Descriptor().Fields().ByName("Offsets")
		if offsetsField == nil {
			log.Panicf("parse proto type %q failed: field %q not found", tableIndex.Descriptor().FullName(), "Offsets")
		}

		entries := tableSortedIndexesData[indexName]
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].Value < entries[j].Value
		})

		values := tableIndex.Mutable(valuesField).List()
		starts := tableIndex.Mutable(startsField).List()
		offsets := tableIndex.Mutable(offsetsField).List()
		starts.Append(protoreflect.ValueOfUint32(0))

		for i, entry := range entries {
			if i == 0 || entry.Value != entries[i-1].Value {
				if i > 0 {
					starts.Append(protoreflect.ValueOfUint32(uint32(i)))
				}
				values.Append(protoreflect.ValueOfUint64(entry.Value))
			}
			offsets.Append(protoreflect.ValueOfUint32(entry.Offset))
		}
		if len(entries) > 0 {
			starts.Append(protoreflect.ValueOfUint32(uint32(len(entries))))
		}
	})

	return tableMsg.Interface()
}

func computeIndexValue(rowMsg protoreflect.Message, fields []protoreflect.FieldDescriptor) (uint64, error) {
	if len(fields) <= 0 {
		return 0, errors.New("index fields cannot be empty")
	}

	if len(fields) == 1 && !excelutils.ProtoMessageFieldNeedHashIndex(fields[0]) {
		return excelutils.ProtoMessageFieldToIndex(rowMsg, fields[0])
	}

	h := excelutils.NewHash()
	for _, field := range fields {
		if err := excelutils.AnyToHash(h, rowMsg.Get(field)); err != nil {
			return 0, err
		}
	}
	return h.Sum64(), nil
}

func appendIndexOffset(tableMsg protoreflect.Message, indexName string, key protoreflect.MapKey, offset uint32) {
	indexField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName))
	if indexField == nil {
		log.Panicf("parse proto type %q failed: index field %q not found", tableMsg.Descriptor().FullName(), indexName)
	}

	bucket := tableMsg.Mutable(indexField).Map().Mutable(key).Message()

	offsetsField := bucket.Descriptor().Fields().ByName("Offsets")
	if offsetsField == nil {
		log.Panicf("parse proto type %q failed: field %q not found", bucket.Descriptor().FullName(), "Offsets")
	}

	bucket.Mutable(offsetsField).List().Append(protoreflect.ValueOfUint32(offset))
}

func findIndexDuplicateOffset(tableMsg protoreflect.Message, indexName string, key protoreflect.MapKey, primaryOffset uint32, rowMsg protoreflect.Message, fields []protoreflect.FieldDescriptor) (uint32, bool) {
	tableRows := tableMsg.Get(tableMsg.Descriptor().Fields().ByName("Rows")).List()
	if excelutils.ProtoMessageFieldsEqual(tableRows.Get(int(primaryOffset)).Message(), rowMsg, fields...) {
		return primaryOffset, true
	}

	collisionField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName + "Collisions"))
	if collisionField == nil {
		return 0, false
	}

	collisionBucket := tableMsg.Get(collisionField).Map().Get(key)
	if !collisionBucket.IsValid() {
		return 0, false
	}

	offsetsField := collisionBucket.Message().Descriptor().Fields().ByName("Offsets")
	if offsetsField == nil {
		return 0, false
	}

	offsets := collisionBucket.Message().Get(offsetsField).List()
	for i := 0; i < offsets.Len(); i++ {
		offset := uint32(offsets.Get(i).Uint())
		if excelutils.ProtoMessageFieldsEqual(tableRows.Get(int(offset)).Message(), rowMsg, fields...) {
			return offset, true
		}
	}

	return 0, false
}

func setFieldFromString(msg protoreflect.Message, field protoreflect.FieldDescriptor, value string, extensions *Extensions) error {
	if !matchTargets(field, extensions) {
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	if field.Kind() != protoreflect.MessageKind {
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			values, err := parseYAMLScalarListValue(value, sep)
			if err != nil {
				return err
			}
			return appendScalarListValues(msg, field, values, extensions)
		}

		value, err := decodeYAMLQuotedScalar(value)
		if err != nil {
			return err
		}
		fieldValue, err := parseScalarFieldValue(field, value, extensions)
		if err != nil {
			return err
		}
		msg.Set(field, fieldValue)
		return nil
	}

	switch {
	case field.IsMap():
		fieldValue, err := parseStructValue(value)
		if err != nil {
			return err
		}
		if fieldValue.Kind != yaml.DocumentNode || len(fieldValue.Content) == 0 {
			return nil
		}

		fieldValue = fieldValue.Content[0]
		if fieldValue.Kind != yaml.MappingNode {
			return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to a map type", value)
		}
		return setFieldMappingValue(msg, field, fieldValue, extensions)

	case field.IsList():
		fieldValue, err := parseYAMLValue(value)
		if err == nil && fieldValue.Kind == yaml.SequenceNode {
			return appendMessageListValues(msg, field, fieldValue.Content, extensions)
		}

		sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
		values, err := splitYAMLListValue(value, sep)
		if err != nil {
			return err
		}

		fieldValues := make([]*yaml.Node, 0, len(values))
		for _, item := range values {
			childValue, err := parseStructValue(item)
			if err != nil {
				return err
			}
			if childValue.Kind != yaml.DocumentNode || len(childValue.Content) == 0 {
				continue
			}

			childValue = childValue.Content[0]
			if childValue.Kind != yaml.MappingNode {
				return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to an object type", childValue.Value)
			}
			fieldValues = append(fieldValues, childValue)
		}
		return appendMessageListValues(msg, field, fieldValues, extensions)

	default:
		fieldValue, err := parseStructValue(value)
		if err != nil {
			return err
		}

		if fieldValue.Kind != yaml.DocumentNode || len(fieldValue.Content) <= 0 {
			return nil
		}
		fieldValue = fieldValue.Content[0]

		if fieldValue.Kind != yaml.MappingNode {
			return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to an object type", value)
		}

		return setFieldStructValue(msg, field, fieldValue, extensions)
	}
}

func parseScalarFieldValue(field protoreflect.FieldDescriptor, value string, extensions *Extensions) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.BoolKind:
		v, err := strconv.ParseBool(value)
		return protoreflect.ValueOfBool(v), err
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		v, err := strconv.ParseInt(value, 10, 32)
		return protoreflect.ValueOfInt32(int32(v)), err
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		v, err := strconv.ParseInt(value, 10, 64)
		return protoreflect.ValueOfInt64(v), err
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		v, err := strconv.ParseUint(value, 10, 32)
		return protoreflect.ValueOfUint32(uint32(v)), err
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		v, err := strconv.ParseUint(value, 10, 64)
		return protoreflect.ValueOfUint64(v), err
	case protoreflect.FloatKind:
		v, err := strconv.ParseFloat(value, 32)
		return protoreflect.ValueOfFloat32(float32(v)), err
	case protoreflect.DoubleKind:
		v, err := strconv.ParseFloat(value, 64)
		return protoreflect.ValueOfFloat64(v), err
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(value), nil
	case protoreflect.BytesKind:
		v, err := base64.URLEncoding.DecodeString(value)
		return protoreflect.ValueOfBytes(v), err
	case protoreflect.EnumKind:
		return parseEnumValue(field.Enum(), value, extensions)
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported scalar field kind %v", field.Kind())
	}
}

func appendScalarListValues(msg protoreflect.Message, field protoreflect.FieldDescriptor, values []string, extensions *Extensions) error {
	fieldValues := make([]protoreflect.Value, 0, len(values))
	for _, value := range values {
		fieldValue, err := parseScalarFieldValue(field, value, extensions)
		if err != nil {
			return err
		}
		fieldValues = append(fieldValues, fieldValue)
	}

	list := msg.Mutable(field).List()
	for _, fieldValue := range fieldValues {
		list.Append(fieldValue)
	}
	return nil
}

func appendMessageListValues(msg protoreflect.Message, field protoreflect.FieldDescriptor, values []*yaml.Node, extensions *Extensions) error {
	fieldType, err := protoregistry.GlobalTypes.FindMessageByName(field.Message().FullName())
	if err != nil {
		return err
	}

	fieldValues := make([]protoreflect.Message, 0, len(values))
	for _, value := range values {
		fieldValue, err := makeStructValue(fieldType, value, extensions)
		if err != nil {
			return err
		}
		fieldValues = append(fieldValues, fieldValue)
	}

	list := msg.Mutable(field).List()
	for _, fieldValue := range fieldValues {
		list.Append(protoreflect.ValueOf(fieldValue))
	}
	return nil
}

func setFieldStructValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, fieldValue *yaml.Node, extensions *Extensions) error {
	if !matchTargets(field, extensions) {
		return nil
	}

	if field.IsMap() {
		if fieldValue.Kind != yaml.MappingNode {
			return errors.New("field value is not a mapping node")
		}
		return setFieldMappingValue(msg, field, fieldValue, extensions)
	}

	if field.Kind() != protoreflect.MessageKind {
		if field.IsList() {
			switch fieldValue.Kind {
			case yaml.SequenceNode:
				values := make([]string, 0, len(fieldValue.Content))
				for _, c := range fieldValue.Content {
					if c.Kind != yaml.ScalarNode {
						return errors.New("field value is not a scalar node")
					}
					values = append(values, c.Value)
				}
				return appendScalarListValues(msg, field, values, extensions)

			case yaml.ScalarNode:
				sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
				values, err := parseYAMLScalarListValue(fieldValue.Value, sep)
				if err != nil {
					return err
				}
				return appendScalarListValues(msg, field, values, extensions)

			default:
				return errors.New("field value is not a sequence or scalar node")
			}

		} else {
			if fieldValue.Kind != yaml.ScalarNode {
				return errors.New("field value is not a scalar node")
			}
			value, err := parseScalarFieldValue(field, fieldValue.Value, extensions)
			if err != nil {
				return err
			}
			msg.Set(field, value)
			return nil
		}
	}

	if field.IsList() {
		switch fieldValue.Kind {
		case yaml.SequenceNode:
			return appendMessageListValues(msg, field, fieldValue.Content, extensions)

		case yaml.MappingNode:
			return appendMessageListValues(msg, field, []*yaml.Node{fieldValue}, extensions)

		case yaml.ScalarNode:
			return setFieldFromString(msg, field, fieldValue.Value, extensions)

		default:
			return errors.New("field value is not a sequence, mapping, or scalar node")
		}

	} else {
		fieldType, err := protoregistry.GlobalTypes.FindMessageByName(field.Message().FullName())
		if err != nil {
			return err
		}
		fieldMsg, err := makeStructValue(fieldType, fieldValue, extensions)
		if err != nil {
			return err
		}
		msg.Set(field, protoreflect.ValueOf(fieldMsg))
	}

	return nil
}

func setFieldMappingValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, value *yaml.Node, extensions *Extensions) error {
	if !field.IsMap() {
		return errors.New("field type not mapping")
	}

	if value.Kind != yaml.MappingNode {
		return errors.New("field value is not a mapping node")
	}

	mapping := msg.Mutable(field).Map()
	kType := field.MapKey()
	vType := field.MapValue()

	if len(value.Content)%2 != 0 {
		return errors.New("field value is not a valid mapping node")
	}
	if kType == nil || kType.Kind() == protoreflect.MessageKind || kType.IsList() || kType.IsMap() {
		return errors.New("field map key type is invalid")
	}
	if vType == nil {
		return errors.New("field map value type is invalid")
	}

	for i := 0; i < len(value.Content); i += 2 {
		k := value.Content[i]
		v := value.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return errors.New("field map key is not a scalar node")
		}
		if vType.Kind() != protoreflect.MessageKind && v.Kind != yaml.ScalarNode {
			return errors.New("field map value is not a scalar node")
		}

		var ik, iv any

		switch kType.Kind() {
		case protoreflect.BoolKind:
			b, err := strconv.ParseBool(k.Value)
			if err != nil {
				return err
			}
			ik = b
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			n, err := strconv.ParseInt(k.Value, 10, 32)
			if err != nil {
				return err
			}
			ik = int32(n)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			n, err := strconv.ParseInt(k.Value, 10, 64)
			if err != nil {
				return err
			}
			ik = n
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			n, err := strconv.ParseUint(k.Value, 10, 32)
			if err != nil {
				return err
			}
			ik = uint32(n)
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			n, err := strconv.ParseUint(k.Value, 10, 64)
			if err != nil {
				return err
			}
			ik = n
		case protoreflect.StringKind:
			ik = k.Value
		default:
			return fmt.Errorf("unsupported map key kind %v", kType.Kind())
		}

		switch vType.Kind() {
		case protoreflect.BoolKind:
			b, err := strconv.ParseBool(v.Value)
			if err != nil {
				return err
			}
			iv = b
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			n, err := strconv.ParseInt(v.Value, 10, 32)
			if err != nil {
				return err
			}
			iv = int32(n)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			n, err := strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				return err
			}
			iv = n
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			n, err := strconv.ParseUint(v.Value, 10, 32)
			if err != nil {
				return err
			}
			iv = uint32(n)
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			n, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return err
			}
			iv = n
		case protoreflect.FloatKind:
			n, err := strconv.ParseFloat(v.Value, 32)
			if err != nil {
				return err
			}
			iv = float32(n)
		case protoreflect.DoubleKind:
			n, err := strconv.ParseFloat(v.Value, 64)
			if err != nil {
				return err
			}
			iv = n
		case protoreflect.StringKind:
			iv = v.Value
		case protoreflect.BytesKind:
			bs, err := base64.URLEncoding.DecodeString(v.Value)
			if err != nil {
				return err
			}
			iv = bs
		case protoreflect.EnumKind:
			enumValue, err := parseEnumValue(vType.Enum(), v.Value, extensions)
			if err != nil {
				return err
			}
			iv = enumValue.Enum()
		case protoreflect.MessageKind:
			if v.Kind != yaml.MappingNode {
				return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to an object type", v.Value)
			}

			ty, err := protoregistry.GlobalTypes.FindMessageByName(vType.Message().FullName())
			if err != nil {
				return err
			}

			iv, err = makeStructValue(ty, v, extensions)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported map value kind %v", vType.Kind())
		}

		mapping.Set(protoreflect.ValueOf(ik).MapKey(), protoreflect.ValueOf(iv))
	}

	return nil
}

func parseEnumValue(enumDesc protoreflect.EnumDescriptor, value string, extensions *Extensions) (protoreflect.Value, error) {
	enumValueDesc := enumDesc.Values().ByName(protoreflect.Name(value))
	if enumValueDesc != nil {
		return protoreflect.ValueOfEnum(enumValueDesc.Number()), nil
	}

	enumNum, err := strconv.Atoi(value)
	if err == nil {
		enumValueDesc := enumDesc.Values().ByNumber(protoreflect.EnumNumber(enumNum))
		if enumValueDesc != nil {
			return protoreflect.ValueOfEnum(enumValueDesc.Number()), nil
		}
	}

	for i := range enumDesc.Values().Len() {
		enumValueDesc := enumDesc.Values().Get(i)
		enumValueAlias := proto.GetExtension(enumValueDesc.Options(), extensions.EnumValueAlias).(string)

		if enumValueAlias == value {
			return protoreflect.ValueOfEnum(enumValueDesc.Number()), nil
		}
	}

	return protoreflect.Value{}, fmt.Errorf("unsupported enum value %q", value)
}

func parseStructValue(value string) (*yaml.Node, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") {
		if !strings.HasPrefix(value, "{") {
			value = "{\n" + value + "\n}"
		}
	}

	return parseYAMLDocument(value)
}

func parseYAMLDocument(value string) (*yaml.Node, error) {
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(value), node); err != nil {
		return nil, err
	}
	if err := validateYAMLDuplicateKeys(node); err != nil {
		return nil, err
	}

	return node, nil
}

func parseYAMLValue(value string) (*yaml.Node, error) {
	node, err := parseYAMLDocument(value)
	if err != nil {
		return nil, err
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return nil, errors.New("YAML value has no document content")
	}

	return node.Content[0], nil
}

func makeStructValue(ty protoreflect.MessageType, value *yaml.Node, extensions *Extensions) (protoreflect.Message, error) {
	if value.Kind != yaml.MappingNode {
		return nil, errors.New("field value is not a mapping node")
	}

	msg := ty.New()
	if err := validateYAMLObjectKeys(value); err != nil {
		return nil, err
	}

	for i := range msg.Descriptor().Fields().Len() {
		field := msg.Descriptor().Fields().Get(i)

		fieldValue := findYAMLMappingValue(value, string(field.Name()))
		if fieldValue == nil {
			fieldAlias := proto.GetExtension(field.Options(), extensions.FieldAlias).(string)
			if fieldAlias != "" {
				fieldValue = findYAMLMappingValue(value, fieldAlias)
			}
			if fieldValue == nil {
				continue
			}
		}

		err := setFieldStructValue(msg, field, fieldValue, extensions)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func findYAMLMappingValue(value *yaml.Node, key string) *yaml.Node {
	if value.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
			return value.Content[i+1]
		}
	}

	return nil
}

func validateYAMLObjectKeys(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(value.Content); i += 2 {
		key := value.Content[i]
		if key.Kind != yaml.ScalarNode || !strings.ContainsRune(key.Value, ':') {
			continue
		}
		return fmt.Errorf("YAML object key %q contains ':'; add whitespace after ':' when assigning a value", key.Value)
	}

	return nil
}

func validateYAMLDuplicateKeys(value *yaml.Node) error {
	if value.Kind == yaml.MappingNode {
		keys := make(map[string]*yaml.Node, len(value.Content)/2)
		for i := 0; i+1 < len(value.Content); i += 2 {
			key := value.Content[i]
			if key.Kind == yaml.ScalarNode {
				keyID := key.Tag + "\x00" + key.Value
				if previous := keys[keyID]; previous != nil {
					return fmt.Errorf(
						"YAML mapping contains duplicate key %q; positions within the cell: first occurrence at line %d, column %d, duplicate occurrence at line %d, column %d",
						key.Value, previous.Line, previous.Column, key.Line, key.Column,
					)
				}
				keys[keyID] = key
			}
		}
	}

	for _, child := range value.Content {
		if err := validateYAMLDuplicateKeys(child); err != nil {
			return err
		}
	}
	return nil
}

func decodeYAMLQuotedScalar(value string) (string, error) {
	if value == "" || (value[0] != '\'' && value[0] != '"') {
		return value, nil
	}
	return decodeYAMLListScalar(value)
}

func isYAMLListQuoteStart(value string, start, index int) bool {
	prefix := strings.TrimSpace(value[start:index])
	if prefix == "" {
		return true
	}

	switch prefix[len(prefix)-1] {
	case ':', '[', '{', ',', '-':
		return true
	default:
		return false
	}
}

func isYAMLListCollectionStart(value string, start, index int) bool {
	prefix := strings.TrimSpace(value[start:index])
	if prefix == "" {
		return true
	}

	switch prefix[len(prefix)-1] {
	case ':', '[', '{', ',':
		return true
	default:
		return false
	}
}

func splitYAMLListValue(value, separator string) ([]string, error) {
	if separator == "" {
		return nil, errors.New("list separator cannot be empty")
	}

	values := make([]string, 0, strings.Count(value, separator)+1)
	start := 0
	var quote byte
	var sequenceDepth, mappingDepth int

	for index := 0; index < len(value); {
		character := value[index]

		if quote != 0 {
			switch quote {
			case '\'':
				if character == '\'' {
					if index+1 < len(value) && value[index+1] == '\'' {
						index += 2
						continue
					}
					quote = 0
				}
			case '"':
				if character == '\\' && index+1 < len(value) {
					index += 2
					continue
				}
				if character == '"' {
					quote = 0
				}
			}

			index++
			continue
		}

		if (character == '\'' || character == '"') && isYAMLListQuoteStart(value, start, index) {
			quote = character
			index++
			continue
		}

		if sequenceDepth == 0 && mappingDepth == 0 && strings.HasPrefix(value[index:], separator) {
			values = append(values, strings.TrimSpace(value[start:index]))
			index += len(separator)
			start = index
			continue
		}

		switch character {
		case '[':
			if isYAMLListCollectionStart(value, start, index) {
				sequenceDepth++
			}
		case ']':
			if sequenceDepth > 0 {
				sequenceDepth--
			}
		case '{':
			if isYAMLListCollectionStart(value, start, index) {
				mappingDepth++
			}
		case '}':
			if mappingDepth > 0 {
				mappingDepth--
			}
		}

		index++
	}

	if quote != 0 {
		return nil, fmt.Errorf("unterminated YAML %c-quoted list value", quote)
	}
	if sequenceDepth > 0 || mappingDepth > 0 {
		return nil, errors.New("unterminated YAML collection in list value")
	}

	return append(values, strings.TrimSpace(value[start:])), nil
}

func parseYAMLScalarListValue(value, separator string) ([]string, error) {
	fieldValue, err := parseYAMLValue(value)
	if err == nil && fieldValue.Kind == yaml.SequenceNode {
		values := make([]string, 0, len(fieldValue.Content))
		for _, item := range fieldValue.Content {
			if item.Kind != yaml.ScalarNode {
				return nil, errors.New("field value is not a scalar node")
			}
			values = append(values, item.Value)
		}
		return values, nil
	}

	items, err := splitYAMLListValue(value, separator)
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			values = append(values, "")
			continue
		}
		fieldValue, err := parseYAMLValue(item)
		if err != nil {
			return nil, err
		}
		if fieldValue.Kind != yaml.ScalarNode {
			return nil, errors.New("field value is not a scalar node")
		}
		values = append(values, fieldValue.Value)
	}

	return values, nil
}

func decodeYAMLListScalar(value string) (string, error) {
	var decoded string
	if err := yaml.Unmarshal([]byte(value), &decoded); err != nil {
		return "", err
	}
	return decoded, nil
}

func encodeYAMLListScalar(value string) (string, error) {
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
		Style: yaml.DoubleQuotedStyle,
	}

	encoded, err := yaml.Marshal(node)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(encoded)), nil
}

type Extensions struct {
	IsColumns, IsTable, IsEnum,
	Separator, FieldAlias, Scope, IndexType, IndexFields,
	HashUniqueIndexTag, SortedUniqueIndexTag, HashIndexTag, SortedIndexTag,
	EnumValueAlias protoreflect.ExtensionType
}

func parseExtensions(pbTypes *protoregistry.Types) (*Extensions, error) {
	extensions := &Extensions{}
	var err error

	extName := protoreflect.FullName(fmt.Sprintf("%s.IsColumns", viper.GetString("pb_package")))
	extensions.IsColumns, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IsTable", viper.GetString("pb_package")))
	extensions.IsTable, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IsEnum", viper.GetString("pb_package")))
	extensions.IsEnum, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.Separator", viper.GetString("pb_package")))
	extensions.Separator, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.FieldAlias", viper.GetString("pb_package")))
	extensions.FieldAlias, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.Scope", viper.GetString("pb_package")))
	extensions.Scope, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexType", viper.GetString("pb_package")))
	extensions.IndexType, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexFields", viper.GetString("pb_package")))
	extensions.IndexFields, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.HashUniqueIndexTag", viper.GetString("pb_package")))
	extensions.HashUniqueIndexTag, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.SortedUniqueIndexTag", viper.GetString("pb_package")))
	extensions.SortedUniqueIndexTag, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.HashIndexTag", viper.GetString("pb_package")))
	extensions.HashIndexTag, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.SortedIndexTag", viper.GetString("pb_package")))
	extensions.SortedIndexTag, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.EnumValueAlias", viper.GetString("pb_package")))
	extensions.EnumValueAlias, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	return extensions, nil
}

func matchTargets(field protoreflect.FieldDescriptor, extensions *Extensions) bool {
	targets := viper.GetStringSlice("targets")
	if len(targets) <= 0 {
		return true
	}

	if field.Options().ProtoReflect().Get(extensions.HashUniqueIndexTag.TypeDescriptor()).List().Len() > 0 {
		return true
	}

	if field.Options().ProtoReflect().Get(extensions.SortedUniqueIndexTag.TypeDescriptor()).List().Len() > 0 {
		return true
	}

	if field.Options().ProtoReflect().Get(extensions.HashIndexTag.TypeDescriptor()).List().Len() > 0 {
		return true
	}

	if field.Options().ProtoReflect().Get(extensions.SortedIndexTag.TypeDescriptor()).List().Len() > 0 {
		return true
	}

	scope := field.Options().ProtoReflect().Get(extensions.Scope.TypeDescriptor()).List()
	if scope.Len() <= 0 {
		return true
	}

	return pie.Of(targets).Map(func(target string) string {
		return strings.TrimSpace(target)
	}).Filter(func(target string) bool {
		return target != ""
	}).Any(func(target string) bool {
		for i := 0; i < scope.Len(); i++ {
			if strings.EqualFold(scope.Get(i).String(), target) {
				return true
			}
		}
		return false
	})
}
