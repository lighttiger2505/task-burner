package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli"
)

const (
	ExitCodeOK    int = iota //0
	ExitCodeError int = iota //1
)

func main() {
	err := newApp().Run(os.Args)
	var exitCode = ExitCodeOK
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		exitCode = ExitCodeError
	}
	os.Exit(exitCode)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "task-burner"
	app.HelpName = "tabn"
	app.Usage = "Let's editing the burner list."
	app.UsageText = "tabn [command] [--option]"
	app.Version = "0.0.1"
	app.Author = "lighttiger2505"
	app.Email = "lighttiger2505@gmail.com"
	app.Commands = []cli.Command{
		cli.Command{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add the burner list",
			Action:  AddCommand,
			Flags:   []cli.Flag{},
		},
		cli.Command{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List the burner list's",
			Action:  ListCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:     "recurse, R",
					Usage:    "recurse into burner list",
					EnvVar:   "",
					FilePath: "",
					Required: false,
				},
				cli.BoolFlag{
					Name:     "tree, T",
					Usage:    "tree into burner list",
					EnvVar:   "",
					FilePath: "",
					Required: false,
				},
			},
		},
		cli.Command{
			Name:    "edit",
			Aliases: []string{"e"},
			Usage:   "Edit the burner list",
			Action:  EditCommand,
			Flags:   []cli.Flag{
				// TODO
				// cli.BoolFlag{
				// 	Name:     "editor-options, o",
				// 	Usage:    "lptions for editor to open burner list",
				// 	EnvVar:   "",
				// 	FilePath: "",
				// 	Required: false,
				// },
			},
		},
		cli.Command{
			Name:    "rm",
			Aliases: []string{"d"},
			Usage:   "Remove the burner list",
			Action:  RemoveCommand,
			Flags:   []cli.Flag{},
		},
		cli.Command{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Edit and Fetch a configuration",
			Action:  ConfigCommand,
			Flags:   []cli.Flag{},
		},
	}
	return app
}

func AddCommand(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("the required arguments were not provided: <burner list name>")
	}

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	listName := c.Args()[0]
	listHome := filepath.Join(cfg.HomeDir, listName)

	if isFileExist(listHome) {
		return fmt.Errorf("the duplicated burner list: %s", listName)
	}

	if err := os.MkdirAll(listHome, 0700); err != nil {
		return fmt.Errorf("cannot create directory, %s", err)
	}

	for _, burnnerFileName := range cfg.BurnerNames {
		newPath := filepath.Join(listHome, burnnerFileName)
		if _, err := os.Create(newPath); err != nil {
			if err := os.RemoveAll(listHome); err != nil {
				return fmt.Errorf("cannot remove burner list, %s", err)
			}
			return fmt.Errorf("cannot create burner file, %s", err.Error())
		}
	}

	return nil
}

func ListCommand(c *cli.Context) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	burnerLists, err := ioutil.ReadDir(cfg.HomeDir)
	if err != nil {
		return err
	}

	if c.Bool("recurse") {
		for _, burnerList := range burnerLists {
			fmt.Println(fmt.Sprintf("%s:", burnerList.Name()))

			burnerFiles, err := ioutil.ReadDir(filepath.Join(cfg.HomeDir, burnerList.Name()))
			if err != nil {
				return err
			}
			for _, burnerFile := range burnerFiles {
				fmt.Println(burnerFile.Name())
			}
			fmt.Println("")
		}
		return nil
	}

	if c.Bool("tree") {
		for _, burnerList := range burnerLists {
			fmt.Println(burnerList.Name())

			burnerFiles, err := ioutil.ReadDir(filepath.Join(cfg.HomeDir, burnerList.Name()))
			if err != nil {
				return err
			}
			for i, burnerFile := range burnerFiles {
				if (i + 1) < len(burnerFiles) {
					fmt.Println(fmt.Sprintf(" ├── %s", burnerFile.Name()))
				} else {
					fmt.Println(fmt.Sprintf(" └── %s", burnerFile.Name()))
				}
			}
		}
		return nil
	}

	for _, burnerList := range burnerLists {
		fmt.Println(burnerList.Name())
	}
	return nil
}

func EditCommand(c *cli.Context) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	var listName string
	if len(c.Args()) != 0 {
		listName = c.Args()[0]
	} else {
		var err error
		listName, err = fuzzyfindBurnerList()
		fmt.Println(listName)
		if err != nil {
			return err
		}
		if listName == "" {
			return nil
		}
	}

	listHome := filepath.Join(cfg.HomeDir, listName)
	if !isFileExist(listHome) {
		return fmt.Errorf("not found burner list: %s", listHome)
	}

	burnerFiles, err := ioutil.ReadDir(listHome)
	if err != nil {
		return err
	}

	fileNames := []string{}
	for _, burnerFile := range burnerFiles {
		fileNames = append(fileNames, filepath.Join(listHome, burnerFile.Name()))
	}

	cmdArgs := []string{}
	if len(cfg.EditorOptions) > 0 {
		cmdArgs = append(cmdArgs, cfg.EditorOptions...)
	}
	cmdArgs = append(cmdArgs, fileNames...)

	if err := OpenEditor(cfg.Editor, cmdArgs...); err != nil {
		return fmt.Errorf("failed edit, %s", err)
	}
	return nil
}

func fuzzyfindBurnerList() (string, error) {
	cfg, err := GetConfig()
	if err != nil {
		return "", err
	}

	burnerLists, err := ioutil.ReadDir(cfg.HomeDir)
	if err != nil {
		return "", err
	}

	index, err := fuzzyfinder.Find(
		burnerLists,
		func(i int) string {
			return burnerLists[i].Name()
		},
	)

	if err != nil {
		if err.Error() == fuzzyfinder.ErrAbort.Error() {
			return "", nil
		}
		return "", err
	}
	return burnerLists[index].Name(), nil
}

func RemoveCommand(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("the required arguments were not provided: <burner list name>")
	}

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	listName := c.Args()[0]
	listHome := filepath.Join(cfg.HomeDir, listName)

	if err := os.RemoveAll(listHome); err != nil {
		return fmt.Errorf("cannot remove burner list, %s", err)
	}

	return nil
}

func ConfigCommand(c *cli.Context) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	if c.String("get") != "" {
		switch c.String("get") {
		case "homedir":
			fmt.Println(cfg.HomeDir)
		case "editor":
			fmt.Println(cfg.Editor)
		default:
			return errors.New("Key does not contain a section")
		}
		return nil
	}
	if c.String("get-all") != "" {
	}

	OpenEditor(cfg.Editor, cfg.Path())
	return nil
}

func OpenEditor(program string, args ...string) error {
	cmdargs := strings.Join(args, " ")
	command := program + " " + cmdargs

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
