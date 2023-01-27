package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"
)

func getRepo() error {
	client, err := gitlab.New("https://gitlab.git.dev.weave.works")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	client.Client = &http.Client{
		Transport: &transport.PrivateToken{
			Token: os.Getenv("GITLAB_TOKEN"),
		},
	}

	repo, res, err := client.Repositories.Find(context.TODO(), "wge/foot/infra/wge-sub-dev")
	if err != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}

		return fmt.Errorf("failed to get repo: %w\n%+v\n%s", err, res, body)
	}

	fmt.Println(repo.ID)

	return nil
}

func main() {
	err := getRepo()
	if err != nil {
		panic(err)
	}
}
