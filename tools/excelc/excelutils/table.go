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
