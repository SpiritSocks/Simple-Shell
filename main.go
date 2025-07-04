package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execInput(input string) error {

	//We trim the newline character
	input = strings.TrimSuffix(input, "\n")

	//Split the input to separate the command and arguments
	args := strings.Split(input, " ")

	//We check for built in commands
	switch args[0] {
	case "cd":
		// 'cd' to home dir with empty path not yet supported
		if len(args) < 2 {
			return errors.New("error. path is required")
		}
		// Change the directory and return the error.
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	}

	//Prepare the command to execute
	cmd := exec.Command(args[0], args[1:]...)

	//Set the correct output device
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	//Execute the command and return the error
	return cmd.Run()
}

func getHostName() (string, error) {

	hostName, err := os.Hostname()
	if err != nil {
		return "", errors.New("error. failed to retrive host name")
	}

	hostName = strings.TrimSuffix(hostName, ".local")

	return hostName, nil
}

func getCurrentDir() (string, error) {

	dir, err := os.Getwd()
	if err != nil {
		return "", errors.New("error. failed to retrive current directory")
	}

	dir = strings.TrimPrefix(dir, "/*****/*****/*****/*****/*****") // For safety measures blurred the path
	if len(dir) != 0 && dir[0] == '/' {
		dir = strings.TrimPrefix(dir, "/")
	}

	return dir, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	hostName, err := getHostName()

	if err != nil {
		os.Exit(1)
	}

	for {
		dir, _ := getCurrentDir()

		if dir == "" {
			fmt.Printf(">(base) %s ~ %% ", hostName)
		} else {
			fmt.Printf(">(base) %s %s %% ", hostName, dir)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if err = execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

	}

}
