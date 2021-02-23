package payload

import (
	"time"
)

type UserView struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	When  time.Time `json:"when"`
}

type CommitView struct {
	Hash         string   `json:"hash"`
	PGPSignature string   `json:"pgpSignature"`
	Message      string   `json:"message"`
	TreeHash     string   `json:"treeHash"`
	ParentHashes []string `json:"parentHashes"`
	Author       UserView `json:"author"`
	Committer    UserView `json:"committer"`
}

type BranchView struct {
	Name string     `json:"name"`
	Head CommitView `json:"head"`
}

type BranchesView struct {
	Branches []BranchView `json:"data"`
}
