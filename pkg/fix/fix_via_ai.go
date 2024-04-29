package fix

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/secguro/secguro-cli/pkg/output"
	"github.com/secguro/secguro-cli/pkg/types"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const openAiApiKeyEnvVarName = "OPEN_AI_API_KEY"

func fixProblemViaAi(directoryToScan string,
	previousStep func() error, unifiedFinding types.UnifiedFinding) error {
	return fixProblemViaAiStep1(directoryToScan, previousStep, unifiedFinding)
}

func fixProblemViaAiStep1(directoryToScan string,
	previousStep func() error, unifiedFinding types.UnifiedFinding) error {
	newFileContent, diff, err := getFixedFileContentAndDiff(directoryToScan, unifiedFinding)
	if err != nil {
		return err
	}

	return fixProblemViaAiStep2(directoryToScan, previousStep,
		func() error { return fixProblemViaAiStep1(directoryToScan, previousStep, unifiedFinding) },
		unifiedFinding.File, newFileContent, diff)
}

func fixProblemViaAiStep2(directoryToScan string, previousStep func() error, retry func() error,
	filePath string, newFileContent string, diff string) error {
	prompt := "Does the following fix of file " + filePath + " look okay?\n\n" + diff

	choices := []string{"back", "retry", "accept"}
	choiceIndex, err := getOptionChoice(prompt, choices)

	if err != nil {
		return err
	}
	switch choiceIndex {
	case -1, 0:
		return previousStep()
	case 1:
		return retry()
	case 2:
		fmt.Print("Applying fix...")
		err := replaceFileContents(directoryToScan, filePath, newFileContent)
		if err != nil {
			return err
		}
		fmt.Println("done")

		return nil
	}

	return errors.New("unexpected choice index")
}

func getFixedFileContentAndDiff(directoryToScan string,
	unifiedFinding types.UnifiedFinding) (string, string, error) {
	fileContentByteArr, err := os.ReadFile(directoryToScan + "/" + unifiedFinding.File)
	if err != nil {
		return "", "", err
	}

	fileContent := string(fileContentByteArr)

	newFileContent, err := GetFixedFileContentFromChatGpt(fileContent,
		unifiedFinding.LineStart, unifiedFinding.Hint)
	if err != nil {
		return "", "", err
	}

	diff := getDiff(fileContent, newFileContent)

	return newFileContent, diff, nil
}

func GetFixedFileContentFromChatGpt(fileContent string, problemLineNumber int, hint string) (string, error) {
	patchStr, err := GetFileContentPatchFromChatGpt(fileContent, problemLineNumber, hint)
	if err != nil {
		return "", err
	}

	fmt.Println(patchStr)
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(removeDiffCodeBlockBackticksIfAny(patchStr))
	if err != nil {
		return "", err
	}
	newContent, patchAppliedArr := dmp.PatchApply(patches, fileContent)
	fmt.Println(patchAppliedArr)
	fmt.Println(newContent)

	return newContent, nil
}

func GetFileContentPatchFromChatGpt(fileContent string, problemLineNumber int, hint string) (string, error) {
	fmt.Print("Requesting fix suggestion...")

	client := openai.NewClient(os.Getenv(openAiApiKeyEnvVarName))
	query := fmt.Sprintf("Fix the problem in line %d of the following code:\n",
		problemLineNumber) +
		"```\n" + fileContent + "\n```\n\nHint: " + hint + "\n\n" +
		"Only provide the changes in git's diff format. Make sure to include the patch header. " +
		"The file's name is: file.txt\n" +
		"Do not provide any explanation.\n" +
		"Do not remove comments.\n" +
		"Do not remove unnecessary whitespace.\n" +
		"Under all circumstances make sure that you do not introduce any new security vulnerability.\n"
	fmt.Println(query)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{ //nolint: exhaustruct
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{ //nolint: exhaustruct
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	fmt.Println("done")

	patch := resp.Choices[0].Message.Content

	return patch, nil
}

func getDiff(contentBefore string, contentAfter string) string {
	dmp := diffmatchpatch.New()
	diff := dmp.DiffMain(contentBefore, contentAfter, false)
	hunks := getDiffHunks(diff)

	diffFormatted := ""
	for i, hunk := range hunks {
		if i != 0 {
			diffFormatted += output.ChangeColor(output.Cyan) + "-----" + output.ChangeColor(output.NoColor) + "\n"
		}

		diffFormatted += dmp.DiffPrettyText(hunk)
	}

	return diffFormatted
}

func getDiffHunks(diff []diffmatchpatch.Diff) [][]diffmatchpatch.Diff {
	const contextSize = 4

	diffSplitByLines := getDiffSplitByLines(diff)

	result := make([][]diffmatchpatch.Diff, 0)

	// when currentHunk is nil, the loop is in skip mode
	var currentHunk []diffmatchpatch.Diff = nil //nolint: prealloc

	availableContext := contextSize
	for diffLineIndex, diffLine := range diffSplitByLines {
		skipMode := currentHunk == nil

		if skipMode {
			if diffLine.Type == diffmatchpatch.DiffEqual {
				continue
			} else {
				// exit skip mode and fill current hunk with context
				currentHunk = make([]diffmatchpatch.Diff, 0)
				for i := max(0, diffLineIndex-contextSize-1); i < diffLineIndex; i++ {
					currentHunk = append(currentHunk, diffSplitByLines[i])
				}
			}
		}

		if diffLine.Type == diffmatchpatch.DiffEqual {
			availableContext--
		} else {
			availableContext = contextSize
		}

		currentHunk = append(currentHunk, diffLine)

		if availableContext == 0 || diffLineIndex == len(diffSplitByLines)-1 {
			result = append(result, currentHunk)
			currentHunk = nil
		}
	}

	return result
}

func getDiffSplitByLines(diff []diffmatchpatch.Diff) []diffmatchpatch.Diff {
	result := make([]diffmatchpatch.Diff, 0)

	for _, diffEntry := range diff {
		textFragments := strings.Split(diffEntry.Text, "\n")
		for i, textFragment := range textFragments {
			completeTextFragment := textFragment
			if i != len(textFragments)-1 {
				completeTextFragment += "\n"
			}

			diffFragment := diffmatchpatch.Diff{
				Type: diffEntry.Type,
				Text: completeTextFragment,
			}

			result = append(result, diffFragment)
		}
	}

	return result
}

func replaceFileContents(directoryToScan string, filePath string, newFileContent string) error {
	const fileMode fs.FileMode = 0666
	file, err := os.OpenFile(directoryToScan+"/"+filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(newFileContent)

	return err
}

func removeDiffCodeBlockBackticksIfAny(s string) string {
	indexOfFirstHunk := strings.Index(s, "@@")
	sWithoutPrefix := s[indexOfFirstHunk:]

	if sWithoutPrefix[len(sWithoutPrefix)-3:] == "```" {
		fmt.Println("case 1")
		fmt.Println(sWithoutPrefix[:len(sWithoutPrefix)-3])
		return sWithoutPrefix[:len(sWithoutPrefix)-3]
	} else {
		fmt.Println("case 2")
		fmt.Println(sWithoutPrefix)
		return sWithoutPrefix
	}
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
