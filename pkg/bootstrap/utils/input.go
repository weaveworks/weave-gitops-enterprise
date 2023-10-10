package utils

import (
	"errors"

	"github.com/manifoldco/promptui"
)

const (
	inputStringErrMsg = "value can't be empty"
	blueInfo          = "\x1b[34;1m✔  %s\x1b[0m\n"
)

// GetPasswordInput prompt to enter password.
func GetPasswordInput(msg string) (string, error) {
	prompt := promptui.Prompt{
		Label: msg,
		Mask:  '*',
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
