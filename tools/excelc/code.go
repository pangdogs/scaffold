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

func cmdGenCode(cmd *cobra.Command, args []string) {
	genDependencyCode()

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
		genCode(path, &globalDecls)
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

			genCode(path, &globalDecls)
			return nil
		})
	}
}

func genDependencyCode() {
	if outDir := viper.GetString("pb_out"); outDir != "" {
		genDependencyProtobuf(outDir)
	}
}

func genCode(excelPath string, globalDecls *generic.SliceMap[Type, *Decl]) {
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
