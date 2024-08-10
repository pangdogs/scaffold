package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/scaffold/tools/excelc/excelutils"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func genProtobufMessage(file *excelize.File) proto.Message {
	sheets := slices.DeleteFunc(file.GetSheetList(), func(sheet string) bool {
		return sheet == "" || !unicode.IsLetter(rune(sheet[0]))
	})
	if len(sheets) <= 0 {
		return nil
	}

	pbTypes := protoregistry.GlobalTypes

	extensions, err := parseExtensions(pbTypes)
	if err != nil {
		panic(fmt.Errorf("读取Excel文件 %q 失败，%s", file.Path, err))
	}

	var columnsType, tableType protoreflect.MessageType
	var tableMsg protoreflect.Message
	var tableUniqueIndexes, tableUniqueSortedIndexes generic.UnorderedSliceMap[string, []string]
	tableUniqueSortedIndexesData := map[string]*generic.SliceMap[uint64, uint32]{}

	indexItemTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexItem", viper.GetString("pb_package")))
	indexItemType, err := pbTypes.FindMessageByName(indexItemTypeName)
	if err != nil {
		panic(fmt.Errorf("解析Protobuf类型 %q 失败，%s", indexItemTypeName, err))
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
				panic(fmt.Errorf("读取Excel文件 %q Sheet %q 失败，%s", file.Path, sheet, err))
			}
			defer rows.Close()

			var columns []*Column

			for i := 1; rows.Next(); i++ {
				if i < SheetTableHeader+SheetTableHeaderSize {
					switch i {
					case 1:
						row, err := rows.Columns()
						if err != nil {
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，%s", file.Path, sheet, i, err))
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
						panic(fmt.Errorf("解析Protobuf类型 %q 失败，%s", columnsName, err))
					}
				}

				if tableType == nil {
					tableName := fmt.Sprintf("%s.%s", viper.GetString("pb_package"), snake2Camel(strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path)))+"Table")

					tableType, err = pbTypes.FindMessageByName(protoreflect.FullName(tableName))
					if err != nil {
						panic(fmt.Errorf("解析Protobuf类型 %q 失败，%s", tableName, err))
					}

					indexTypeName := protoreflect.FullName(fmt.Sprintf("%s.IndexType.Enum", viper.GetString("pb_package")))
					indexType, err := pbTypes.FindEnumByName(indexTypeName)
					if err != nil {
						panic(fmt.Errorf("解析Protobuf类型 %q 失败，%s", indexTypeName, err))
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

						switch indexTypeValueDesc.Name() {
						case "UniqueIndex":
							tableUniqueIndexes.Add(string(field.Name()), strings.Split(indexFields, ","))
						case "UniqueSortedIndex":
							tableUniqueSortedIndexes.Add(string(field.Name()), strings.Split(indexFields, ","))
						}
					}

					tableMsg = tableType.New()
				}

				_row, err := rows.Columns()
				if err != nil {
					panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，%s", file.Path, sheet, i, err))
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

					if err := setFieldFromString(rowMsg, field, row.Get(columns[j].Index), extensions); err != nil {
						panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 列 %q 失败，%s", file.Path, sheet, i, field.Name(), err))
					}
				}

				tableRows := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName("Rows"))
				offset := uint32(tableRows.List().Len())
				offsetLines = append(offsetLines, OffsetLine{
					Sheet: sheet,
					Line:  i,
				})
				tableRows.List().Append(protoreflect.ValueOf(rowMsg))

				tableUniqueIndexes.Each(func(indexName string, fields []string) {
					tableIndex := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName)))

					if len(fields) > 1 || excelutils.ProtoMessageFieldNeedHashIndex(columnsType.Descriptor().Fields().ByName(protoreflect.Name(fields[0]))) {
						h := excelutils.NewHash()

						for _, fieldName := range fields {
							fieldValue := rowMsg.Get(rowMsg.Descriptor().Fields().ByName(protoreflect.Name(fieldName)))

							if err := excelutils.AnyToHash(h, fieldValue); err != nil {
								panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，计算索引 %q 值失败，%s", file.Path, sheet, i, indexName, err))
							}
						}

						key := protoreflect.ValueOfUint64(h.Sum64()).MapKey()

						if existed := tableIndex.Map().Get(key); existed.IsValid() {
							conflictedRow := offsetLines[existed.Uint()]
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，索引 %q 值 %d 冲突，与 Sheet %q 行 %d 冲突", file.Path, sheet, i, indexName, h.Sum64(), conflictedRow.Sheet, conflictedRow.Line))
						}

						tableIndex.Map().Set(key, protoreflect.ValueOfUint32(offset))

					} else {
						indexValue, err := excelutils.ProtoMessageFieldToIndex(rowMsg, columnsType.Descriptor().Fields().ByName(protoreflect.Name(fields[0])))
						if err != nil {
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，计算索引 %q 值失败，%s", file.Path, sheet, i, indexName, err))
						}

						key := protoreflect.ValueOfUint64(indexValue).MapKey()

						if existed := tableIndex.Map().Get(key); existed.IsValid() {
							conflictedRow := offsetLines[existed.Uint()]
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，索引 %q 值 %d 冲突，与 Sheet %q 行 %d 冲突", file.Path, sheet, i, indexName, indexValue, conflictedRow.Sheet, conflictedRow.Line))
						}

						tableIndex.Map().Set(protoreflect.ValueOfUint64(indexValue).MapKey(), protoreflect.ValueOfUint32(offset))
					}
				})

				tableUniqueSortedIndexes.Each(func(indexName string, fields []string) {
					indexData, ok := tableUniqueSortedIndexesData[indexName]
					if !ok {
						indexData = &generic.SliceMap[uint64, uint32]{}
						tableUniqueSortedIndexesData[indexName] = indexData
					}

					if len(fields) > 1 || excelutils.ProtoMessageFieldNeedHashIndex(columnsType.Descriptor().Fields().ByName(protoreflect.Name(fields[0]))) {
						h := excelutils.NewHash()

						for _, fieldName := range fields {
							fieldValue := rowMsg.Get(rowMsg.Descriptor().Fields().ByName(protoreflect.Name(fieldName)))

							if err := excelutils.AnyToHash(h, fieldValue); err != nil {
								panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，计算索引 %q 值失败，%s", file.Path, sheet, i, indexName, err))
							}
						}

						if existed, ok := indexData.Get(h.Sum64()); ok {
							conflictedRow := offsetLines[existed]
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，索引 %q 值 %d 冲突，与 Sheet %q 行 %d 冲突", file.Path, sheet, i, indexName, h.Sum64(), conflictedRow.Sheet, conflictedRow.Line))
						}

						indexData.Add(h.Sum64(), offset)

					} else {
						indexValue, err := excelutils.ProtoMessageFieldToIndex(rowMsg, columnsType.Descriptor().Fields().ByName(protoreflect.Name(fields[0])))
						if err != nil {
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，计算索引 %q 值失败，%s", file.Path, sheet, i, indexName, err))
						}

						if existed, ok := indexData.Get(indexValue); ok {
							conflictedRow := offsetLines[existed]
							panic(fmt.Errorf("读取Excel文件 %q Sheet %q 行 %d 失败，索引 %q 值 %d 冲突，与 Sheet %q 行 %d 冲突", file.Path, sheet, i, indexName, indexValue, conflictedRow.Sheet, conflictedRow.Line))
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

	tableUniqueSortedIndexes.Each(func(indexName string, _ []string) {
		tableIndex := tableMsg.Mutable(tableMsg.Descriptor().Fields().ByName(protoreflect.Name(indexName)))

		indexData, ok := tableUniqueSortedIndexesData[indexName]
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

func setFieldFromString(msg protoreflect.Message, field protoreflect.FieldDescriptor, value string, extensions *Extensions) error {
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
		if err != nil {
			return err
		}

		if len(fieldValue.Content) <= 0 {
			return nil
		}
		fieldValue = fieldValue.Content[0]

		return setFieldStructValue(msg, field, fieldValue, extensions)
	}

	return nil
}

func setFieldStructValue(msg protoreflect.Message, field protoreflect.FieldDescriptor, fieldValue *yaml.Node, extensions *Extensions) error {
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

	return protoreflect.Value{}, fmt.Errorf("不支持的枚举值 %q", value)
}

func parseStructValue(value string) (*yaml.Node, error) {
	if !strings.HasPrefix(value, "{") {
		value = "{\n" + value
	}
	if !strings.HasSuffix(value, "}") {
		value += "\n}"
	}

	node := &yaml.Node{}

	err := yaml.Unmarshal([]byte(value), node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type Extensions struct {
	UniqueIndex, UniqueSortedIndex, IndexTyp, IndexFields,
	Separator, FieldAlias, EnumValueAlias protoreflect.ExtensionType
}

func parseExtensions(pbTypes *protoregistry.Types) (*Extensions, error) {
	extensions := &Extensions{}
	var err error

	extName := protoreflect.FullName(fmt.Sprintf("%s.UniqueIndex", viper.GetString("pb_package")))
	extensions.UniqueIndex, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.UniqueSortedIndex", viper.GetString("pb_package")))
	extensions.UniqueSortedIndex, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexTyp", viper.GetString("pb_package")))
	extensions.IndexTyp, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.IndexFields", viper.GetString("pb_package")))
	extensions.IndexFields, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.Separator", viper.GetString("pb_package")))
	extensions.Separator, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.FieldAlias", viper.GetString("pb_package")))
	extensions.FieldAlias, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	extName = protoreflect.FullName(fmt.Sprintf("%s.EnumValueAlias", viper.GetString("pb_package")))
	extensions.EnumValueAlias, err = pbTypes.FindExtensionByName(extName)
	if err != nil {
		return nil, fmt.Errorf("查找Protobuf Option %q 失败，%s", extName, err)
	}

	return extensions, nil
}
