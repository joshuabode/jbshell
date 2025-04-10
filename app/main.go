package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"golang.org/x/term"
)

var builtins = []string{"type", "exit", "echo", "pwd", "cd"}

var ErrCtrlC = errors.New("Ctrl-C")

func parseInput(input []rune) (string, []string) {
	args := []string{}

	var inQuote = false
	var arg strings.Builder
	for _, r := range input {
		if r == '\'' {
			inQuote = !inQuote
		} else if (r == ' ' && !inQuote) {
			args = append(args, arg.String())
			arg.Reset()
		} else {
			arg.WriteRune(r)
		}
	}
	args = append(args, arg.String())
	return args[0], args[1:]
}

func interactiveReader() ([]rune, error) {
	fd := int(os.Stdin.Fd())

	// Put terminal in raw mode.
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	// Ensure we restore the terminal before exit.
	defer term.Restore(fd, oldState)

	// Print prompt
	fmt.Fprint(os.Stdout, "JBShell * > ")

	reader := bufio.NewReader(os.Stdin)
	var result []rune

	// Read input character by character.
	for {
		char0, _, err0 := reader.ReadRune()
		if err0 != nil {
			return []rune{}, err0
		}

		// On tab | TODO: implemet tab completion
		if char0 == '\t' {
			continue
		}

		// On any arrow key | TODO: allow arrow key edits
		if char0 == 27 { // catch escape character
			char1, _, err1 := reader.ReadRune()
			if err1 != nil {
				return []rune{}, err1
			}
			if char1 == '[' {
				char2, _, err2 := reader.ReadRune()
				if err2 != nil {
					return []rune{}, err2
				}
				if 'A' <= char2 && char2 <= 'D' {
					continue
				}
			}
		}

		// On Enter (handle both CR and LF)
		if char0 == '\r' || char0 == '\n' {
			break
		}

		// On Crtl-C
		if char0 == 3 {
			return nil, ErrCtrlC
		}

		// Handle backspace: typically '\b' (8) or DEL (127)
		if char0 == '\b' || char0 == 127 {
			if len(result) > 0 {
				// Remove last character from the input buffer.
				result = result[:len(result)-1]
				// Move cursor back, overwrite the character with space, then move back again.
				fmt.Fprint(os.Stdout, "\b \b")
			}
		} else {
			// Append to our buffer and echo the character.
			result = append(result, char0)
			fmt.Fprint(os.Stdout, string(char0))
		}
	}
	defer fmt.Fprintln(os.Stdout, "\r")
	return result, nil
}

func execute(executable string, args []string) error {
	path, err := exec.LookPath(executable)
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
	}
	if slices.Contains(builtins, args[0]) {
		return args[0] + " is a shell builtin command"
	}
	path, err := exec.LookPath(args[0])
	if err == nil || errors.Is(err, exec.ErrDot) {
		return args[0] + " is " + path
	} else {
		return args[0] + ": is not recognised"
	}
}

func _exit(args []string) {
	var code int
	var err error
	if len(args) == 1 {
		code, err = strconv.Atoi(args[0])
	} else {
		code = 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid argument passed, argument must be an integer")
		os.Exit(1)
	}
	os.Exit(code)
}

func _echo(args []string) string {
	argstr := strings.Join(args, " ")
	return argstr
}

func _pwd(args []string) string {
	if len(args) != 0 {
		return "usage: pwd\t(no arguments)"
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not get current directory: "+err.Error())
		return ""
	}
	return wd
}

func _cd(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: cd [path]")
		return
	}
	var err error
	if args[0] == "~" {
		err = os.Chdir(os.Getenv("HOME"))
	} else {
		err = os.Chdir(args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not change directory to: "+args[0]+" "+err.Error())
	}
}

func main() {
	for {
		input, err := interactiveReader()
		if errors.Is(err, ErrCtrlC) {
			fmt.Fprintln(os.Stdout)
			continue
		}
		cmd, args := parseInput(input)
		switch cmd {
		case "exit":
			_exit(args)
		case "echo":
			fmt.Println(_echo(args))
		case "type":
			fmt.Println(_type(args))
		case "pwd":
			fmt.Println(_pwd(args))
		case "cd":
			_cd(args)
		case "":
			continue
		default:
			err = execute(cmd, args)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
