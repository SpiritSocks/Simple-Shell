package main

import (
	"bufio"
	"fmt"
	"os"

	"go.mod/commands"
)

func main() {

	user, host := commands.GetHostAndUser()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s@%s:~$ ", user, host)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if err = commands.ExecInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

}
