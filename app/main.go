package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"golang.org/x/term"
)

var builtins = []string{"type", "exit", "echo", "pwd", "cd"}

func unquote(args []string) []string {
	for i, str := range args {
		args[i] = strings.Trim(str, "'")
	}
	return args
}

func getExecPath(e string) (string, error) {
	var pathArray = strings.Split(os.Getenv("PATH"), ":")
	for _, path := range pathArray {
		entries, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading PATH environment variable: "+path+":\n"+err.Error())
			continue
		}
		for _, entry := range entries {
			if e == entry.Name() {
				return path + "/" + entry.Name(), nil
			}
		}
	}
	return "", errors.New("no path found for " + e)
}

func interactiveReader() (string, error) {
	fd := int(os.Stdin.Fd())

	// Put terminal in raw mode.
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	// Ensure we restore the terminal before exit.
	defer func() {
		// Restore terminal state
		term.Restore(fd, oldState)
	}()

	// Print prompt
	fmt.Fprint(os.Stdout, "JBShell★ ")

	reader := bufio.NewReader(os.Stdin)
	var result []rune

	// Read input character by character.
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		// On Enter (handle both CR and LF)
		if char == '\r' || char == '\n' || char == 3 {
			// Print newline (to complete the prompt line)
			break
		}

		// Handle backspace: typically '\b' (8) or DEL (127)
		if char == '\b' || char == 127 {
			if len(result) > 0 {
				// Remove last character from the input buffer.
				result = result[:len(result)-1]
				// Move cursor back, overwrite the character with space, then move back again.
				fmt.Fprint(os.Stdout, "\b \b")
			}
		} else {
			// Append to our buffer and echo the character.
			result = append(result, char)
			fmt.Fprint(os.Stdout, string(char))
		}
	}
	return string(result), nil
}

func execute(executable string, args []string) error {
	path, err := getExecPath(executable)
	if err != nil {
		return errors.New(executable + ": command not found")
	}
	cmd := exec.Command(path, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}

func _type(args []string) string {
	if len(args) != 1 {
		return "usage: type [command]"
	} else {
		for _, cmd := range builtins {
			if args[0] == cmd {
				return args[0] + " is a shell builtin command"
			}
		}
		path, err := getExecPath(args[0])
		if err != nil {
			return args[0] + ": is not recognised"
		} else {
			return args[0] + " is " + path
		}
	}
}

func _exit(args []string) string {
	var code int
	var err error
	if len(args) > 0 {
		code, err = strconv.Atoi(args[0])
	} else {
		code = 0
	}
	if err != nil {
		return "invalid argument passed, argument must be an integer"
	}
	os.Exit(code)
	return ""
}

func _echo(args []string) string {
	argstr := strings.Join(args, " ")
	return argstr + " "
}

func _pwd(args []string) string {
	if len(args) != 0 {
		return "usage: pwd\t(no arguments)"
	}
	wd, err := os.Getwd()
	if err != nil {
		panic("Could not get working directory" + err.Error())
	}
	return wd
}

func _cd(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: cd [path]")
	}
	var err error
	if args[0] == "~" {
		err = os.Chdir(os.Getenv("HOME"))
	} else {
		err = os.Chdir(args[0])
	}
	if err != nil {
		panic("Could not change directory to: " + args[0] + " " + err.Error())
	}
}

func main() {
	for {
		input, err := interactiveReader()
		var command string = strings.Split(input, " ")[0]
		args := unquote(strings.Split(input, " ")[1:])
		switch command {
		case "exit":
			fmt.Fprintln(os.Stderr, _exit(args))
		case "echo":
			fmt.Println(_echo(args))
		case "type":
			fmt.Println(_type(args))
		case "pwd":
			fmt.Println(_pwd(args))
		case "cd":
			_cd(args)
		default:
			err = execute(command, args)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
