package git

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gerrors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/pkg/github/hub"
	"github.com/weaveworks/wksctl/pkg/utilities/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/yaml.v3"
)

const (
	DefaultAuthor = "Weaveworks Kubernetes Platform"
	DefaultEmail  = "support@weave.works"

	ComponentsFileName  = "components.yaml"
	GeneratorScriptName = "config.js"
)

type Set map[string]bool

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
		return "", gerrors.Wrap(err, "get remote 'origin'")
	}
	remoteCfg := remote.Config()
	if len(remoteCfg.URLs) < 1 {
		return "", gerrors.New("remote has no URLs")
	}
	return remoteCfg.URLs[0], nil
}

func NewGithubRepoToTempDir(parentDir, repoName string) (*GitRepo, *ssh.KeyPair, error) {
	log.Infof("Creating a temp directory...")
	gitDir, err := ioutil.TempDir(parentDir, "git-")
	if err != nil {
		return nil, nil, gerrors.Wrap(err, "TempDir")
	}
	log.Infof("Temp directory %q created.", gitDir)

	log.Infof("Initializing an empty Git repository in %q...", gitDir)
	repo, err := git.PlainInit(gitDir, false)
	if err != nil {
		return nil, nil, gerrors.Wrap(err, "git init")
	}

	log.Infof("Creating the GitHub repository %q...", repoName)
	// XXX: hub.Create succeeds if the remote repo already exists.
	if _, err := hub.Create(gitDir, true, repoName); err != nil {
		return nil, nil, gerrors.Wrap(err, "create github repo")
	}

	bits := 4096
	log.Infof("Generating a %d-bit SSH key pair...", bits)
	keyPair, err := ssh.GenerateKeyPair(bits)
	if err != nil {
		return nil, nil, gerrors.Wrap(err, "ssh generate key pair")
	}

	log.Infof("Registering the SSH deploy key with the GitHub repository %q...", repoName)
	if err := hub.RegisterDeployKey(repoName, "wkp-gitops-key", string(keyPair.PublicRSA), false); err != nil {
		return nil, nil, gerrors.Wrap(err, "register deploy key")
	}

	auth, err := gitssh.NewPublicKeys("git", keyPair.PrivatePEM, "")
	if err != nil {
		return nil, nil, gerrors.Wrap(err, "private key read")
	}

	gr := &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}

	log.Infof("Pushing an initial (empty) commit to the Git repository %q...", repoName)
	if err := gr.CommitAll(DefaultAuthor, DefaultEmail, "initial commit"); err != nil {
		return nil, nil, gerrors.Wrap(err, "initial commit")
	}

	if err := gr.Push(); err != nil {
		return nil, nil, gerrors.Wrap(err, "git push")
	}

	return gr, keyPair, nil
}

func CloneToTempDir(parentDir, gitURL, branch string, privKey []byte) (*GitRepo, error) {
	log.Infof("Creating a temp directory...")
	gitDir, err := ioutil.TempDir(parentDir, "git-")
	if err != nil {
		return nil, gerrors.Wrap(err, "TempDir")
	}
	log.Infof("Temp directory %q created.", gitDir)

	log.Infof("Cloning the Git repository %q to %q...", gitURL, gitDir)
	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, gerrors.Wrap(err, "private key read")
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
		return nil, gerrors.Wrap(err, "git clone")
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
		return gerrors.Wrap(err, "worktree")
	}

	if _, err := wt.Add("."); err != nil {
		return gerrors.Wrap(err, "git add")
	}

	co := git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	}
	if _, err := wt.Commit(msg, &co); err != nil {
		return gerrors.Wrap(err, "commit")
	}

	return nil
}

func (gr *GitRepo) Push() error {
	log.Infof("Pushing local commits to the Git remote...")
	po := git.PushOptions{Auth: gr.auth}
	if err := po.Validate(); err != nil {
		return gerrors.Wrap(err, "push options validate")
	}
	return gerrors.Wrap(gr.repo.Push(&po), "git push")
}

func (gr *GitRepo) Close() error {
	log.Infof("Deleting the temp directory %q...", gr.worktreeDir)
	return gerrors.Wrapf(os.RemoveAll(gr.worktreeDir), "deleting directory %q", gr.worktreeDir)
}

// Support for updating manifest files in git

type objectInfo struct {
	filePath   string
	filePerms  os.FileMode
	fileNode   *yaml.Node
	objectNode *yaml.Node
}

func findNestedField(data *yaml.Node, path ...string) *yaml.Node {
	current := data
	for _, elem := range path {
		intval, err := strconv.ParseInt(elem, 10, 64)
		if err == nil {
			current = current.Content[intval]
			continue
		}
		if current.Kind != 0x4 {
			// path went past end of structure
			return nil
		}
		isKey := true
		for idx, entry := range current.Content {
			if isKey && entry.Value == elem {
				current = current.Content[idx+1]
				break
			}
			isKey = !isKey
		}
	}
	return current
}

func updateNestedField(data *yaml.Node, value interface{}, path ...string) error {
	if fieldNode := findNestedField(data, path...); fieldNode != nil {
		fieldNode.Value = fmt.Sprintf("%v", value)
		return nil
	}
	return fmt.Errorf("Could not locate field: %s in object", strings.Join(path, "."))
}

func findObjectNode(top *yaml.Node, namespace, name string) *yaml.Node {
	objects := top.Content
	for _, object := range objects {
		ns := findNestedField(object, "metadata", "namespace")
		if (ns.Value == "" && namespace != "default") || (ns.Value != "" && ns.Value != namespace) {
			continue
		}
		n := findNestedField(object, "metadata", "name")
		if n == nil || n.Value == "" || n.Value != name {
			continue
		}
		return object
	}
	return nil
}

func findObject(dir, kind, namespace, name string) (*objectInfo, error) {
	var info *objectInfo
	filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !IsYAMLFile(path) {
			return nil
		}
		stats, err := os.Stat(path)
		if err != nil {
			return nil
		}
		perms := stats.Mode() | os.ModePerm
		filebytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		var top yaml.Node
		err = yaml.Unmarshal(filebytes, &top)
		if err != nil {
			return nil
		}
		localNode := findObjectNode(&top, namespace, name)
		if localNode == nil {
			return nil
		}
		info = &objectInfo{filePath: path, filePerms: perms, fileNode: &top, objectNode: localNode}
		// After finding the object, return an error to stop the file walk
		return gerrors.New("STOP")
	})
	if info != nil {
		return info, nil
	}
	return nil, fmt.Errorf("Could not find %s: %s/%s", kind, namespace, name)
}

func performManifestUpdate(repo *GitRepo, info *objectInfo, fieldPath string, value interface{}) error {
	err := updateNestedField(info.objectNode, value, strings.Split(fieldPath, ".")...)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	defer encoder.Close()
	encoder.SetIndent(2)
	err = encoder.Encode(info.fileNode)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(info.filePath, out.Bytes(), info.filePerms)
	if err != nil {
		return nil
	}
	err = repo.CommitAll(DefaultAuthor, DefaultEmail, "updated by WKP UI")
	if err != nil {
		return nil
	}
	err = repo.Push()
	if err != nil {
		return nil
	}
	log.Infof("Updated repo...")
	return nil
}

// Update a manifest in git; used to write back changes for gitops
func UpdateManifest(gitURL, gitBranch string, key []byte, kind, namespace, name, fieldPath string, value interface{}) error {
	repo, err := CloneToTempDir("", gitURL, gitBranch, key)
	if err != nil {
		return err
	}
	defer repo.Close()
	cluster, err := findObject(repo.WorktreeDir(), kind, namespace, name)
	if err != nil {
		return err
	}
	return performManifestUpdate(repo, cluster, fieldPath, value)
}

// Operations used in git actions for policy checking
func MakeErrorWrapper(context string) func(error) error {
	return func(err error) error {
		return gerrors.Wrap(err, context)
	}
}

func GetFileType(filename string) string {
	return strings.TrimLeft(filepath.Ext(filename), ".")
}

func IsYAMLFile(filename string) bool {
	ftype := GetFileType(filename)
	return ftype == "yaml" || ftype == "yml"
}

func CheckExist(filePath, commit string) bool {
	_, noExistErr := exec.Command("git", "cat-file", "-e", commit+":"+filePath).Output()
	return noExistErr == nil
}

// FileLinesToSet splits command output containing file names at newlines and creates a set of the
// file names
func FileLinesToSet(lines []byte) Set {
	setval := Set{}
	for _, str := range strings.Split(string(lines), "\n") {
		if strings.TrimSpace(str) != "" {
			setval[str] = true
		}
	}
	return setval
}

func ExecError(err error) error {
	return errors.New(string(err.(*exec.ExitError).Stderr))
}

func WithoutGenerationFiles(files Set) Set {
	result := Set{}
	for file := range files {
		if file != GeneratorScriptName && file != ComponentsFileName {
			result[file] = true
		}
	}
	return result
}

func GeneratorCmd(srcDir string) []string {
	return []string{"jk", "generate", "-o", srcDir, "-f", filepath.Join(srcDir, ComponentsFileName),
		filepath.Join(srcDir, GeneratorScriptName)}
}

func CopyRepoAtCommit(targetDir, repoDir, commit string) error {
	if err := exec.Command("git", "clone", repoDir, targetDir).Run(); err != nil {
		return err
	}
	if err := exec.Command("git", "-C", targetDir, "checkout", commit).Run(); err != nil {
		return err
	}
	return nil
}

func ListTree(repoDir, commit string) (Set, error) {
	wrap := MakeErrorWrapper("gitListTree")
	empty := Set{}
	existingDir, err := os.Getwd()
	if err != nil {
		return empty, wrap(err)
	}
	defer func() { os.Chdir(existingDir) }()
	os.Chdir(repoDir)
	out, err := exec.Command("git", "ls-tree", "--name-only", "-r", commit).Output()
	if err != nil {
		return empty, wrap(ExecError(err))
	}
	return FileLinesToSet(out), nil
}

func DiffTree(repoDir, oldCommit, newCommit string) (Set, error) {
	log.Debug("Comparing git commits to find new and deleted files...")
	wrap := MakeErrorWrapper("gitDiffTree")
	empty := Set{}
	out, err := exec.Command("git", "-C", repoDir, "diff-tree", "--no-commit-id", "--name-only", "-r", oldCommit, newCommit).Output()
	if err != nil {
		return empty, wrap(ExecError(err))
	}
	return FileLinesToSet(out), nil
}

func ReadFile(repoDir, filePath, commit string) ([]byte, error) {
	log.Debugf("Reading file %q from git commit %q...", filePath, commit)
	wrap := MakeErrorWrapper("gitReadFile")
	existingDir, err := os.Getwd()
	if err != nil {
		return nil, wrap(err)
	}
	defer func() { os.Chdir(existingDir) }()
	os.Chdir(repoDir)
	data, err := exec.Command("git", "show", commit+":"+filePath).Output()
	if err != nil {
		if !CheckExist(filePath, commit) {
			return []byte(""), nil
		}
		return nil, MakeErrorWrapper("gitReadFile")(ExecError(err))
	}
	return data, nil
}
