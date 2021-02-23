package payload

import "time"

type GitCommitInfo struct {
	Token  string     `json:"token"`
	Commit CommitView `json:"commit"`
}

type CommitView struct {
	Sha       string   `json:"sha"`
	Author    UserView `json:"author"`
	Committer UserView `json:"committer"`
	Message   string   `json:"message"`
}

type UserView struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}
