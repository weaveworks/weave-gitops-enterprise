package git

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/pkg/github/hub"
	"github.com/weaveworks/wksctl/pkg/utilities/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	DefaultAuthor = "Weaveworks Kubernetes Platform"
	DefaultEmail  = "support@weave.works"
)

type GitRepo struct {
	worktreeDir string
	repo        *git.Repository
	auth        *gitssh.PublicKeys
}

func (gr *GitRepo) WorktreeDir() string {
	return gr.worktreeDir
}

func (gr *GitRepo) RemoteURL(name string) (string, error) {
	remote, err := gr.repo.Remote(name)
	if err != nil {
		return "", errors.Wrap(err, "get remote 'origin'")
	}
	remoteCfg := remote.Config()
	if len(remoteCfg.URLs) < 1 {
		return "", errors.New("remote has no URLs")
	}
	return remoteCfg.URLs[0], nil
}

func NewGithubRepoToTempDir(parentDir, repoName string) (*GitRepo, *ssh.KeyPair, error) {
	log.Infof("Creating a temp directory...")
	gitDir, err := ioutil.TempDir(parentDir, "git-")
	if err != nil {
		return nil, nil, errors.Wrap(err, "TempDir")
	}
	log.Infof("Temp directory %q created.", gitDir)

	log.Infof("Initializing an empty Git repository in %q...", gitDir)
	repo, err := git.PlainInit(gitDir, false)
	if err != nil {
		return nil, nil, errors.Wrap(err, "git init")
	}

	log.Infof("Creating the GitHub repository %q...", repoName)
	// XXX: hub.Create succeeds if the remote repo already exists.
	if _, err := hub.Create(gitDir, true, repoName); err != nil {
		return nil, nil, errors.Wrap(err, "create github repo")
	}

	bits := 4096
	log.Infof("Generating a %d-bit SSH key pair...", bits)
	keyPair, err := ssh.GenerateKeyPair(bits)
	if err != nil {
		return nil, nil, errors.Wrap(err, "ssh generate key pair")
	}

	log.Infof("Registering the SSH deploy key with the GitHub repository %q...", repoName)
	if err := hub.RegisterDeployKey(repoName, "wkp-gitops-key", string(keyPair.PublicRSA), false); err != nil {
		return nil, nil, errors.Wrap(err, "register deploy key")
	}

	auth, err := gitssh.NewPublicKeys("git", keyPair.PrivatePEM, "")
	if err != nil {
		return nil, nil, errors.Wrap(err, "private key read")
	}

	gr := &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}

	log.Infof("Pushing an initial (empty) commit to the Git repository %q...", repoName)
	if err := gr.CommitAll(DefaultAuthor, DefaultEmail, "initial commit"); err != nil {
		return nil, nil, errors.Wrap(err, "initial commit")
	}

	if err := gr.Push(); err != nil {
		return nil, nil, errors.Wrap(err, "git push")
	}

	return gr, keyPair, nil
}

func CloneToTempDir(parentDir, gitURL, branch string, privKey []byte) (*GitRepo, error) {
	log.Infof("Creating a temp directory...")
	gitDir, err := ioutil.TempDir(parentDir, "git-")
	if err != nil {
		return nil, errors.Wrap(err, "TempDir")
	}
	log.Infof("Temp directory %q created.", gitDir)

	log.Infof("Cloning the Git repository %q to %q...", gitURL, gitDir)
	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "private key read")
	}

	repo, err := git.PlainClone(gitDir, false, &git.CloneOptions{
		URL:           gitURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),

		SingleBranch: true,
		Tags:         git.NoTags,
		Depth:        10,
	})
	if err != nil {
		return nil, errors.Wrap(err, "git clone")
	}

	return &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}, nil
}

func (gr *GitRepo) CommitAll(name, email, msg string) error {
	wt, err := gr.repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "worktree")
	}

	if _, err := wt.Add("."); err != nil {
		return errors.Wrap(err, "git add")
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

func (gr *GitRepo) Push() error {
	log.Infof("Pushing local commits to the Git remote...")
	po := git.PushOptions{Auth: gr.auth}
	if err := po.Validate(); err != nil {
		return errors.Wrap(err, "push options validate")
	}
	return errors.Wrap(gr.repo.Push(&po), "git push")
}

func (gr *GitRepo) Close() error {
	log.Infof("Deleting the temp directory %q...", gr.worktreeDir)
	return errors.Wrapf(os.RemoveAll(gr.worktreeDir), "deleting directory %q", gr.worktreeDir)
}
