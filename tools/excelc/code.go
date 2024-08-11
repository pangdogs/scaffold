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
		if info.IsDir() {
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
