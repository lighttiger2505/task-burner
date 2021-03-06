package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

var configFilePath = filepath.Join(getXDGConfigPath(runtime.GOOS), "config.yml")

type Config struct {
	Version       float32  `yaml:"version"`
	HomeDir       string   `yaml:"homedir"`
	Editor        string   `yaml:"editor"`
	EditorOptions []string `yaml:"editoroptions"`
	BurnerNames   []string `yaml:"burnernames"`
}

var loadedConfig *Config

func loadConfig() (*Config, error) {
	cfg := newConfig()
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetConfig() (*Config, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}
	return loadConfig()
}

func (c *Config) Path() string {
	return configFilePath
}

func (c *Config) Read() (string, error) {
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return "", fmt.Errorf("cannot create directory, %s", err)
	}

	if !IsFileExists(configFilePath) {
		_, err := os.Create(configFilePath)
		if err != nil {
			return "", fmt.Errorf("cannot create config, %s", err.Error())
		}
	}

	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("cannot open config, %s", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("cannot read config, %s", err)
	}

	return string(b), nil
}

func (c *Config) Load() error {
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("cannot create directory, %s", err)
	}

	if !IsFileExists(configFilePath) {
		if err := createNewConfig(); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("cannot open config, %s", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read config, %s", err)
	}

	if err = yaml.Unmarshal(b, c); err != nil {
		return fmt.Errorf("failed unmarshal yaml. \nError: %s \nBuffer: %s", err, string(b))
	}
	return nil
}

func (c *Config) Save() error {
	file, err := os.OpenFile(configFilePath, os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file, %s", err)
	}
	defer file.Close()

	out, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("Failed marshal config. Error: %v", err)
	}

	if _, err = io.WriteString(file, string(out)); err != nil {
		return fmt.Errorf("Failed write config file. Error: %s", err)
	}
	return nil
}

func newConfig() *Config {
	cfg := &Config{}
	return cfg
}

const (
	Version                = 1.0
	FileNameFrontBurner    = "1_front-burner.md"
	FileNameBackBurner     = "2_back-burner.md"
	FileNameKitchenSink    = "3_kitchen-sink.md"
	VimOptionOpenWindow    = "-o"
	VimOptionCommand       = "-c"
	VimOptionCommandLayout = "\"wincmd H\""
)

func createNewConfig() error {
	// Create new config file
	_, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("cannot create config, %s", err.Error())
	}

	// Add default settings
	cfg := newConfig()

	cfg.Version = Version

	configPath := getXDGConfigPath(runtime.GOOS)
	diaryDirPath := filepath.Join(configPath, "_post")
	cfg.HomeDir = diaryDirPath

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cfg.Editor = editor
	cfg.EditorOptions = []string{
		VimOptionOpenWindow,
		VimOptionCommand,
		VimOptionCommandLayout,
	}

	cfg.BurnerNames = []string{
		FileNameFrontBurner,
		FileNameBackBurner,
		FileNameKitchenSink,
	}

	cfg.Save()
	return nil
}

func IsFileExists(fPath string) bool {
	_, err := os.Stat(fPath)
	return err == nil || !os.IsNotExist(err)
}

const APP_NAME = "task-burner"

func getXDGConfigPath(goos string) string {
	var dir string
	if goos == "windows" {
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", APP_NAME)
		}
		dir = filepath.Join(dir, "lab")
	} else {
		dir = filepath.Join(os.Getenv("HOME"), ".config", APP_NAME)
	}
	return dir
}
