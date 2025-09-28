package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"go.mod/internal/commands"
)

func main() {
	// --- Этап 2: флаги конфигурации ---
	var vfsPath string
	var scriptPath string
	flag.StringVar(&vfsPath, "vfs", "", "Path to physical VFS location")
	flag.StringVar(&scriptPath, "script", "", "Path to startup script file")
	flag.Parse()

	// Отладочный вывод конфигурации
	fmt.Printf("conf: vfs_path=%q, start_script=%q\n", vfsPath, scriptPath)

	// --- REPL: приглашение ---
	username, hostname := commands.GetHostAndUser()
	prompt := fmt.Sprintf("%s@%s:~$ ", username, hostname)

	// Если указан стартовый скрипт — выполняем, останавливаемся при первой ошибке
	if scriptPath != "" {
		if err := commands.ExecScript(scriptPath, prompt); err != nil {
			os.Exit(1)
		}
	}

	// --- REPL-цикл ---
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if err = commands.ExecInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
