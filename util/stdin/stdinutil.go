package stdin

import (
	"bufio"
	"fmt"
	"os"
)

func ReadInput(message string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf(message+" [%s]\n", defaultValue)
	} else {
		fmt.Printf(message + "\n")
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		if (defaultValue != "") && (scanner.Text() == "") {
			return defaultValue
		} else if scanner.Text() != "" {
			return scanner.Text()
		} else {
			return ReadInput(message, defaultValue)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input: ", err)
		os.Exit(1)
	}

	return ""
}
