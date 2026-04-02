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
	tableSortedUniqueIndexesData := map[string]*generic.SliceMap[uint64, uint32]{}

	indexItemTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexItem", viper.GetString("pb_package")))
	indexItemType, err := pbTypes.FindMessageByName(indexItemTypeName)
	if err != nil {
		log.Panicf("parse proto type %q failed, %s", indexItemTypeName, err)
	}

	indexConflictTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexConflict", viper.GetString("pb_package")))
	indexConflictType, err := pbTypes.FindMessageByName(indexConflictTypeName)
	if err != nil {
		log.Panicf("parse proto type %q failed, %s", indexConflictTypeName, err)
	}

	type Column struct {
		Name  string
		Index int
	}

	type OffsetLine struct {
		Sheet string
		Line  int
	}

	var offsetLines []OffsetLine

	for _, sheet := range sheets {
		func() {
			rows, err := file.Rows(sheet)
			if err != nil {
				log.Panicf("read excel file %q sheet %q failed, %s", file.Path, sheet, err)
			}
			defer rows.Close()

			var columns []*Column

			for i := 1; rows.Next(); i++ {
				if i < SheetTableHeader+SheetTableHeaderSize {
					switch i {
					case 1:
						row, err := rows.Columns()
						if err != nil {
							log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, sheet, i, err)
						}

						for i, cell := range row {
							columns = append(columns, &Column{
								Name:  snake2Camel(cell),
								Index: i,
							})
						}

						for _, col := range columns {
							col.Name = strings.NewReplacer("\r", "", "\n", "\\n").Replace(strings.TrimSpace(col.Name))
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

					indexTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexType.Enum", viper.GetString("pb_package")))
					indexType, err := pbTypes.FindEnumByName(indexTypeName)
					if err != nil {
						log.Panicf("parse proto type %q failed, %s", indexTypeName, err)
					}

					for j := range tableType.Descriptor().Fields().Len() {
						field := tableType.Descriptor().Fields().Get(j)

						indexTypeValue, ok := proto.GetExtension(field.Options(), extensions.IndexTyp).(protoreflect.EnumNumber)
						if !ok || indexTypeValue <= 0 {
							continue
						}

						indexFields := proto.GetExtension(field.Options(), extensions.IndexFields).(string)
						if indexFields == "" {
							continue
						}

						indexTypeValueDesc := indexType.Descriptor().Values().ByNumber(indexTypeValue)
						if indexTypeValueDesc == nil {
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

						switch indexTypeValueDesc.Name() {
						case "HashUniqueIndex":
							tableHashUniqueIndexes.Add(string(field.Name()), fieldDescs)
						case "SortedUniqueIndex":
							tableSortedUniqueIndexes.Add(string(field.Name()), fieldDescs)
						}
					}

					tableMsg = tableType.New()
				}

				_row, err := rows.Columns()
				if err != nil {
					log.Panicf("read excel file %q sheet %q row %d failed, %s", file.Path, sheet, i, err)
				}
				if len(_row) > len(columns) {
					_row = _row[:len(columns)]
				}
				row := Row(_row)

				if row.Empty() {
					continue
				}

				rowMsg := columnsType.New()

				for j := 0; j < rowMsg.Descriptor().Fields().Len(); j++ {
					field := rowMsg.Descriptor().Fields().Get(j)
					columnIdx := int(field.Number()) - 1
					if columnIdx < 0 || columnIdx >= len(columns) {
						log.Panicf("read excel file %q sheet %q row %d column %q failed: field number %d is out of range", file.Path, sheet, i, field.Name(), field.Number())
					}

					if err := setFieldFromString(rowMsg, field, row.Get(columns[columnIdx].Index), extensions); err != nil {
						log.Panicf("read excel file %q sheet %q row %d column %q failed, %s", file.Path, sheet, i, field.Name(), err)
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

							log.Printf("read excel file %q sheet %q row %d warning: index %q value %d collides with sheet %q row %d; stored in conflict bucket", file.Path, sheet, i, indexName, h.Sum64(), offsetLines[existed.Uint()].Sheet, offsetLines[existed.Uint()].Line)
							appendIndexConflictOffset(tableMsg, indexConflictType, indexName, key, offset)
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

							log.Printf("read excel file %q sheet %q row %d warning: index %q value %d collides with sheet %q row %d; stored in conflict bucket", file.Path, sheet, i, indexName, h.Sum64(), offsetLines[existed].Sheet, offsetLines[existed].Line)
							appendIndexConflictOffset(tableMsg, indexConflictType, indexName, key, offset)
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
			}
		}()
	}

	if tableMsg == nil {
		return nil
	}

	tableSortedUniqueIndexes.Each(func(indexName string, _ []protoreflect.FieldDescriptor) {
		tableIndex := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName)))

		indexData, ok := tableSortedUniqueIndexesData[indexName]
		if !ok {
			return
		}

		indexData.Each(func(value uint64, offset uint32) {
			indexItem := indexItemType.New()
			indexItem.Set(indexItem.Descriptor().Fields().ByName("Value"), protoreflect.ValueOfUint64(value))
			indexItem.Set(indexItem.Descriptor().Fields().ByName("Offset"), protoreflect.ValueOfUint32(offset))

			tableIndex.List().Append(protoreflect.ValueOf(indexItem))
		})
	})

	return tableMsg.Interface()
}

func appendIndexConflictOffset(tableMsg protoreflect.Message, indexConflictType protoreflect.MessageType, indexName string, key protoreflect.MapKey, offset uint32) {
	conflictField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName + "Conflict"))
	if conflictField == nil {
		log.Panicf("parse proto type %q failed: conflict field %q not found", tableMsg.Descriptor().FullName(), indexName+"Conflict")
	}

	conflictIndex := tableMsg.Mutable(conflictField)
	bucket := conflictIndex.Map().Mutable(key).Message()
	if !bucket.IsValid() {
		bucket = indexConflictType.New()
		conflictIndex.Map().Set(key, protoreflect.ValueOfMessage(bucket))
	}

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

	conflictField := tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName + "Conflict"))
	if conflictField == nil {
		return 0, false
	}

	conflictBucket := tableMsg.Get(conflictField).Map().Get(key)
	if !conflictBucket.IsValid() {
		return 0, false
	}

	offsetsField := conflictBucket.Message().Descriptor().Fields().ByName("Offsets")
	if offsetsField == nil {
		return 0, false
	}

	offsets := conflictBucket.Message().Get(offsetsField).List()
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

	if value == "" {
		return nil
	}

	switch field.Kind() {
	case protoreflect.BoolKind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				b, err := strconv.ParseBool(c)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfBool(b))
			}

		} else {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfBool(b))
		}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseInt(c, 10, 32)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfInt32(int32(n)))
			}

		} else {
			n, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfInt32(int32(n)))
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseInt(c, 10, 64)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfInt64(n))
			}

		} else {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfInt64(n))
		}

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseUint(c, 10, 32)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfUint32(uint32(n)))
			}

		} else {
			n, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfUint32(uint32(n)))
		}

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseUint(c, 10, 64)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfUint64(n))
			}

		} else {
			n, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfUint64(n))
		}

	case protoreflect.FloatKind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseFloat(c, 32)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfFloat32(float32(n)))
			}

		} else {
			n, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfFloat32(float32(n)))
		}

	case protoreflect.DoubleKind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				n, err := strconv.ParseFloat(c, 64)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfFloat64(n))
			}

		} else {
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfFloat64(n))
		}

	case protoreflect.StringKind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				l.Append(protoreflect.ValueOfString(c))
			}

		} else {
			msg.Set(field, protoreflect.ValueOfString(value))
		}

	case protoreflect.BytesKind:
		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				bs, err := base64.URLEncoding.DecodeString(c)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOfBytes(bs))
			}

		} else {
			bs, err := base64.URLEncoding.DecodeString(value)
			if err != nil {
				return err
			}
			msg.Set(field, protoreflect.ValueOfBytes(bs))
		}

	case protoreflect.EnumKind:
		enumDesc := field.Enum()

		if field.IsList() {
			sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
			l := msg.Mutable(field).List()

			for _, c := range strings.Split(value, sep) {
				enumValue, err := parseEnumValue(enumDesc, c, extensions)
				if err != nil {
					return err
				}
				l.Append(enumValue)
			}

		} else {
			enumValue, err := parseEnumValue(enumDesc, value, extensions)
			if err != nil {
				return err
			}
			msg.Set(field, enumValue)
		}

	case protoreflect.MessageKind:
		fieldValue, err := parseStructValue(value)

		switch {
		case field.IsList():
			if err != nil {
				sep := proto.GetExtension(field.Options(), extensions.Separator).(string)

				for _, v := range strings.Split(value, sep) {
					childValue, err := parseStructValue(v)
					if err != nil {
						return err
					}

					if childValue.Kind != yaml.DocumentNode || len(childValue.Content) <= 0 {
						continue
					}
					childValue = childValue.Content[0]

					if childValue.Kind != yaml.MappingNode {
						return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to an object type", childValue.Value)
					}

					err = setFieldStructValue(msg, field, childValue, extensions)
					if err != nil {
						return err
					}
				}

			} else {
				if fieldValue.Kind != yaml.DocumentNode || len(fieldValue.Content) <= 0 {
					return nil
				}
				fieldValue = fieldValue.Content[0]

				switch fieldValue.Kind {
				case yaml.SequenceNode:
					for _, c := range fieldValue.Content {
						if c.Kind != yaml.MappingNode {
							return fmt.Errorf("YAML config %q is not a MappingNode and cannot be assigned to an object type", c.Value)
						}
						err := setFieldStructValue(msg, field, c, extensions)
						if err != nil {
							return err
						}
					}
				case yaml.MappingNode:
					err := setFieldStructValue(msg, field, fieldValue, extensions)
					if err != nil {
						return err
					}
				}
			}

		case field.IsMap():
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

			return setMappingValue(msg, field, fieldValue, extensions)

		default:
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

	return nil
}

func setFieldStructValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, fieldValue *yaml.Node, extensions *Extensions) error {
	if !matchTargets(field, extensions) {
		return nil
	}

	if field.Kind() != protoreflect.MessageKind {
		if field.IsList() {
			switch fieldValue.Kind {
			case yaml.SequenceNode:
				sep := proto.GetExtension(field.Options(), extensions.Separator).(string)
				sb := strings.Builder{}

				for _, c := range fieldValue.Content {
					if c.Kind != yaml.ScalarNode {
						return errors.New("field value is not a scalar node")
					}

					if sb.Len() > 0 {
						sb.WriteString(sep)
					}
					sb.WriteString(c.Value)
				}

				return setFieldFromString(msg, field, sb.String(), extensions)

			case yaml.ScalarNode:
				return setFieldFromString(msg, field, fieldValue.Value, extensions)

			default:
				return errors.New("field value is not a sequence or scalar node")
			}

		} else {
			if fieldValue.Kind != yaml.ScalarNode {
				return errors.New("field value is not a scalar node")
			}
			return setFieldFromString(msg, field, fieldValue.Value, extensions)
		}
	}

	fieldType, err := protoregistry.GlobalTypes.FindMessageByName(field.Message().FullName())
	if err != nil {
		return err
	}

	if field.IsList() {
		l := msg.Mutable(field).List()

		switch fieldValue.Kind {
		case yaml.SequenceNode:
			for _, c := range fieldValue.Content {
				fieldMsg, err := makeStructValue(fieldType, c, extensions)
				if err != nil {
					return err
				}
				l.Append(protoreflect.ValueOf(fieldMsg))
			}

		case yaml.MappingNode:
			fieldMsg, err := makeStructValue(fieldType, fieldValue, extensions)
			if err != nil {
				return err
			}
			l.Append(protoreflect.ValueOf(fieldMsg))

		default:
			return errors.New("field value is not a sequence or scalar node")
		}

	} else {
		fieldMsg, err := makeStructValue(fieldType, fieldValue, extensions)
		if err != nil {
			return err
		}
		msg.Set(field, protoreflect.ValueOf(fieldMsg))
	}

	return nil
}

func makeStructValue(ty protoreflect.MessageType, value *yaml.Node, extensions *Extensions) (protoreflect.Message, error) {
	if value.Kind != yaml.MappingNode {
		return nil, errors.New("field value is not a mapping node")
	}

	msg := ty.New()

	for i := range msg.Descriptor().Fields().Len() {
		field := msg.Descriptor().Fields().Get(i)

		idx := slices.IndexFunc(value.Content, func(node *yaml.Node) bool {
			return node.Value == string(field.Name())
		})
		if idx < 0 {
			fieldAlias := proto.GetExtension(field.Options(), extensions.FieldAlias).(string)
			if fieldAlias != "" {
				idx = slices.IndexFunc(value.Content, func(node *yaml.Node) bool {
					return node.Value == fieldAlias
				})
			}
			if idx < 0 {
				continue
			}
		}

		err := setFieldStructValue(msg, field, value.Content[idx+1], extensions)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func setMappingValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, value *yaml.Node, extensions *Extensions) error {
	if !field.IsMap() {
		return errors.New("field type not mapping")
	}

	if value.Kind != yaml.MappingNode {
		return errors.New("field value is not a mapping node")
	}

	mapping := msg.Mutable(field).Map()
	kType := field.MapKey()
	vType := field.MapValue()

	for i := 0; i < len(value.Content); i += 2 {
		k := value.Content[i]
		v := value.Content[i+1]

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
			enumValue, err := parseEnumValue(field.Enum(), v.Value, extensions)
			if err != nil {
				return err
			}
			iv = enumValue
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
	if !strings.HasPrefix(value, "[") {
		if !strings.HasPrefix(value, "{") {
			value = "{\n" + value + "\n}"
		}
	}

	node := &yaml.Node{}

	err := yaml.Unmarshal([]byte(value), node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type Extensions struct {
	IsColumns, IsTable,
	Separator, FieldAlias, Scope, IsRows, IndexTyp, IndexFields, HashUniqueIndex, SortedUniqueIndex,
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

	extName = protoreflect.FullName(fmt.Sprintf("%s.IsRows", viper.GetString("pb_package")))
	extensions.IsRows, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexTyp", viper.GetString("pb_package")))
	extensions.IndexTyp, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexFields", viper.GetString("pb_package")))
	extensions.IndexFields, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.HashUniqueIndex", viper.GetString("pb_package")))
	extensions.HashUniqueIndex, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("find proto option %q failed, %s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.SortedUniqueIndex", viper.GetString("pb_package")))
	extensions.SortedUniqueIndex, err = pbTypes.FindExtensionByName(extName)
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

	if field.Options().ProtoReflect().Get(extensions.HashUniqueIndex.TypeDescriptor()).List().Len() > 0 {
		return true
	}

	if field.Options().ProtoReflect().Get(extensions.SortedUniqueIndex.TypeDescriptor()).List().Len() > 0 {
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
