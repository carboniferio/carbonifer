package data

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

//go:embed data/*
var data embed.FS

// ReadDataFile reads a file from the data directory
func ReadDataFile(filename string) []byte {
	dataPath := viper.GetString("data.path")
	if dataPath != "" {
		// If the environment variable is set, read from the specified file
		filePath := filepath.Join(dataPath, filename)
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			log.Debugf("  reading datafile '%v' from: %v", filename, filePath)
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}
			return data
		}
		return readEmbeddedFile(filename)

	}
	return readEmbeddedFile(filename)
}

func readEmbeddedFile(filename string) []byte {
	log.Debugf("  reading datafile '%v' embedded", filename)
	data, err := fs.ReadFile(data, "data/"+filename)
	if err != nil {
		errW := errors.Wrap(err, "cannot read embedded data file")
		log.Fatal(errW)
	}
	return data
}
