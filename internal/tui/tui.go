package tui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/neoduck0/ivo/internal/config"
	"github.com/neoduck0/ivo/internal/model"
)

const userInput = "> "

// Run starts an interactive conversation and returns on /exit or EOF.
func Run(cfg config.Config) error {
	reader := bufio.NewReader(os.Stdin)
	conversation := model.NewConversation(cfg)
	for {
		input, err := promptUser(reader)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if input == "/exit" {
			break
		}

		response, err := conversation.SendMessage(input)
		if err != nil {
			if printErr := printError(err.Error()); printErr != nil {
				return printErr
			}
			continue
		}

		if err := printResponse(response); err != nil {
			return err
		}
	}
	return nil
}

func promptUser(reader *bufio.Reader) (string, error) {
	fmt.Print(color.CyanString(userInput))

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	return input, nil
}

func printResponse(response string) error {
	response = strings.Trim(response, "\r\n")
	if response == "" {
		return nil
	}
	_, err := fmt.Println(response)
	return err
}

func printError(errorMessage string) error {
	_, err := fmt.Println(color.RedString(errorMessage))
	return err
}
