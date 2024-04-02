package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	config, err := getConfig()
	if err != nil {
		red()
		fmt.Println(err)
		reset()
		return
	}

	startTime := time.Now()
	err = redub(config.Path, config.OldName, config.NewName)
	if err != nil {
		red()
		fmt.Println(err)
		reset()
		return
	}
	elapsedTime := time.Since(startTime)
	green()
	fmt.Printf("Redub completed in %s.\n", elapsedTime)
	reset()
}

func redub(path string, oldName string, newName string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfos, err := file.Readdir(-1)
	if err != nil {
		return err
	}

	files := []os.FileInfo{}
	dirs := []os.FileInfo{}

	for _, fileInfo := range fileInfos {
		if shouldProcess(fileInfo) {
			if fileInfo.IsDir() {
				dirs = append(dirs, fileInfo)
			} else {
				files = append(files, fileInfo)
			}
		} else {
			gray()
			fmt.Println("Skipping", fileInfo.Name())
			reset()
		}
	}

	for _, fileInfo := range files {
		err := redubContents(path, fileInfo, oldName, newName)
		if err != nil {
			return err
		}
	}

	for _, fileInfo := range files {
		err := redubName(path, fileInfo, oldName, newName)
		if err != nil {
			return err
		}
	}

	for _, fileInfo := range dirs {
		err := redub(appendPath(path, fileInfo.Name()), oldName, newName)
		if err != nil {
			return err
		}

		err = redubName(path, fileInfo, oldName, newName)
		if err != nil {
			return err
		}
	}

	return nil
}

func redubContents(path string, fileInfo os.FileInfo, oldName, newName string) error {
	abspath := appendPath(path, fileInfo.Name())
	content, err := os.ReadFile(abspath)

	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	occurrences := 0

	for i, line := range lines {
		if strings.Contains(line, oldName) {
			occurrences++
			lines[i] = strings.Replace(line, oldName, newName, -1)
		}
	}

	if occurrences > 0 {
		err = os.WriteFile(abspath, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return err
		}
		gray()
		fmt.Printf("Redubbed %d occurrences in %s\n", occurrences, abspath)
		reset()
	}

	return nil
}

func redubName(path string, fileInfo os.FileInfo, oldName, newName string) error {
	if strings.Contains(fileInfo.Name(), oldName) {
		oldPath := appendPath(path, fileInfo.Name())
		newPath := appendPath(path, strings.Replace(fileInfo.Name(), oldName, newName, -1))

		err := os.Rename(oldPath, newPath)
		if err != nil {
			return err
		}

		gray()
		if fileInfo.IsDir() {
			fmt.Printf("Redubbed directory from %s to %s\n", oldPath, newPath)
		} else {
			fmt.Printf("Redubbed file from %s to %s\n", oldPath, newPath)
		}
		reset()
	}

	return nil
}

func shouldProcess(fileInfo os.FileInfo) bool {
	dir := fileInfo.IsDir()
	name := fileInfo.Name()

	if dir && name == "node_modules" {
		return false
	}

	if dir && strings.HasPrefix(name, ".") {
		return false
	}

	return true
}

func appendPath(path string, name string) string {
	if path == "" {
		return name
	}

	if strings.HasSuffix(path, "/") {
		return path + name
	}

	return path + "/" + name
}

func getFileInfos(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfos, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	filtered := []os.FileInfo{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() && fileInfo.Name() == "node_modules" {
			continue
		}

		if fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), ".") {
			continue
		}

		filtered = append(filtered, fileInfo)
	}

	return filtered, nil
}

func getConfig() (*Config, error) {
	if len(os.Args) > 5 {
		return nil, errors.New("Too many arguments, expected <path?> <current?> <new?> <config?>.")
	}

	ansiRedub := `
██████╗ ███████╗██████╗ ██╗   ██╗██████╗ ██╗
██╔══██╗██╔════╝██╔══██╗██║   ██║██╔══██╗██║
██████╔╝█████╗  ██║  ██║██║   ██║██████╔╝██║
██╔══██╗██╔══╝  ██║  ██║██║   ██║██╔══██╗╚═╝
██║  ██║███████╗██████╔╝╚██████╔╝██████╔╝██╗
╚═╝  ╚═╝╚══════╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝

A simple tool for renaming projects in a hurry, because who has time for
tedious manual renaming? Not you, that's who.`

	ansiWarning := `
Redub operates with all the subtlety of a bull in a china shop and the precision of a blindfolded dart thrower. Use at your own risk. You have been warned.
`

	blue()
	fmt.Println(ansiRedub)
	reset()
	yellow()
	fmt.Println(ansiWarning)
	reset()

	config := Config{
		Path:    "",
		OldName: "",
		NewName: "",
	}
	var confirm string

	if len(os.Args) > 1 {
		config.Path = os.Args[1]
	} else {
		config.Path = prompt("Enter the path to the project")
	}

	_, err := os.Stat(config.Path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Path %s does not exist.", config.Path))
	}

	if len(os.Args) > 2 {
		config.OldName = os.Args[2]
	} else {
		config.OldName = prompt("Enter the current project name")
	}

	if len(os.Args) > 3 {
		config.NewName = os.Args[3]
	} else {
		config.NewName = prompt("Enter the new project name")
	}

	if len(os.Args) > 4 {
		confirm = os.Args[4]
	} else {
		confirm = prompt("Are you sure you want to proceed? (yes/no)")
	}

	if confirm != "yes" {
		return nil, errors.New("Operation cancelled.")
	}

	return &config, nil
}

func prompt(message string) string {
	blue()
	fmt.Print(message + ": ")
	reset()
	var input string
	fmt.Scanln(&input)
	return input
}

type Config struct {
	Path    string
	OldName string
	NewName string
}

func blue() {
	fmt.Print("\033[1;34m")
}

func green() {
	fmt.Print("\033[1;32m")
}

func red() {
	fmt.Print("\033[1;31m")
}

func gray() {
	fmt.Print("\033[1;30m")
}

func yellow() {
	fmt.Print("\033[1;33m")
}

func reset() {
	fmt.Print("\033[0m")
}
