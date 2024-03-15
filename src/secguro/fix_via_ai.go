package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

const openAiApiKeyEnvVarName = "OPEN_AI_API_KEY"

func fixProblemViaAi(_previousStep func() error, unifiedFinding UnifiedFinding) error {
	// TODO: implement wizard and use previousStep
	return fixProblemViaChatGpt(unifiedFinding)
}

func fixProblemViaChatGpt(unifiedFinding UnifiedFinding) error {
	fileContents, err := os.ReadFile(directoryToScan + "/" + unifiedFinding.File)
	if err != nil {
		return err
	}

	client := openai.NewClient(os.Getenv(openAiApiKeyEnvVarName))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{ //nolint: exhaustruct
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{ //nolint: exhaustruct
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Fix the problem in line %d of the following code:\n",
						unifiedFinding.LineStart) +
						"```\n" + string(fileContents) + "\n```\n\nHint: " + unifiedFinding.Hint + "\n\n" +
						"Just provide the corrected code nicely formatted without any further explanation.\n" +
						"Do not remove comments.\n" +
						"Under all circumstances make sure you do not introduce any other security vulnerability.",
				},
			},
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(resp.Choices[0].Message.Content)

	return nil
}
