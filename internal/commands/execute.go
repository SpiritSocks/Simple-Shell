package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	osuser "os/user"
	"strings"
)

// ExecInput исполняет одну строку ввода.
// На Этапе 1 команды ls и cd сделаны заглушками.
func ExecInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	args, err := parseArgs(input)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return errors.New("cd: path required")
		}
		fmt.Printf("cd %s\n", strings.Join(args[1:], " "))
		return nil

	case "ls":
		fmt.Printf("ls %s\n", strings.Join(args[1:], " "))
		return nil

	case "exit":
		os.Exit(0)
	}

	// остальные команды — пробуем запустить через exec
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.Error
		if errors.As(err, &ee) {
			return fmt.Errorf("unknown command: %s", args[0])
		}
		var xe *exec.ExitError
		if errors.As(err, &xe) {
			return fmt.Errorf("%s: exited with status %d", args[0], xe.ExitCode())
		}
		return err
	}
	return nil
}

// GetHostAndUser -> (username, hostname) для приглашения
func GetHostAndUser() (string, string) {
	currUser, err := osuser.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения пользователя: %v\n", err)
		return "", ""
	}
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения hostname: %v\n", err)
		return "", ""
	}
	username := currUser.Username
	return username, hostName
}

// Парсер с кавычками и экранированием
func parseArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if escapeNext {
			current.WriteByte(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(ch)
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if inQuotes {
		return nil, errors.New("unmatched quote")
	}
	return args, nil
}
