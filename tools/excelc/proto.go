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
	"git.golaxy.org/core/utils/generic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"io/fs"
	"log"
	"path/filepath"
	"unicode"
)

func cmdGenProto(cmd *cobra.Command, args []string) {
	genDependencyProto()

	var globalDecls generic.SliceMap[Type, *Decl]
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
		genProto(path, &globalDecls)
	}

	excelDir := viper.GetString("excel_dir")
	if excelDir != "" {
		filepath.Walk(excelDir, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() || skip(path) {
				return nil
			}

			fileName := filepath.Base(path)

			if fileName == "" || filepath.Ext(fileName) != ".xlsx" || !unicode.IsLetter(rune(fileName[0])) {
				return nil
			}

			genProto(path, &globalDecls)
			return nil
		})
	}
}

func genDependencyProto() {
	if outDir := viper.GetString("pb_out"); outDir != "" {
		genDependencyProtobuf(outDir)
	}
}

func genProto(excelPath string, globalDecls *generic.SliceMap[Type, *Decl]) {
	excelFile, err := excelize.OpenFile(excelPath)
	if err != nil {
		panic(fmt.Errorf("打开Excel文件 %q 失败，%s", excelPath, err))
	}
	defer excelFile.Close()

	if outDir := viper.GetString("pb_out"); outDir != "" {
		genProtobuf(excelFile, globalDecls, outDir)
		log.Printf("生成Excel文件 %q 结构Protobuf文件成功。", excelPath)
	}
}
