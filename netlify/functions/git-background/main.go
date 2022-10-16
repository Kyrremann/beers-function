package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

func gitIt(user string, untappd []byte) error {
	storage := memory.NewStorage()
	fs := memfs.New()

	fmt.Println("git clone next")
	repository, err := git.Clone(storage, fs, &git.CloneOptions{
		URL: "https://github.com/Kyrremann/beers",
		Auth: &githttp.BasicAuth{
			Username: "x-token",
			Password: os.Getenv("github_path"),
		},
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.json", user)
	file, err := fs.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.Write(untappd)
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	_, err = worktree.Add(fileName)
	if err != nil {
		return err
	}

	_, err = worktree.Commit(fmt.Sprintf("New data for %s", user), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Beers CI",
			Email: "beers-function@users.noreply.github.com",
		},
	})
	if err != nil {
		return err
	}

	err = repository.Push(&git.PushOptions{})
	if err != nil {
		return err
	}

	return nil
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       "Post please",
		}, nil
	}

	user := strings.ToLower(request.Headers["user"])
	userToken := request.Headers["user-token"]
	if userToken == "" || user == "" {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing user or user-token",
		}, nil
	}

	controlUserToken := os.Getenv(user)
	if controlUserToken != userToken {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Wrong user-token",
		}, nil
	}

	err := gitIt(user, []byte(request.Body))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func main() {
	lambda.Start(handler)
}
