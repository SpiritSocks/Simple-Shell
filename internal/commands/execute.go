package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	osuser "os/user"
	"strings"
)

// ExecInput: обработка одной команды (ls/cd как заглушки; exit завершает процесс)
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

// ExecScript: выполнить стартовый скрипт с эхо и остановкой на первой ошибке
func ExecScript(scriptPath, prompt string) error {
	f, err := os.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("script: %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// эхо "как в консоли"
		fmt.Printf("%s%s\n", prompt, line)

		if err := ExecInput(line); err != nil {
			fmt.Fprintf(os.Stderr, "error on line %d: %v\n", lineNo, err)
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("script read error: %w", err)
	}
	return nil
}

// GetHostAndUser -> (username, hostname) для приглашения
func GetHostAndUser() (string, string) {
	u, err := osuser.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "user error: %v\n", err)
		return "", ""
	}
	h, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "host error: %v\n", err)
		return "", ""
	}
	return u.Username, h
}

// parseArgs: парсер с кавычками и экранированием
func parseArgs(input string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inQuotes := false
	escape := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if escape {
			cur.WriteByte(ch)
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ' ' && !inQuotes {
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(ch)
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	if inQuotes {
		return nil, errors.New("unmatched quote")
	}
	return args, nil
}
