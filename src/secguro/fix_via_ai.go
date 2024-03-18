package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	openai "github.com/sashabaranov/go-openai"
)

const openAiApiKeyEnvVarName = "OPEN_AI_API_KEY"

func fixProblemViaAi(previousStep func() error, unifiedFinding UnifiedFinding) error {
	return fixProblemViaAiStep1(previousStep, unifiedFinding)
}

func fixProblemViaAiStep1(previousStep func() error, unifiedFinding UnifiedFinding) error {
	newFileContent, diff, err := getFixedFileContentAndDiff(unifiedFinding)
	if err != nil {
		return err
	}

	return fixProblemViaAiStep2(previousStep,
		func() error { return fixProblemViaAiStep1(previousStep, unifiedFinding) },
		unifiedFinding.File, newFileContent, diff)
}

func fixProblemViaAiStep2(previousStep func() error, retry func() error,
	filePath string, newFileContent string, diff string) error {
	prompt := "Does the following fix of file " + filePath + " look okay?\n\n" + diff

	choices := []string{"back", "retry", "accept"}
	choiceIndex, err := getOptionChoice(prompt, choices)

	if err != nil {
		return err
	}
	if choiceIndex == -1 || choiceIndex == 0 {
		return previousStep()
	}

	if choiceIndex == 1 {
		return retry()
	}

	if choiceIndex == 2 {
		fmt.Println("Applying fix...")
		err := replaceFileContents(filePath, newFileContent)
		if err != nil {
			return err
		}
		fmt.Println("Appplied fix")
	}

	return errors.New("unexpected choice index")
}

func getFixedFileContentAndDiff(unifiedFinding UnifiedFinding) (string, string, error) {
	fileContentByteArr, err := os.ReadFile(directoryToScan + "/" + unifiedFinding.File)
	if err != nil {
		return "", "", err
	}

	fileContent := string(fileContentByteArr)

	newFileContent, err := getFixedFileContentFromChatGpt(fileContent,
		unifiedFinding.LineStart, unifiedFinding.Hint)
	if err != nil {
		return "", "", err
	}

	diff := getDiff(fileContent, newFileContent)

	return newFileContent, diff, nil
}

func getFixedFileContentFromChatGpt(fileContent string, problemLineNumber int, hint string) (string, error) {
	fmt.Println("Requesting fix suggestion from ChatGPT...")

	client := openai.NewClient(os.Getenv(openAiApiKeyEnvVarName))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{ //nolint: exhaustruct
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{ //nolint: exhaustruct
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Fix the problem in line %d of the following code:\n",
						problemLineNumber) +
						"```\n" + fileContent + "\n```\n\nHint: " + hint + "\n\n" +
						"Just provide the corrected code nicely formatted without any further explanation.\n" +
						"Do not remove comments.\n" +
						"Under all circumstances make sure you do not introduce any other security vulnerability.",
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	fmt.Println("Received fix suggestion")

	newFileContent := assimilateEnding(fileContent,
		removeCodeBlockBackticksIfAny(resp.Choices[0].Message.Content))

	return newFileContent, nil
}

func getDiff(contentBefore string, contentAfter string) string {
	edits := myers.ComputeEdits(span.URIFromPath("before"), contentBefore, contentAfter)
	diff := fmt.Sprint(gotextdiff.ToUnified("before", "after", contentBefore, edits))
	return diff
}

func replaceFileContents(filePath string, newFileContent string) error {
	const fileMode fs.FileMode = 0666
	file, err := os.OpenFile(directoryToScan+"/"+filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(newFileContent)

	return err
}

func removeCodeBlockBackticksIfAny(s string) string {
	if len(s) >= 6 && s[0:3] == "```" && s[len(s)-3:] == "```" {
		return s[3 : len(s)-3]
	}

	return s
}

func assimilateEnding(sample string, s string) string {
	sampleEndsInLinefeed := sample[len(sample)-1:] == "\n"
	sEndsInLinefeed := s[len(s)-1:] == "\n"

	if sampleEndsInLinefeed == sEndsInLinefeed {
		return s
	}

	if sampleEndsInLinefeed {
		return s + "\n"
	}

	return s[0 : len(s)-1]
}
