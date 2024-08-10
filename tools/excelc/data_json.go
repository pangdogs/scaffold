package main

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
)

func genJsonData(tableMsg proto.Message, outDir string, multiline bool, indent string) (string, error) {
	tableData, err := protojson.MarshalOptions{
		Multiline: multiline,
		Indent:    indent,
	}.Marshal(tableMsg)

	outFile := filepath.Join(outDir, strings.TrimSuffix(string(tableMsg.ProtoReflect().Descriptor().Name()), "Table")+".json")

	os.MkdirAll(outDir, os.ModePerm)
	err = os.WriteFile(outFile, tableData, os.ModePerm)
	if err != nil {
		return "", err
	}

	return outFile, nil
}
