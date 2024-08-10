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
