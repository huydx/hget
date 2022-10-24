package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var (
	Stdout     = colorable.NewColorableStdout()
	Stderr     = colorable.NewColorableStderr()
	Default UI = Console{Stdout: Stdout, Stderr: Stderr}
)

// UI represents a simple IO output.
type UI interface {
	Printf(format string, a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)
	Errorf(format string, a ...interface{}) (n int, err error)
	Errorln(a ...interface{}) (n int, err error)
}

// Printf outputs information level logs
func Printf(format string, a ...interface{}) (n int, err error) {
	return Default.Printf(color.CyanString("INFO: ")+format, a...)
}

// Errorf outputs error level logs
func Errorf(format string, a ...interface{}) (n int, err error) {
	return Default.Errorf(color.RedString("ERROR: ")+format, a...)
}

// Warnf outputs warning level logs
func Warnf(format string, a ...interface{}) (n int, err error) {
	return Default.Errorf(color.YellowString("WARN: ")+format, a...)
}

// Errorln is non formatted error printer.
func Errorln(a ...interface{}) (n int, err error) {
	return Default.Errorln(a...)
}

// IsTerminal checks if we have tty
func IsTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd())
}

// Console is an implementation of UI interface
type Console struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Printf prints formatted information logs to Stdout
func (c Console) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stdout, format, a...)
}

// Println prints information logs to Stdout
func (c Console) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(c.Stdout, a...)
}

// Errorf prints formatted error logs to Stderr
func (c Console) Errorf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stderr, format, a...)
}

// Errorln prints error logs to Stderr
func (c Console) Errorln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(c.Stderr, a...)
}
