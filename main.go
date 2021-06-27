package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/chitoku-k/kubernetes-field-selector-extractor/domain"
	"github.com/chitoku-k/kubernetes-field-selector-extractor/service"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	args := os.Args[1:]
	if len(args) == 0 {
		logrus.Error("Directory name must be specified")
		os.Exit(1)
	}

	var selectors []domain.FieldSelector
	for _, dir := range args {
		s, err := os.Stat(dir)
		if os.IsNotExist(err) {
			logrus.Errorf("No such directory: %q", dir)
			os.Exit(1)
		}
		if !s.IsDir() {
			logrus.Errorf("Not a directory: %q", dir)
			os.Exit(1)
		}

		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}
			logrus.Debugf("Directory: %v", path)

			finder := service.NewFinderService(path)
			s, err := finder.Do()
			if err != nil {
				logrus.Errorf("Failed to find: %v", err)
			}
			selectors = append(selectors, s...)

			return nil
		})
	}

	json.NewEncoder(os.Stdout).Encode(selectors)
}
