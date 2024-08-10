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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"unicode"
)

func main() {
	cmd := &cobra.Command{
		Short: "Excel表格处理工具，生成表格结构代码和导出数据文件。",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())

			{
				excelFilePaths := viper.GetStringSlice("excel_files")

				for _, path := range excelFilePaths {
					if filepath.Ext(path) != ".xlsx" {
						panic(fmt.Errorf("[--excel_files]文件 %q 错误，文件名后缀名必须为.xlsx", path))
					}
					if !unicode.IsLetter(rune(filepath.Base(path)[0])) {
						panic(fmt.Errorf("[--excel_files]文件 %q 错误，文件名首字符必须为字母", path))
					}
					stat, err := os.Stat(path)
					if err != nil {
						panic(fmt.Errorf("[--excel_files]文件 %q 错误，%s", path, err))
					}
					if stat.IsDir() {
						panic(fmt.Errorf("[--excel_files]文件 %q 错误，不能为文件夹", path))
					}
				}
			}

			{
				excelDir := viper.GetString("excel_dir")
				if excelDir != "" {
					stat, err := os.Stat(excelDir)
					if err != nil {
						panic(fmt.Errorf("[--excel_dir]文件夹错误，%s", err))
					}
					if !stat.IsDir() {
						panic("[--excel_dir]文件夹错误，必须为文件夹")
					}
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
	}
	cmd.PersistentFlags().StringSlice("excel_files", nil, "指定输入Excel文件列表（优先）。")
	cmd.PersistentFlags().String("excel_dir", "", "指定输入Excel文件目录。")

	codeCmd := &cobra.Command{
		Use:   "code",
		Short: "生成Excel结构代码。",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())

			{
				pkg := viper.GetString("pb_package")
				if pkg == "" {
					panic("[--pb_package]值不能为空")
				}
			}
		},
		Run: cmdGenCode,
	}
	codeCmd.Flags().String("pb_out", "", "指定输出Protobuf文件目录。")
	codeCmd.Flags().String("pb_package", "excel", "指定输出Protobuf文件包名。")
	codeCmd.Flags().StringSlice("pb_imports", nil, "指定输出Protobuf文件导入项。")
	codeCmd.Flags().Int("pb_custom_options", 10000, "Protobuf自定义选项编号。")
	codeCmd.Flags().StringToString("pb_options", map[string]string{"go_package": "./excel"}, "指定输出Protobuf文件选项。")

	dataCmd := &cobra.Command{
		Use:   "data",
		Short: "导出Excel数据。",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())

			{
				pbDir := viper.GetString("pb_dir")
				if pbDir == "" {
					panic("[--pb_dir]值不能为空")
				}
				stat, err := os.Stat(pbDir)
				if err != nil {
					panic(fmt.Errorf("[--pb_dir]文件夹错误，%s", err))
				}
				if !stat.IsDir() {
					panic("[--pb_dir]文件夹错误，必须为文件夹")
				}
			}

			{
				pkg := viper.GetString("pb_package")
				if pkg == "" {
					panic("[--pb_package]值不能为空")
				}
			}
		},
		Run: cmdGenData,
	}
	dataCmd.Flags().String("pb_dir", "", "指定Excel编译输出Protobuf文件的目录。")
	dataCmd.Flags().String("pb_package", "excel", "指定Excel编译输出Protobuf文件的包名。")
	dataCmd.Flags().Bool("gzip", false, "是否使用Gzip压缩数据。")
	dataCmd.Flags().String("binary_out", "", "输出Binary数据目录。")
	dataCmd.Flags().String("json_out", "", "输出Json数据目录。")
	dataCmd.Flags().Bool("json_multiline", false, "Json数据是否多行。")
	dataCmd.Flags().String("json_indent", "", "Json数据缩进。")

	cmd.AddCommand(codeCmd, dataCmd)

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
