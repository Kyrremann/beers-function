package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

func gitIt(untappd []byte) error {
	storage := memory.NewStorage()
	fs := memfs.New()

	fmt.Println("git clone next")
	repository, err := git.Clone(storage, fs, &git.CloneOptions{
		URL: "https://github.com/Kyrremann/beers",
		//Auth: &http.BasicAuth{
		//	Username: "x-token", // anything except an empty string
		//	Password: "",
		//},
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	fmt.Println("git clone done")

	file, err := fs.Create("dat1.txt")
	if err != nil {
		return err
	}

	_, err = file.Write(untappd)
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	_, err = worktree.Add("dat1.txt")
	if err != nil {
		return err
	}

	_, err = worktree.Commit("Test", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Kyrremann",
			Email: "Kyrremann@gmail.com",
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

	user := request.Headers["user-token"]
	fmt.Println(user)
	if user == "" {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing user-token",
		}, nil
	}

	err := gitIt([]byte(request.Body))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func main() {
	lambda.Start(handler)
}
