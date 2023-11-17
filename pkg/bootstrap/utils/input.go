package utils

import (
	"errors"
	"io"
	"strings"

	"github.com/manifoldco/promptui"
)

const (
	inputStringErrMsg = "value can't be empty"
	blueInfo          = "\x1b[34;1m✔  %s\x1b[0m\n"
)

// GetPasswordInput prompt to enter password.
func GetPasswordInput(msg string, required bool, stdin io.ReadCloser) (string, error) {
	validate := func(input string) error {
		if required && len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    msg,
		Validate: validate,
		Mask:     '*',
		Stdin:    stdin,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)

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

	return result, nil
}

// GetStringInput prompt to get string input.
func GetStringInput(msg string, defaultValue string, stdin io.ReadCloser) (string, error) {
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
		Stdin:    stdin,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	return result, nil
}

// GetConfirmInput prompt to get yes or no input.
func GetConfirmInput(msg string, stdin io.ReadCloser) string {
	prompt := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
		Stdin:     stdin,
	}

	result, err := prompt.Run()
	if err != nil {
		return "n"
	}
	result = strings.TrimSpace(result)
	stdin.Close()
	return result
}
