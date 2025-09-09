package color

// Green returns a string formatted with green ANSI color code
func Green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func Red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func Yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func Blue(s string) string {
	return "\033[34m" + s + "\033[0m"
}

func Purple(s string) string {
	return "\033[35m" + s + "\033[0m"
}

func Cyan(s string) string {
	return "\033[36m" + s + "\033[0m"
}

func Orange(s string) string {
	return "\033[38;5;208m" + s + "\033[0m"
}

func White(s string) string {
	return "\033[37m" + s + "\033[0m"
}

func Gray(s string) string {
	return "\033[38;5;244m" + s + "\033[0m"
}
