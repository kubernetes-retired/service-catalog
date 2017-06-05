package utils

import "fmt"

func Green(str string) string {
	return fmt.Sprintf("\x1b[32;1m%s\x1b[0m", str)
}

func Red(str string) string {
	return fmt.Sprintf("\x1b[31;1m%s\x1b[0m", str)
}

func Entity(str string) string {
	return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", str)
}

func Error(msg string) {
	fmt.Printf("%s\n\n%s\n\n", Red("ERROR"), msg)
}
func Ok() {
	fmt.Printf("%s\n\n", Green("OK"))
}
