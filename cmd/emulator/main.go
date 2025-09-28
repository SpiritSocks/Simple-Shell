package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"var27_shell/internal/commands"
	"var27_shell/internal/vfs"
)

func main() {
	// Этап 2+3: флаги
	var vfsPath string
	var scriptPath string
	flag.StringVar(&vfsPath, "vfs", "", "Path to physical VFS directory")
	flag.StringVar(&scriptPath, "script", "", "Path to startup script")
	flag.Parse()

	// Инициализация VFS (теперь из пакета vfs, а не commands)
	if err := vfs.Init(vfsPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Отладка конфигурации (требование Этапа 2)
	fmt.Printf("conf: vfs_path=%q, start_script=%q\n", vfsPath, scriptPath)

	username, hostname := commands.GetHostAndUser()

	// Показываем текущий виртуальный путь вместо "~"
	prompt := func() string {
		return fmt.Sprintf("%s@%s:%s$ ", username, hostname, vfs.Pwd())
	}

	// Если есть стартовый скрипт — выполняем до REPL
	if scriptPath != "" {
		if err := commands.ExecScript(scriptPath, prompt()); err != nil {
			os.Exit(1) // остановка при первой ошибке
		}
	}

	// REPL
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt())
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
