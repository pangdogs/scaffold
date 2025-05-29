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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func cmdGenData(cmd *cobra.Command, args []string) {
	loadDependencyProtobuf()

	skipped := map[string]struct{}{}

	skip := func(p string) bool {
		p, _ = filepath.Abs(p)
		_, ok := skipped[p]
		if !ok {
			skipped[p] = struct{}{}
		}
		return ok
	}

	for _, path := range viper.GetStringSlice("excel_files") {
		if skip(path) {
			continue
		}
		genData(path)
	}

	excelDir := viper.GetString("excel_dir")
	if excelDir != "" {
		filepath.Walk(excelDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() || skip(path) {
				return nil
			}

			fileName := filepath.Base(path)

			if fileName == "" || filepath.Ext(fileName) != ".xlsx" || !unicode.IsLetter(rune(fileName[0])) {
				return nil
			}

			genData(path)
			return nil
		})
	}
}

func loadDependencyProtobuf() {
	pbPath := filepath.Join(viper.GetString("pb_dir"), fmt.Sprintf("%s.protoset", DependencyProtobuf))

	pbData, err := os.ReadFile(pbPath)
	if err != nil {
		panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
	}

	pbSet := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(pbData, pbSet)
	if err != nil {
		panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
	}

	pbFiles := protoregistry.GlobalFiles
	pbTypes := protoregistry.GlobalTypes

	for _, fdProto := range pbSet.File {
		pbFile, err := protodesc.NewFile(fdProto, pbFiles)
		if err != nil {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		_, err = pbFiles.FindFileByPath(pbFile.Path())
		if err == nil {
			continue
		}
		if !errors.Is(err, protoregistry.NotFound) {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		err = pbFiles.RegisterFile(pbFile)
		if err != nil {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		regProtobufTypes(pbTypes, pbFile)
	}
}

func loadProtobuf(pbPath string) {
	pbData, err := os.ReadFile(pbPath)
	if err != nil {
		panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
	}

	pbSet := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(pbData, pbSet)
	if err != nil {
		panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
	}

	pbFiles := protoregistry.GlobalFiles
	pbTypes := protoregistry.GlobalTypes

	for _, fdProto := range pbSet.File {
		pbFile, err := protodesc.NewFile(fdProto, pbFiles)
		if err != nil {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		_, err = pbFiles.FindFileByPath(pbFile.Path())
		if err == nil {
			continue
		}
		if !errors.Is(err, protoregistry.NotFound) {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		err = pbFiles.RegisterFile(pbFile)
		if err != nil {
			panic(fmt.Errorf("读取Protobuf文件 %q 失败，%s", pbPath, err))
		}

		regProtobufTypes(pbTypes, pbFile)
	}
}

func genData(excelPath string) {
	excelFile, err := excelize.OpenFile(excelPath)
	if err != nil {
		panic(fmt.Errorf("打开Excel文件 %q 失败，%s", excelPath, err))
	}
	defer excelFile.Close()

	loadProtobuf(filepath.Join(viper.GetString("pb_dir"), snake2Camel(strings.TrimSuffix(filepath.Base(excelPath), filepath.Ext(excelPath)))+".protoset"))

	tableMsg := genProtobufMessage(excelFile)
	if tableMsg == nil {
		log.Printf("导出Excel文件 %q 没有数据。", excelPath)
		return
	}

	if outDir := viper.GetString("binary_out"); outDir != "" {
		outFile, err := genBinaryData(tableMsg, outDir)
		if err != nil {
			panic(fmt.Errorf("导出Excel文件 %q Binary数据文件失败，%s", excelPath, err))
		}
		log.Printf("导出Excel文件 %q Binary数据文件 %q 成功。", excelPath, outFile)
	}

	if outDir := viper.GetString("json_out"); outDir != "" {
		outFile, err := genJsonData(tableMsg, outDir, viper.GetBool("json_multiline"), viper.GetString("json_indent"))
		if err != nil {
			panic(fmt.Errorf("导出Excel文件 %q Json数据文件失败，%s", excelPath, err))
		}
		log.Printf("导出Excel文件 %q Json数据文件 %q 成功。", excelPath, outFile)
	}
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
