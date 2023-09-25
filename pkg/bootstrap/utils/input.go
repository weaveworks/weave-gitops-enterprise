package utils

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
)

const (
	passwordErrorMsg  = "password must have more than 6 characters"
	inputStringErrMsg = "value can't be empty"
	blueInfo          = "\x1b[34;1m✔  %s\x1b[0m\n"
)

// GetPasswordInput prompt to enter password.
func GetPasswordInput(msg string) (string, error) {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New(passwordErrorMsg)
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    msg,
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// GetSelectInput get option from multiple choices.
func GetSelectInput(msg string, items []string) (string, error) {
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.Select{
			Label: msg,
			Items: items,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		return "", err
	}

	Info("Selected: %s\n", result)

	return result, nil
}

// GetStringInput prompt to get string input.
func GetStringInput(msg string, defaultValue string) (string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New(inputStringErrMsg)
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    msg,
		Validate: validate,
		Default:  defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// GetConfirmInput prompt to get yes or no input.
func GetConfirmInput(msg string) string {
	prompt := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return "n"
	}

	return result
}

// Info should be used to describe the example commands that are about to run.
func Info(msg string, args ...interface{}) {
	fmt.Printf(blueInfo, fmt.Sprintf(msg, args...))
}

// Warning should be used to display a warning.
func Warning(msg string, args ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(msg, args...))
}
