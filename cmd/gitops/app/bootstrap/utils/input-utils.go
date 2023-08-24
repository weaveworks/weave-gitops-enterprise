package utils

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
)

// GetPasswordInput prompt to enter password
func GetPasswordInput(label string) (string, error) {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()
	if err != nil {
		return "", CheckIfError(err, "Prompt failed")
	}

	return result, nil
}

// GetSelectInput get option from multiple choices
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
		return "", CheckIfError(err, "Prompt failed")
	}

	Info("Selected: %s\n", result)

	return result, nil
}

// GetStringInput prompt to get string input
func GetStringInput(msg string, defaultValue string) (string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New("value can't be empty")
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
		return "", CheckIfError(err, "Prompt failed")
	}

	return result, nil
}

// GetConfirmInput prompt to get yes or no input
func GetConfirmInput(msg string) (string, error) {
	prompt := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", CheckIfError(err)
	}

	return result, nil
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(format, args...))
}

// CheckIfError dispaly the error or return nil if none
func CheckIfError(err error, extramsg ...string) error {
	if err == nil {
		return nil
	}
	if len(extramsg) > 0 {
		return fmt.Errorf("\x1b[31;1m%s\x1b[0m", fmt.Sprintf("%s\n%s", err.Error(), extramsg[0]))
	}
	return fmt.Errorf("\x1b[31;1m%s\x1b[0m", err.Error())
}