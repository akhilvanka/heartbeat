package main

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"os"
)

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
)

type exitCode int

var (
	operation int
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	startPage := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Welcome to Heartbeat! Choose which operation mode this system is acting in:").
				Options(
					huh.NewOption("Server", 0),
					huh.NewOption("Client", 1),
					huh.NewOption("Cancel", 2),
				).
				Value(&operation),
		),
	).WithTheme(huh.ThemeBase16())

	err := startPage.Run()
	if err != nil {
		return exitError
	} else {
		switch operation {
		case 0:
			fmt.Println("Server is initiating")
		case 1:
			fmt.Println("Client is initiating")
		case 2:
			return exitCancel
		default:
			return exitOK
		}
	}
	return exitOK
}
