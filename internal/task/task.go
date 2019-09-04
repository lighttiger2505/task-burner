package task

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lighttiger2505/task-burner/internal/config"
)

func GetBurnerLists() ([]os.FileInfo, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	burnerLists, err := ioutil.ReadDir(cfg.HomeDir)
	if err != nil {
		return nil, err
	}
	return burnerLists, nil
}

func GetTaskFiles(name string) ([]os.FileInfo, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	taskFiles, err := ioutil.ReadDir(filepath.Join(cfg.HomeDir, name))
	if err != nil {
		return nil, err
	}
	return taskFiles, nil
}
