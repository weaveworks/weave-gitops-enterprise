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
	"github.com/weaveworks/wks/pkg/cmdutil"
	"github.com/weaveworks/wks/pkg/github/ggp"
	cryptossh "golang.org/x/crypto/ssh"
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

// CreateLocalRepo returns a *GitRepo based on a local git repository, and a private key. It's assumed
// that the local directory already has a default remote.
func CreateLocalRepo(gitDir string, privKey []byte) (*GitRepo, error) {
	log.Infof("Initializing local repository...")
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return nil, errors.Wrap(err, "set up local repo")
	}

	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "private key read")
	}

	return &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}, nil
}

// CreateGithubRepoWithDeployKey creates a remote GitHub repository with a deploy key and either
// updates a local repository to set its origin to the new repository or creates a new local repository
// with an origin of the new GitHub repository
func CreateGithubRepoWithDeployKey(
	org,
	repoName,
	gitDir string,
	privKey []byte) (*GitRepo, error) {

	log.Infof("Initializing local repository %q...", repoName)
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		return nil, errors.Wrap(err, "set up local repo")
	}

	log.Infof("Creating the remote git repository %q...", repoName)
	err = ggp.CreateEmpty(org, repoName, true)
	if err != nil {
		return nil, errors.Wrap(err, "create remote git repo")
	}

	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "private key read")
	}

	pubKey := cryptossh.MarshalAuthorizedKey(auth.Signer.PublicKey())
	log.Infof("Adding Deploy key to the remote git repository %q: %q", repoName, pubKey)
	fullRepoName := fmt.Sprintf("%s/%s", org, repoName)
	err = ggp.RegisterDeployKey(fullRepoName, "wkp-gitops-key", string(pubKey), false)
	if err != nil {
		return nil, errors.Wrap(err, "register deploy key")
	}

	gr := &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}

	log.Info("Updating git remote to point to new repo...")
	gr.DeleteRemote("origin") // okay if not present

	if err := gr.CreateRemote("origin", fmt.Sprintf("git@github.com:%s/%s.git", org, repoName)); err != nil {
		return nil, errors.Wrap(err, "create remote")
	}

	log.Info("Set up master branch")
	gr.repo.DeleteBranch("master") // okay if not present

	if err := gr.repo.CreateBranch(&config.Branch{Name: "master", Remote: "origin", Merge: "refs/heads/master"}); err != nil {
		return nil, errors.Wrap(err, "create branch")
	}

	return gr, nil
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
	err = updateYAMLFileFromStringPath(info, fieldPath, value)
	if err != nil {
		return err
	}
	err = PushAllChanges(gr)
	if err != nil {
		return err
	}
	return nil
}

// GetFileObjectInfo returns all the information necessary to use our YAML lookup and update facilities.
// Specifically, UpdateYAMLFile expects an instance of *ObjectInfo in order to update a YAML file.
func GetFileObjectInfo(path string) (*ObjectInfo, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var top yaml.Node
	err = yaml.Unmarshal(data, &top)
	if err != nil {
		return nil, err
	}

	if len(top.Content) == 0 {
		return nil, fmt.Errorf("YAML file is empty")
	}

	stats, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	perms := stats.Mode() | os.ModePerm
	return &ObjectInfo{FilePath: path, FilePerms: perms, FileNode: &top, ObjectNode: top.Content[0]}, nil
}

// UpdateYAMLFile updates the contents of a YAML file while preserving comments. It takes a field path consisting
// of field names and integers (for array indices), and inserts the substructure defined in the value argument.
// The value argument is also a sequence of items where each entry other than the last is either a field name,
// an integer, or '*' which causes the update to occur on each array element below the location referenced by the entry.
// The final entry in the path must be a valid YAML field value: string, integer, bool, etc.
func UpdateYAMLFile(info *ObjectInfo, fieldPath []string, value interface{}) error {
	err := UpdateNestedFields(info.ObjectNode, value, fieldPath...)
	if err != nil {
		return err
	}
	err = writeYamlNodeToFile(info.FilePath, info.FileNode, info.FilePerms)
	if err != nil {
		return err
	}
	return nil
}

func updateYAMLFileFromStringPath(info *ObjectInfo, fieldPathString string, value interface{}) error {
	return UpdateYAMLFile(info, strings.Split(fieldPathString, "."), value)
}

func FindNestedFields(data *yaml.Node, path ...string) []*yaml.Node {
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
		result = append(result, FindNestedFields(data.Content[intval], pathTail...)...)
	} else if elem == "*" {
		for _, item := range data.Content {
			result = append(result, FindNestedFields(item, pathTail...)...)
		}
	} else if data.Kind == yaml.DocumentNode {
		// Top level document just hods a single map node
		return FindNestedFields(data.Content[0], path...)
	} else if data.Kind == yaml.MappingNode {
		isKey := true
		for idx, entry := range data.Content {
			if isKey && entry.Value == elem {
				result = append(result, FindNestedFields(data.Content[idx+1], pathTail...)...)
				break
			}
			isKey = !isKey
		}
	}
	return result
}

func FindNestedField(data *yaml.Node, path ...string) *yaml.Node {
	nodes := FindNestedFields(data, path...)
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
	fieldNodes := FindNestedFields(data, path...)
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
		ns := FindNestedField(object, "metadata", "namespace")
		if ns == nil {
			if namespace != "default" {
				continue
			}
		} else if (ns.Value == "" && namespace != "default") || (ns.Value != "" && ns.Value != namespace) {
			continue
		}
		n := FindNestedField(object, "metadata", "name")
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
func GetMachinesK8sVersions(repoPath, machinesConfigPath string) ([]string, error) {
	// Parse the machines.yaml file
	machinesFilePath := path.Join(repoPath, machinesConfigPath)
	fileNode, _, err := readYamlNodeFromFile(machinesFilePath)
	if err != nil {
		return nil, err
	}

	// Look at the machine descriptions
	machineNodes := FindNestedField(&fileNode, "0", "items")
	if machineNodes == nil {
		return nil, fmt.Errorf("Machine items not found in %s", machinesConfigPath)
	}
	// Iterate through all the machines and collect their kubelet versions in a map
	versionMap := map[string]bool{}
	for _, machineNode := range machineNodes.Content {
		version := FindNestedField(machineNode, "spec", "versions", "kubelet")
		if version == nil || version.Value == "" {
			return nil, fmt.Errorf("Kubelet version missing for a node in %s", machinesConfigPath)
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

// GetEKSClusterVersion reads the cluster version from the wk-cluster.yaml file
func GetEKSClusterVersion(repoPath, wkClusterConfigPath string) ([]string, error) {
	// Parse the wk-cluster.yaml file
	wkClusterConfigPath = path.Join(repoPath, wkClusterConfigPath)
	fileNode, _, err := readYamlNodeFromFile(wkClusterConfigPath)
	if err != nil {
		return nil, err
	}

	// Find the spec node in the YAML file
	specNode := FindNestedField(&fileNode, "0", "spec")
	if specNode == nil {
		return nil, fmt.Errorf("Spec item not found in %s", wkClusterConfigPath)
	}

	// Read the version field in the spec node
	version := FindNestedField(specNode, "version")
	if version == nil || version.Value == "" {
		return nil, fmt.Errorf("Cluster version missing in %s", wkClusterConfigPath)
	}

	// Return value is an array to match the other tracks,
	// where the version of the kubelet of each node is returned.
	versions := []string{}
	versions = append(versions, version.Value)
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
	machineNodes := FindNestedField(&fileNode, "0", "items")
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
	return writeYamlNodeToFile(path, FindNestedField(&fileNode, "0"), filePerms)
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
	cmd := exec.Command("git", "ls-tree", "--name-only", "-r", commit)
	cmd.Dir = repoDir
	out, err := cmdutil.Output(cmd)
	if err != nil {
		return empty, wrap(err)
	}
	return FileLinesToSet(out), nil
}

func DiffTree(repoDir, oldCommit, newCommit string) (Set, error) {
	log.Debug("Comparing git commits to find new and deleted files...")
	wrap := MakeErrorWrapper("gitDiffTree")
	empty := Set{}
	out, err := cmdutil.Output(exec.Command("git", "-C", repoDir, "diff-tree", "--no-commit-id", "--name-only", "-r", oldCommit, newCommit))
	if err != nil {
		return empty, wrap(err)
	}
	return FileLinesToSet(out), nil
}

func ReadFile(repoDir, filePath, commit string) ([]byte, error) {
	log.Debugf("Reading file %q from git commit %q...", filePath, commit)
	cmd := exec.Command("git", "show", commit+":"+filePath)
	cmd.Dir = repoDir
	data, err := cmdutil.Output(cmd)
	if err != nil {
		if !CheckExist(filePath, commit) {
			return []byte(""), nil
		}
		return nil, MakeErrorWrapper("gitReadFile")(err)
	}
	return data, nil
}
