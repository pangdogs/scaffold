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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"path/filepath"
	"unicode"
)

func cmdGenCode(cmd *cobra.Command, args []string) {
	loadDependencyProtobuf()

	filepath.Walk(viper.GetString("pb_dir"), func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)

		if fileName == "" || filepath.Ext(fileName) != ".protoset" || !unicode.IsLetter(rune(fileName[0])) {
			return nil
		}

		loadProtobuf(path)
		return nil
	})

	goCodeDir := viper.GetString("go_out")
	if goCodeDir != "" {
		genGoCode(goCodeDir)
	}
}
