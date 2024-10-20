package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := ""
	reader := bufio.NewReader(os.Stdin)
	for input == "" {
		fmt.Printf("%-20s%-5s%-20s\n", "get [service]", "-", "Get the config file of the service")
		fmt.Printf("%-20s%-5s%-20s\n", "set [service]", "-", "Set the config file of the service")
		input, _ = reader.ReadString('\n')
		input = input[:len(input)-1]
		args := strings.Fields(input)
		if len(args) == 2 {
			switch args[0] {
			case "get":
				getCommand(args[1])
			case "set":
				setCommand(args[1])
			default:
				input = ""
			}
		}
	}
}

func getCommand(arg string) {

}

func setCommand(arg string) {

}
