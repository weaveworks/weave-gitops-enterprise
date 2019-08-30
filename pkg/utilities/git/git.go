package git

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func CommitAll(repo *git.Repository, name, email, msg string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "worktree")
	}

	if _, err := wt.Add("."); err != nil {
		return errors.Wrap(err, "AddGlob")
	}

	co := git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	}
	if _, err := wt.Commit(msg, &co); err != nil {
		return errors.Wrap(err, "commit")
	}

	return nil
}
