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

package excelutils

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
)

func LoadTableFromBinaryFile(tab proto.Message, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return LoadTableFromBinaryData(tab, data)
}

func LoadTableFromBinaryData(tab proto.Message, data []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(data, tab)
}

func LoadTableFromJsonFile(tab proto.Message, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return LoadTableFromJsonData(tab, data)
}

func LoadTableFromJsonData(tab proto.Message, data []byte) error {
	return protojson.UnmarshalOptions{}.Unmarshal(data, tab)
}
