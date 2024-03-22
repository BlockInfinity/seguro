package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const githubPersonalAccessTokenEnvVarName = "GITHUB_PERSONAL_ACCESS_TOKEN"

func createPrTest() {
	filePath := "path/to/file.txt"
	newContent := "New file contents"
	owner := "secguro"
	repo := "secguro-cli"
	nameNewBranch := "test_branch_5"
	title := "Pull Request Title"
	description := "Description of the pull request"

	err := createPullRequestForFileContentChange(owner, repo,
		nameNewBranch, filePath, newContent, title, description)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Pull request created successfully!")
}

func createPullRequestForFileContentChange(owner, repo, nameNewBranch, filePath, newContent,
	title, description string) error {
	ctx := context.Background()

	client := getGithubClient(ctx)

	nameDefaultBranch, err := getNameOfDefaultBranch(ctx, client, owner, repo)
	if err != nil {
		return err
	}

	err = createBranch(ctx, client, owner, repo, nameDefaultBranch, nameNewBranch)
	if err != nil {
		return err
	}

	err = commitFileChange(ctx, client, owner, repo, nameNewBranch, filePath, newContent)
	if err != nil {
		return err
	}

	err = createPullRequest(ctx, client, owner, repo, nameDefaultBranch, nameNewBranch, title, description)
	if err != nil {
		return err
	}

	return nil
}

func getNameOfDefaultBranch(ctx context.Context, client *github.Client,
	owner, repo string) (string, error) {
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	return repository.GetDefaultBranch(), nil
}

func getGithubClient(ctx context.Context) *github.Client {
	authToken := os.Getenv(githubPersonalAccessTokenEnvVarName)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken}, //nolint: exhaustruct
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client
}

func getBranchSha(ctx context.Context, client *github.Client,
	owner, repo, branchName string) (string, error) {
	branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, branchName)
	if err != nil {
		return "", err
	}

	return branch.Commit.GetSHA(), nil
}

func createBranch(ctx context.Context, client *github.Client,
	owner, repo, nameBaseBranch, nameNewBranch string) error {
	baseBranchSha, err := getBranchSha(ctx, client, owner, repo, nameBaseBranch)
	if err != nil {
		return err
	}

	_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{ //nolint: exhaustruct
		Ref: github.String("refs/heads/" + nameNewBranch),
		Object: &github.GitObject{ //nolint: exhaustruct
			SHA: github.String(baseBranchSha),
		},
	})

	return err
}

func commitFileChange(ctx context.Context, client *github.Client,
	owner, repo, nameNewBranch, filePath, newContent string) error {
	// TODO: set Author and Committer if this is not done automatically when moving this to an action.
	// when using a personal access token, it sets them to the account name
	commit := &github.RepositoryContentFileOptions{ //nolint: exhaustruct
		Message: github.String("Updating file"),
		Content: []byte(newContent),
		Branch:  github.String(nameNewBranch),
		SHA:     github.String(""), // empty string refers to latest commit
	}

	// create or update file
	_, _, err := client.Repositories.CreateFile(ctx, owner, repo, filePath, commit)

	return err
}

func createPullRequest(ctx context.Context, client *github.Client,
	owner, repo, nameBaseBranch, nameNewBranch, title, description string) error {
	pr := &github.NewPullRequest{ //nolint: exhaustruct
		Title: github.String(title),
		Body:  github.String(description),
		Head:  github.String(nameNewBranch),
		Base:  github.String(nameBaseBranch),
	}

	_, _, err := client.PullRequests.Create(ctx, owner, repo, pr)

	return err
}
