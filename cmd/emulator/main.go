package main

import (
	"bufio"
	"fmt"
	"os"

	"go.mod/internal/commands"
)

func main() {
	username, hostname := commands.GetHostAndUser()
	prompt := fmt.Sprintf("%s@%s:~$ ", username, hostname)

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
