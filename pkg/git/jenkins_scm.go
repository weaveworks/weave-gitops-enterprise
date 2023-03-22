package git

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"k8s.io/client-go/util/retry"
)

type jenkinsSCM struct{}

func (p *jenkinsSCM) GetRepository(ctx context.Context, log logr.Logger, client *scm.Client, repoURL *url.URL) (*scm.Repository, error) {
	pathParts := strings.Split(strings.Trim(repoURL.Path, "/"), "/")
	if len(pathParts) != 4 {
		return nil, fmt.Errorf("unbale to parse url %+v", repoURL)
	}

	org := pathParts[0]
	project := pathParts[1]
	name := pathParts[3]

	var repo *scm.Repository
	err := retry.OnError(DefaultBackoff,
		func(err error) bool { return errors.Is(err, scm.ErrNotFound) },
		func() error {
			var err error
			repo, _, err = client.Repositories.Find(ctx, fmt.Sprintf("%s/%s/%s", org, project, name))
			if err != nil {
				log.Info("Retrying getting the repository")
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository %q: %w", repoURL, err)
	}

	return repo, nil
}

func (p *jenkinsSCM) GetCurrentCommitOfBranch(ctx context.Context, client *scm.Client, repo *scm.Repository, branch, path string) (string, error) {
	commits, _, err := client.Git.ListCommits(ctx, repo.FullName, scm.CommitListOptions{Ref: branch, Path: path})
	if err != nil {
		return "", fmt.Errorf("failed to list commits: %w", err)
	}

	return commits[0].Sha, nil
}

func (p *jenkinsSCM) CreateFile(ctx context.Context, client *scm.Client, repo *scm.Repository, path string, params *scm.ContentParams) error {
	_, err := client.Contents.Create(ctx, repo.FullName, path, params)
	if err != nil {
		return fmt.Errorf("unable to write files to branch %q: %w", params.Branch, err)
	}

	return nil
}

func (p *jenkinsSCM) UpdateFile(ctx context.Context, client *scm.Client, repo *scm.Repository, path string, params *scm.ContentParams) error {
	_, err := client.Contents.Update(ctx, repo.FullName, path, params)
	if err != nil {
		return fmt.Errorf("unable to write files to branch %q: %w", params.Branch, err)
	}

	return nil
}

func (p *jenkinsSCM) Endpoint(repoURL, path string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("unbale to parse url %q: %w", repoURL, err)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) != 4 {
		return "", fmt.Errorf("unbale to parse url %+v", u)
	}

	org := pathParts[0]
	project := pathParts[1]
	name := pathParts[3]

	return fmt.Sprintf(
		"%s/%s/_apis/git/repositories/%s/%s?api-version=6.0",
		org,
		project,
		name,
		path,
	), nil
}

// Here comes a tiny hack.
//
// jenkins-x/go-scm commits changes file by file, which is a pain. So here is a
// function that does the same as the library does but with multiple files per
// commit.
//
// See:
// https://github.com/jenkins-x/go-scm/blob/main/scm/driver/azure/content.go#L91

func (p *jenkinsSCM) CommitFilesRequest(sha, url, head, message string, files []CommitFile) *scm.Request {
	endpoint, _ := p.Endpoint(url, "pushes")

	ref := refUpdate{
		Name:        fmt.Sprintf("refs/heads/%s", head),
		OldObjectID: sha,
	}

	com := commitRef{
		Comment: message,
		Changes: []change{},
	}

	for _, file := range files {
		cha := change{}
		if file.Content == nil {
			cha.ChangeType = "edit"
		} else {
			cha.ChangeType = "add"
			cha.NewContent.Content = base64.StdEncoding.EncodeToString([]byte(*file.Content))
			cha.NewContent.ContentType = "base64encoded"
		}

		cha.Item.Path = file.Path

		com.Changes = append(com.Changes, cha)
	}

	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(&contentCreateUpdate{
		RefUpdates: []refUpdate{ref},
		Commits:    []commitRef{com},
	})

	return &scm.Request{
		Method: "POST",
		Path:   endpoint,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: buf,
	}
}

type refUpdate struct {
	Name        string `json:"name"`
	OldObjectID string `json:"oldObjectId,omitempty"`
}

type change struct {
	ChangeType string `json:"changeType"`
	Item       struct {
		Path string `json:"path"`
	} `json:"item"`
	NewContent struct {
		Content     string `json:"content,omitempty"`
		ContentType string `json:"contentType,omitempty"`
	} `json:"newContent,omitempty"`
}

type commitRef struct {
	Comment  string   `json:"comment,omitempty"`
	Changes  []change `json:"changes,omitempty"`
	CommitID string   `json:"commitId"`
	URL      string   `json:"url"`
}

type contentCreateUpdate struct {
	RefUpdates []refUpdate `json:"refUpdates"`
	Commits    []commitRef `json:"commits"`
}
