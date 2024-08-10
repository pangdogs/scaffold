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
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
)

func genBinaryData(tableMsg proto.Message, outDir string) (string, error) {
	tableData, err := proto.Marshal(tableMsg)
	if err != nil {
		return "", err
	}

	outFile := filepath.Join(outDir, strings.TrimSuffix(string(tableMsg.ProtoReflect().Descriptor().Name()), "Table")+".bin")

	os.MkdirAll(outDir, os.ModePerm)
	err = os.WriteFile(outFile, tableData, os.ModePerm)
	if err != nil {
		return "", err
	}

	return outFile, nil
}
