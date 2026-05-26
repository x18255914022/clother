package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

type Prompter struct {
	In  io.Reader
	Out io.Writer
}

func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{In: in, Out: out}
}

func (p *Prompter) Prompt(label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.Out, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(p.Out, "%s: ", label)
	}
	reader := bufio.NewReader(p.In)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (p *Prompter) PromptSecret(label string) (string, error) {
	fmt.Fprintf(p.Out, "%s: ", label)

	if err := setConsoleMode(false); err != nil {
		// Fallback to normal prompt if we can't hide input
		return p.Prompt(label, "")
	}
	defer func() { _ = setConsoleMode(true) }()

	reader := bufio.NewReader(os.Stdin)
	value, readErr := reader.ReadString('\n')
	fmt.Fprintln(p.Out)
	if readErr != nil && readErr != io.EOF {
		return "", readErr
	}
	return strings.TrimSpace(value), nil
}

func (p *Prompter) Confirm(label string, defaultYes bool) (bool, error) {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}
	answer, err := p.Prompt(label+" "+hint, "")
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "" {
		return defaultYes, nil
	}
	return strings.HasPrefix(answer, "y"), nil
}

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode      = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode      = kernel32.NewProc("SetConsoleMode")
	ENABLE_ECHO_INPUT       = 0x0004
	ENABLE_LINE_INPUT       = 0x0002
	getStdHandle            = kernel32.NewProc("GetStdHandle")
	STD_INPUT_HANDLE        = -10
)

func setConsoleMode(echo bool) error {
	handle, _, _ := getStdHandle.Call(uintptr(STD_INPUT_HANDLE))
	if handle == 0 {
		return fmt.Errorf("failed to get console handle")
	}

	var mode uint32
	r, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return fmt.Errorf("failed to get console mode")
	}

	if echo {
		mode |= uint32(ENABLE_ECHO_INPUT) | uint32(ENABLE_LINE_INPUT)
	} else {
		mode &^= uint32(ENABLE_ECHO_INPUT)
	}

	r, _, _ = procSetConsoleMode.Call(handle, uintptr(mode))
	if r == 0 {
		return fmt.Errorf("failed to set console mode")
	}
	return nil
}
