package main

import "os"

// banner is Alexiares' startup wordmark, generated with figlet
// (font: doom) rather than hand-drawn — figlet's letterforms are a
// 30-year-proven character set that renders identically across
// terminal fonts, unlike freehand Unicode art, which does not.
const banner = "  ___   _      _______   _______  ___  ______ _____ _____ \n" +
	" / _ \\ | |    |  ___\\ \\ / /_   _|/ _ \\ | ___ \\  ___/  ___|\n" +
	"/ /_\\ \\| |    | |__  \\ V /  | | / /_\\ \\| |_/ / |__ \\ `--. \n" +
	"|  _  || |    |  __| /   \\  | | |  _  ||    /|  __| `--. \\\n" +
	"| | | || |____| |___/ /^\\ \\_| |_| | | || |\\ \\| |___/\\__/ /\n" +
	"\\_| |_/\\_____/\\____/\\/   \\/\\___/\\_| |_/\\_| \\_\\____/\\____/ \n" +
	"\n" +
	"        guard the gate before the wallet connects\n\n"

// isInteractiveTerminal reports whether stdout is a real terminal —
// not a pipe, a file redirect, or a CI log capture. The banner is
// decorative and must never appear in scripted or piped output, so
// every call site gates on this before printing it.
func isInteractiveTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
