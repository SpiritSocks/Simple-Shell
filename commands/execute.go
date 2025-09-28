package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func ExecInput(input string) error {
	input = strings.TrimSuffix(input, "\n")

	args, err := parseArgs(input)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "ls":
    	fmt.Printf("Command: ls, Arguments: %v\n", args[1:])
    	return nil
	case "cd":
    	if len(args) < 2 {
			return errors.New("path required")
		}
    	fmt.Printf("Command: cd, Arguments: %v\n", args[1:])
    	return nil
	case "exit":
		os.Exit(0)
	default:
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

func GetHostAndUser() (string, string) {

	currUser, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения пользователя: %v\n", err)
		return "", ""
	}

	hostName, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения hostname: %v\n", err)
		return "", ""
	}

	user := currUser.Username

	return user, hostName
}

func parseArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if escapeNext {
			current.WriteRune(rune(char))
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inQuotes = !inQuotes
			continue
		}

		if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(rune(char))
	}

	// Добавляем последний аргумент
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	// Если остались незакрытые кавычки — ошибка
	if inQuotes {
		return nil, errors.New("unmatched quote")
	}

	return args, nil
}
