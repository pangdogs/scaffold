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
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func genBinaryData(tableMsg proto.Message, outDir string) (string, error) {
	tableData, err := proto.MarshalOptions{Deterministic: true}.Marshal(tableMsg)
	if err != nil {
		return "", err
	}

	outFile := filepath.Join(outDir, string(tableMsg.ProtoReflect().Descriptor().Name())+".bin")

	os.MkdirAll(outDir, os.ModePerm)
	err = os.WriteFile(outFile, tableData, os.ModePerm)
	if err != nil {
		return "", err
	}

	return outFile, nil
}

func genChunkedBinaryData(tableMsg proto.Message, outDir string) (string, int, error) {
	table := tableMsg.ProtoReflect()
	fields := table.Descriptor().Fields()

	rowsField := fields.ByName("Rows")
	if rowsField == nil {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q not found", table.Descriptor().FullName(), "Rows")
	}
	if !rowsField.IsList() || rowsField.Kind() != protoreflect.MessageKind {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q must be a repeated message", table.Descriptor().FullName(), "Rows")
	}

	chunkSizeField := fields.ByName("ChunkSize")
	if chunkSizeField == nil {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q not found", table.Descriptor().FullName(), "ChunkSize")
	}
	if chunkSizeField.Kind() != protoreflect.Uint32Kind {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q must be uint32", table.Descriptor().FullName(), "ChunkSize")
	}

	chunksField := fields.ByName("Chunks")
	if chunksField == nil {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q not found", table.Descriptor().FullName(), "Chunks")
	}
	if !chunksField.IsList() || chunksField.Kind() != protoreflect.MessageKind {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q must be a repeated message", table.Descriptor().FullName(), "Chunks")
	}

	chunkFields := chunksField.Message().Fields()
	chunkOffsetField := chunkFields.ByName("Offset")
	if chunkOffsetField == nil {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q not found", chunksField.Message().FullName(), "Offset")
	}

	chunkCountField := chunkFields.ByName("Count")
	if chunkCountField == nil {
		return "", 0, fmt.Errorf("parse proto type %q failed: field %q not found", chunksField.Message().FullName(), "Count")
	}

	chunkSize := viper.GetUint32("binary_chunk_size")
	if chunkSize <= 0 {
		return "", 0, fmt.Errorf("[--binary_chunk_size] value must be greater than 0")
	}

	baseFile := filepath.Join(outDir, string(table.Descriptor().Name())+".bin")
	idxFile := baseFile + ".idx"

	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return "", 0, err
	}

	rows := table.Get(rowsField).List()
	tableIdxMsg := table.New()
	table.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if field.Number() != rowsField.Number() {
			tableIdxMsg.Set(field, value)
		}
		return true
	})
	tableIdxMsg.Set(chunkSizeField, protoreflect.ValueOfUint32(chunkSize))
	tableIdxMsg.Clear(chunksField)
	chunks := tableIdxMsg.Mutable(chunksField).List()

	chunksNum := 0

	for offset := 0; offset < rows.Len(); offset += int(chunkSize) {
		count := int(chunkSize)
		if remain := rows.Len() - offset; remain < count {
			count = remain
		}

		chunkMeta := chunks.NewElement().Message()
		chunkMeta.Set(chunkOffsetField, protoreflect.ValueOfUint32(uint32(offset)))
		chunkMeta.Set(chunkCountField, protoreflect.ValueOfUint32(uint32(count)))
		chunks.Append(protoreflect.ValueOfMessage(chunkMeta))

		tableChunkMsg := table.New()
		chunkRows := tableChunkMsg.Mutable(rowsField).List()

		for i := 0; i < count; i++ {
			chunkRows.Append(rows.Get(offset + i))
		}

		chunkData, err := proto.MarshalOptions{Deterministic: true}.Marshal(tableChunkMsg.Interface())
		if err != nil {
			return "", 0, err
		}

		chunkFile := fmt.Sprintf("%s.chk_%d", baseFile, chunksNum)
		if err := os.WriteFile(chunkFile, chunkData, os.ModePerm); err != nil {
			return "", 0, err
		}

		chunksNum++
	}

	tableIdxData, err := proto.MarshalOptions{Deterministic: true}.Marshal(tableIdxMsg.Interface())
	if err != nil {
		return "", 0, err
	}

	if err := os.WriteFile(idxFile, tableIdxData, os.ModePerm); err != nil {
		return "", 0, err
	}

	return idxFile, chunksNum, nil
}
