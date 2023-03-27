package git

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"k8s.io/client-go/util/retry"
)

type JenkinsSCM struct{}

func (p *JenkinsSCM) ParseURL(repoURL *url.URL) (string, string, string, error) {
	pathParts := strings.Split(strings.Trim(repoURL.Path, "/"), "/")
	// This is a naiv implementation here. Right now use only Azure DevOps with
	// this provider. When we start to move other providers to use this scm
	// library instead of go-git-provider, this function has to be revised.
	if len(pathParts) != 4 {
		return "", "", "", fmt.Errorf("unbale to parse url %+v", repoURL)
	}

	org := pathParts[0]
	project := pathParts[1]
	name := pathParts[3]

	return org, project, name, nil
}

func (p *JenkinsSCM) GetRepository(ctx context.Context, log logr.Logger, client *scm.Client, repoURL *url.URL) (*scm.Repository, error) {
	org, project, name, err := p.ParseURL(repoURL)
	if err != nil {
		return nil, err
	}

	var repo *scm.Repository
	err = retry.OnError(DefaultBackoff,
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

func (p *JenkinsSCM) GetCurrentCommitOfBranch(ctx context.Context, client *scm.Client, repo *scm.Repository, branch, path string) (string, error) {
	commits, _, err := client.Git.ListCommits(ctx, repo.FullName, scm.CommitListOptions{Ref: branch, Path: path})
	if err != nil {
		return "", fmt.Errorf("failed to list commits: %w", err)
	}

	return commits[0].Sha, nil
}

func (p *JenkinsSCM) CreateFile(ctx context.Context, client *scm.Client, repo *scm.Repository, path string, params *scm.ContentParams) error {
	_, err := client.Contents.Create(ctx, repo.FullName, path, params)
	if err != nil {
		return fmt.Errorf("unable to write files to branch %q: %w", params.Branch, err)
	}

	return nil
}

func (p *JenkinsSCM) UpdateFile(ctx context.Context, client *scm.Client, repo *scm.Repository, path string, params *scm.ContentParams) error {
	_, err := client.Contents.Update(ctx, repo.FullName, path, params)
	if err != nil {
		return fmt.Errorf("unable to write files to branch %q: %w", params.Branch, err)
	}

	return nil
}

func (p *JenkinsSCM) Endpoint(repoURL, path string, params url.Values) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("unbale to parse url %q: %w", repoURL, err)
	}

	org, project, name, err := p.ParseURL(u)
	if err != nil {
		return "", err
	}

	params.Add("api-version", "6.0")

	return fmt.Sprintf(
		"%s/%s/_apis/git/repositories/%s/%s?%s",
		org,
		project,
		name,
		path,
		params.Encode(),
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
func (p *JenkinsSCM) CommitFilesRequest(sha, repoURL, head, message string, files []CommitFile) *scm.Request {
	endpoint, _ := p.Endpoint(repoURL, "pushes", url.Values{})

	ref := jscmRefUpdate{
		Name:        fmt.Sprintf("refs/heads/%s", head),
		OldObjectID: sha,
	}

	com := jscmCommitRef{
		Comment: message,
		Changes: []jscmChange{},
	}

	for _, file := range files {
		cha := jscmChange{}
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
	_ = json.NewEncoder(buf).Encode(&jscmContentCreateUpdate{
		RefUpdates: []jscmRefUpdate{ref},
		Commits:    []jscmCommitRef{com},
	})

	return &scm.Request{
		Method: http.MethodPost,
		Path:   endpoint,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: buf,
	}
}

// Why do we have this function?
//
// The "Contents.List()" call in the jenkins-x/go-scm library uses "path" and
// "recursionLevel=Full", but the Azure API does not like that and complains to
// use "scopePath" instead.
//
// References:
// - https://github.com/jenkins-x/go-scm/blob/main/scm/driver/azure/content.go#LL137C2-L137C2
// - https://learn.microsoft.com/en-us/rest/api/azure/devops/git/items/list?view=azure-devops-rest-6.0&tabs=HTTP
func (p *JenkinsSCM) ListContents(repoURL, path, ref string) (*scm.Request, error) {
	params := url.Values{}
	params.Add("scopePath", path)
	params.Add("recursionLevel", "full")
	params.Add("format", "json")

	endpoint, err := p.Endpoint(repoURL, "items", params)
	if err != nil {
		return nil, err
	}

	p.addRefToParams(&params, ref)

	return &scm.Request{
		Method: http.MethodGet,
		Path:   endpoint,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}, nil
}

// Based on:
// https://github.com/jenkins-x/go-scm/blob/main/scm/driver/azure/content.go#L204
func (p *JenkinsSCM) addRefToParams(params *url.Values, ref string) {
	if ref == "" {
		return
	}

	if len(ref) == 40 {
		params.Add("versionDescriptor.versionType", "commit")
	} else {
		params.Add("versionDescriptor.versionType", "branch")
	}

	params.Add("versionDescriptor.version", ref)
}
