package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	osuser "os/user"
	"strings"
	"var27_shell/internal/vfs"
)

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
	case "pwd":
		fmt.Println(vfs.Pwd())
		return nil

	case "cd":
		target := "."
		if len(args) >= 2 {
			target = strings.Join(args[1:], " ")
		}
		return vfs.Cd(target)

	case "ls":
		target := "."
		if len(args) >= 2 {
			target = strings.Join(args[1:], " ")
		}
		entries, err := vfs.Ls(target)
		if err != nil {
			return err
		}
		for _, e := range entries {
			fmt.Println(e)
		}
		return nil

	case "exit":
		os.Exit(0)
	}
	// остальные — хостовые команды
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
