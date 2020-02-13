package git

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/pkg/github/hub"
	"github.com/weaveworks/wksctl/pkg/utilities/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	yaml "gopkg.in/yaml.v3"
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
		return "", errors.Wrap(err, "get remote 'origin'")
	}
	remoteCfg := remote.Config()
	if len(remoteCfg.URLs) < 1 {
		return "", errors.New("remote has no URLs")
	}
	return remoteCfg.URLs[0], nil
}

func (gr *GitRepo) CreateRemote(name, url string) error {
	_, err := gr.repo.CreateRemote(&config.RemoteConfig{Name: name, URLs: []string{url}})
	return err
}

func (gr *GitRepo) DeleteRemote(name string) error {
	return gr.repo.DeleteRemote(name)
}

func TarballToGithubRepo(user, repoName, tarPath, repoDir string, privKey []byte) (*GitRepo, error) {
	log.Infof("Creating a temp directory...")
	tmpDir, err := ioutil.TempDir("", "git-")
	if err != nil {
		return nil, errors.Wrap(err, "TempDir")
	}
	log.Infof("Temp directory %q created.", tmpDir)

	log.Infof("Unrolling tar ball into local git repo")
	cmdstr := fmt.Sprintf("cd %q && tar xz -f %q", tmpDir, tarPath)
	cmd := exec.Command("sh", "-c", cmdstr)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	gitDir := filepath.Join(tmpDir, repoDir)
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return nil, errors.Wrap(err, "git open")
	}

	log.Infof("Creating the GitHub repository %q...", repoName)
	if _, err := hub.CreateEmpty(user, repoName, true); err != nil {
		return nil, errors.Wrap(err, "create github repo")
	}

	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "private key read")
	}

	gr := &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}

	log.Info("Updating git remote to point to new repo...")
	if err := gr.CreateRemote("origin", fmt.Sprintf("git@github.com:%s/%s.git", user, repoName)); err != nil {
		return nil, errors.Wrap(err, "create remote")
	}

	log.Info("Set up master branch")
	if err := gr.repo.CreateBranch(&config.Branch{Name: "master", Remote: "origin", Merge: "refs/heads/master"}); err != nil {
		return nil, errors.Wrap(err, "create branch")
	}

	log.Infof("Pushing initial commit to the Git repository %q...", repoName)
	if err := gr.Push(); err != nil {
		return nil, errors.Wrap(err, "git push")
	}

	return gr, nil
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
	})
	if err != nil {
		return nil, errors.Wrap(err, "git clone")
	}

	log.Infof("Cloned repo: %s", gitURL)

	return &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}, nil
}

func (gr *GitRepo) CommitAll(name, email, msg string) error {
	return gr.CommitPath(name, email, msg, []string{"."})
}

func (gr *GitRepo) CommitPath(name, email, msg string, paths []string) error {
	wt, err := gr.repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "worktree")
	}

	for _, path := range paths {
		if _, err := wt.Add(path); err != nil {
			return errors.Wrap(err, "git add")
		}
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

func (gr *GitRepo) PushWithOptions(po git.PushOptions) error {
	log.Infof("Pushing local commits to the Git remote...")
	if po.Auth == nil {
		po.Auth = gr.auth
	}
	if err := po.Validate(); err != nil {
		return errors.Wrap(err, "push options validate")
	}
	return errors.Wrap(gr.repo.Push(&po), "git push")
}

func (gr *GitRepo) Remove(path string) error {
	wtree, err := gr.repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "retrieve repo worktree")
	}
	_, err = wtree.Remove(path)
	return err
}

func (gr *GitRepo) Close() error {
	log.Infof("Deleting the temp directory %q...", gr.worktreeDir)
	return errors.Wrapf(os.RemoveAll(gr.worktreeDir), "deleting directory %q", gr.worktreeDir)
}

// PushAllChanges commits and pushes all the file changes to a git repo
func PushAllChanges(repo *GitRepo) error {
	err := repo.CommitAll(DefaultAuthor, DefaultEmail, "updated by WKP UI")
	if err != nil {
		return err
	}
	err = repo.Push()
	if err != nil {
		return err
	}
	log.Infof("Updated repo...")
	return nil
}

// YAML file <-> node helpers

func readYamlNodeFromFile(path string) (yaml.Node, os.FileMode, error) {
	var fileNode yaml.Node
	var filePerms os.FileMode
	if !IsYAMLFile(path) {
		return fileNode, filePerms, fmt.Errorf("Not a YAML file")
	}
	stats, err := os.Stat(path)
	if err != nil {
		return fileNode, filePerms, err
	}
	filePerms = stats.Mode() | os.ModePerm
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fileNode, filePerms, err
	}
	err = yaml.Unmarshal(fileBytes, &fileNode)
	if err != nil {
		return fileNode, filePerms, err
	}
	return fileNode, filePerms, nil
}

func writeYamlNodeToFile(path string, fileNode *yaml.Node, filePerms os.FileMode) error {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	defer encoder.Close()
	encoder.SetIndent(2)
	err := encoder.Encode(fileNode)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, out.Bytes(), filePerms)
	if err != nil {
		return err
	}
	return nil
}

// Support for updating manifest files in git

type ObjectInfo struct {
	FilePath   string
	FilePerms  os.FileMode
	FileNode   *yaml.Node
	ObjectNode *yaml.Node
}

// Update a manifest in a local git repository and push the results
func (gr *GitRepo) PerformManifestUpdate(kind, namespace, name, fieldPath string, value interface{}) error {
	info, err := findObject(gr.WorktreeDir(), kind, namespace, name)
	if err != nil {
		return err
	}
	err = gr.UpdateYAMLFile(info, value, fieldPath)
	if err != nil {
		return err
	}
	err = PushAllChanges(gr)
	if err != nil {
		return err
	}
	return nil
}

func (gr *GitRepo) UpdateYAMLFile(info *ObjectInfo, value interface{}, fieldPath string) error {
	err := UpdateNestedFields(info.ObjectNode, value, strings.Split(fieldPath, ".")...)
	if err != nil {
		return err
	}
	err = writeYamlNodeToFile(info.FilePath, info.FileNode, info.FilePerms)
	if err != nil {
		return err
	}
	return nil
}

func findNestedFields(data *yaml.Node, path ...string) []*yaml.Node {
	if len(path) == 0 {
		return []*yaml.Node{data}
	}

	if data == nil {
		return []*yaml.Node{}
	}

	result := []*yaml.Node{}
	elem := path[0]
	pathTail := path[1:]

	intval, err := strconv.ParseInt(elem, 10, 64)
	if err == nil {
		result = append(result, findNestedFields(data.Content[intval], pathTail...)...)
	} else if elem == "*" {
		for _, item := range data.Content {
			result = append(result, findNestedFields(item, pathTail...)...)
		}
	} else if data.Kind == yaml.DocumentNode {
		// Top level document just hods a single map node
		return findNestedFields(data.Content[0], path...)
	} else if data.Kind == yaml.MappingNode {
		isKey := true
		for idx, entry := range data.Content {
			if isKey && entry.Value == elem {
				result = append(result, findNestedFields(data.Content[idx+1], pathTail...)...)
				break
			}
			isKey = !isKey
		}
	}
	return result
}

func findNestedField(data *yaml.Node, path ...string) *yaml.Node {
	nodes := findNestedFields(data, path...)
	nodeCount := len(nodes)
	if nodeCount == 0 || nodeCount > 1 {
		return nil
	}
	return nodes[0]
}

func UpdateNestedFields(data *yaml.Node, val interface{}, path ...string) error {
	vals, ok := val.([]interface{})
	if !ok {
		vals = []interface{}{val}
	}
	if len(vals) == 0 {
		return fmt.Errorf("Missing value for field: %s", strings.Join(path, "."))
	}
	fieldNodes := findNestedFields(data, path...)
	if len(fieldNodes) == 0 {
		return fmt.Errorf("Could not locate fields: %s in object", strings.Join(path, "."))
	}
	if len(vals) == 1 {
		for _, fieldNode := range fieldNodes {
			fieldNode.Value = fmt.Sprintf("%v", val)
		}
		return nil
	}

	// We're inserting into an existing structure. If the following index is a number, it's a sequence node; otherwise scalar
	// We know "values" contains at least two entries at this point.
	subNode := createSubNode(vals[1:])
	value := vals[0]
	intVal, ok := value.(int)
	for _, fieldNode := range fieldNodes {
		if ok {
			if err := insertNodeIntoSequence(fieldNode, intVal, subNode); err != nil {
				return err
			}
		} else {
			insertNodeIntoMap(fieldNode, value.(string), subNode)
		}
	}
	return nil
}

func insertNodeIntoSequence(data *yaml.Node, idx int, subNode *yaml.Node) error {
	seq := data.Content
	if len(seq) < idx {
		return fmt.Errorf("Sequence not large enough to insert value at index: %d", idx)
	}
	data.Content = append(append(seq[0:idx], subNode), seq[idx:]...)
	return nil
}

func insertNodeIntoMap(data *yaml.Node, key string, subNode *yaml.Node) {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	data.Content = append(data.Content, keyNode, subNode)
}

func createSubNode(vals []interface{}) *yaml.Node {
	val := vals[0]
	tag := "!!str"
	if _, isInt := val.(int); isInt {
		tag = "!!int"
	}

	if len(vals) == 1 {
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", val), Tag: tag}
	}

	subNode := createSubNode(vals[1:])
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: val.(string)}
	return &yaml.Node{Kind: yaml.MappingNode, Value: "", Content: []*yaml.Node{keyNode, subNode}}
}

func findObjectNode(top *yaml.Node, kind, namespace, name string) *yaml.Node {
	objects := top.Content
	if kind == "List" {
		return objects[0]
	}
	for _, object := range objects {
		ns := findNestedField(object, "metadata", "namespace")
		if ns == nil {
			if namespace != "default" {
				continue
			}
		} else if (ns.Value == "" && namespace != "default") || (ns.Value != "" && ns.Value != namespace) {
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

func findObject(dir, kind, namespace, name string) (*ObjectInfo, error) {
	var info *ObjectInfo
	filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		fileNode, filePerms, err := readYamlNodeFromFile(path)
		if err != nil {
			return nil
		}
		localNode := findObjectNode(&fileNode, kind, namespace, name)
		if localNode == nil {
			return nil
		}
		info = &ObjectInfo{FilePath: path, FilePerms: filePerms, FileNode: &fileNode, ObjectNode: localNode}
		// After finding the object, return an error to stop the file walk
		return errors.New("STOP")
	})
	if info != nil {
		return info, nil
	}
	return nil, fmt.Errorf("Could not find %s: %s/%s", kind, namespace, name)
}

// Update a manifest in git; used to write back changes for gitops
func UpdateManifest(gitURL, gitBranch string, key []byte, kind, namespace, name, fieldPath string, value interface{}) error {
	repo, err := CloneToTempDir("", gitURL, gitBranch, key)
	if err != nil {
		return err
	}
	defer repo.Close()
	return repo.PerformManifestUpdate(kind, namespace, name, fieldPath, value)
}

// GetMachinesK8sVersions gets a list of unique K8s versions for all cluster machines
func GetMachinesK8sVersions(repoPath, fileSubPath string) ([]string, error) {
	// Parse the YAML file
	path := path.Join(repoPath, fileSubPath)
	fileNode, _, err := readYamlNodeFromFile(path)
	if err != nil {
		return nil, err
	}
	// Look at the machine descriptions
	machineNodes := findNestedField(&fileNode, "0", "items")
	if machineNodes == nil {
		return nil, fmt.Errorf("Machine items not found in %s", fileSubPath)
	}
	// Iterate through all the machines and collect their kubelet versions in a map
	versionMap := map[string]bool{}
	for _, machineNode := range machineNodes.Content {
		version := findNestedField(machineNode, "spec", "versions", "kubelet")
		if version == nil || version.Value == "" {
			return nil, fmt.Errorf("Kubelet version missing for a node in %s", fileSubPath)
		}
		versionMap[version.Value] = true
	}
	// Return a sorted list of unique versions
	versions := []string{}
	for version := range versionMap {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return versions, nil
}

// UpdateMachinesK8sVersions updates all machines in machines.yaml to the same K8s version
func UpdateMachinesK8sVersions(repoPath, fileSubPath string, version string) error {
	// Parse the YAML file
	path := path.Join(repoPath, fileSubPath)
	fileNode, filePerms, err := readYamlNodeFromFile(path)
	if err != nil {
		return err
	}
	// Look at the machine descriptions
	machineNodes := findNestedField(&fileNode, "0", "items")
	if machineNodes == nil {
		return fmt.Errorf("Machine items not found in %s", fileSubPath)
	}
	// Iterate through all the machines and update their kubelet versions
	for _, machineNode := range machineNodes.Content {
		err := UpdateNestedFields(machineNode, version, "spec", "versions", "kubelet")
		if err != nil {
			return err
		}
	}
	// Write the updated results back to the file with same permissions
	return writeYamlNodeToFile(path, findNestedField(&fileNode, "0"), filePerms)
}

// Operations used in git actions for policy checking
func MakeErrorWrapper(context string) func(error) error {
	return func(err error) error {
		return errors.Wrap(err, context)
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
