package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GitHubAuth represents GitHub authentication details
type GitHubAuth struct {
	Token string
}

// TODO: dynamically figure out name of default branch (appears to be in the result when querying the entire repo)
const nameDefaultBranch = "master"

const githubPersonalAccessTokenEnvVarName = "GITHUB_PERSONAL_ACCESS_TOKEN"

func createPullRequest(owner, repo, branch, filePath, newContent,
	title, description string, auth GitHubAuth) error {
	// authenticate with GitHub
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: auth.Token}, //nolint: exhaustruct
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	defaultBranch, _, err := client.Repositories.GetBranch(ctx, owner, repo, nameDefaultBranch)
	if err != nil {
		return err
	}

	// create new branch
	_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{ //nolint: exhaustruct
		Ref: github.String("refs/heads/" + branch),
		Object: &github.GitObject{ //nolint: exhaustruct
			SHA: github.String(defaultBranch.Commit.GetSHA()),
		},
	})
	if err != nil {
		return err
	}

	// prepare commit details
	// TODO: set Author and Committer if this is not done automatically when moving this to an action.
	// when using a personal access token, it sets them to the account name
	commit := &github.RepositoryContentFileOptions{ //nolint: exhaustruct
		Message: github.String("Updating file"),
		Content: []byte(newContent),
		Branch:  github.String(branch),
		SHA:     github.String(""), // empty string refers to latest commit
	}

	// create or update file
	_, _, err = client.Repositories.CreateFile(ctx, owner, repo, filePath, commit)
	if err != nil {
		return err
	}

	// create pull request
	pr := &github.NewPullRequest{ //nolint: exhaustruct
		Title: github.String(title),
		Body:  github.String(description),
		Head:  github.String(branch),
		Base:  github.String(nameDefaultBranch),
	}

	_, _, err = client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return err
	}

	return nil
}

func createPrTest() {
	filePath := "path/to/file.txt"
	newContent := "New file contents"
	owner := "secguro"
	repo := "secguro-cli"
	branch := "test_branch_4"
	title := "Pull Request Title"
	description := "Description of the pull request"
	authToken := os.Getenv(githubPersonalAccessTokenEnvVarName)

	auth := GitHubAuth{Token: authToken}

	err := createPullRequest(owner, repo, branch, filePath, newContent, title, description, auth)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Pull request created successfully!")
}
