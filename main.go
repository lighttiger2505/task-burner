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
	app.Name = "liary"
	app.HelpName = "liary"
	app.Usage = "liary is fastest cli tool for create a diary."
	app.UsageText = "liary [options] [write content for diary]"
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
			Flags:   []cli.Flag{},
		},
		cli.Command{
			Name:    "edit",
			Aliases: []string{"e"},
			Usage:   "Edit the burner list",
			Action:  EditCommand,
			Flags:   []cli.Flag{},
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

	bunnerFileNames := []string{
		"1_front-burner.md",
		"2_back-burner.md",
		"3_kitchen-sink.md",
	}
	for _, burnnerFileName := range bunnerFileNames {
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
		panic(err)
	}

	for _, burnerList := range burnerLists {
		fmt.Println(burnerList)
	}
	return nil
}

func EditCommand(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("the required arguments were not provided: <burner list name>")
	}

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	listName := c.Args()[0]
	listHome := filepath.Join(cfg.HomeDir, listName)

	if err := OpenEditor(cfg.Editor, listHome); err != nil {
		return fmt.Errorf("failed edit, %s", err)
	}
	return nil
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
